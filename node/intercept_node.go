package node

import (
	"context"
	"fmt"

	cfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/service"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/p2p/pex"
	"github.com/cometbft/cometbft/privval"
	"github.com/cometbft/cometbft/proxy"
	sm "github.com/cometbft/cometbft/state"
	"github.com/cometbft/cometbft/statesync"
	"github.com/cometbft/cometbft/types"
)

func InterceptNode(config *cfg.Config, logger log.Logger) (*Node, error) {
	nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
	if err != nil {
		return nil, fmt.Errorf("failed to load or gen node key %s: %w", config.NodeKeyFile(), err)
	}

	return NewInterceptNode(context.Background(), config,
		privval.LoadOrGenFilePV(config.PrivValidatorKeyFile(), config.PrivValidatorStateFile()),
		nodeKey,
		proxy.DefaultClientCreator(config.ProxyApp, config.ABCI, config.DBDir()),
		DefaultGenesisDocProviderFunc(config),
		cfg.DefaultDBProvider,
		DefaultMetricsProvider(config.Instrumentation),
		logger,
	)
}

func NewInterceptNode(ctx context.Context,
	config *cfg.Config,
	privValidator types.PrivValidator,
	nodeKey *p2p.NodeKey,
	clientCreator proxy.ClientCreator,
	genesisDocProvider GenesisDocProvider,
	dbProvider cfg.DBProvider,
	metricsProvider MetricsProvider,
	logger log.Logger,
	options ...Option,
) (*Node, error) {
	blockStore, stateDB, err := initDBs(config, dbProvider)
	if err != nil {
		return nil, err
	}

	stateStore := sm.NewStore(stateDB, sm.StoreOptions{
		DiscardABCIResponses: config.Storage.DiscardABCIResponses,
	})

	state, genDoc, err := LoadStateFromDBOrGenesisDocProvider(stateDB, genesisDocProvider, config.Storage.GenesisHash)
	if err != nil {
		return nil, err
	}

	// The key will be deleted if it existed.
	// Not checking whether the key is there in case the genesis file was larger than
	// the max size of a value (in rocksDB for example), which would cause the check
	// to fail and prevent the node from booting.
	logger.Info("WARNING: deleting genesis file from database if present, the database stores a hash of the original genesis file now")

	err = stateDB.Delete(genesisDocKey)
	if err != nil {
		logger.Error("Failed to delete genesis doc from DB ", err)
	}

	csMetrics, p2pMetrics, memplMetrics, smMetrics, abciMetrics, bsMetrics, ssMetrics := metricsProvider(genDoc.ChainID)

	// Create the proxyApp and establish connections to the ABCI app (consensus, mempool, query).
	proxyApp, err := createAndStartProxyAppConns(clientCreator, logger, abciMetrics)
	if err != nil {
		return nil, err
	}

	// EventBus and IndexerService must be started before the handshake because
	// we might need to index the txs of the replayed block as this might not have happened
	// when the node stopped last time (i.e. the node stopped after it saved the block
	// but before it indexed the txs)
	eventBus, err := createAndStartEventBus(logger)
	if err != nil {
		return nil, err
	}

	indexerService, txIndexer, blockIndexer, err := createAndStartIndexerService(config,
		genDoc.ChainID, dbProvider, eventBus, logger)
	if err != nil {
		return nil, err
	}

	// If an address is provided, listen on the socket for a connection from an
	// external signing process.
	if config.PrivValidatorListenAddr != "" {
		// FIXME: we should start services inside OnStart
		privValidator, err = createAndStartPrivValidatorSocketClient(config.PrivValidatorListenAddr, genDoc.ChainID, logger)
		if err != nil {
			return nil, fmt.Errorf("error with private validator socket client: %w", err)
		}
	}

	pubKey, err := privValidator.GetPubKey()
	if err != nil {
		return nil, fmt.Errorf("can't get pubkey: %w", err)
	}

	// Determine whether we should attempt state sync.
	stateSync := config.StateSync.Enable && !onlyValidatorIsUs(state, pubKey)
	if stateSync && state.LastBlockHeight > 0 {
		logger.Info("Found local state with non-zero height, skipping state sync")
		stateSync = false
	}

	// Create the handshaker, which calls RequestInfo, sets the AppVersion on the state,
	// and replays any blocks as necessary to sync CometBFT with the app.
	consensusLogger := logger.With("module", "consensus")
	if !stateSync {
		if err := doHandshake(ctx, stateStore, state, blockStore, genDoc, eventBus, proxyApp, consensusLogger); err != nil {
			return nil, err
		}

		// Reload the state. It will have the Version.Consensus.App set by the
		// Handshake, and may have other modifications as well (ie. depending on
		// what happened during block replay).
		state, err = stateStore.Load()
		if err != nil {
			return nil, sm.ErrCannotLoadState{Err: err}
		}
	}

	// Determine whether we should do block sync. This must happen after the handshake, since the
	// app may modify the validator set, specifying ourself as the only validator.
	blockSync := !onlyValidatorIsUs(state, pubKey)
	waitSync := stateSync || blockSync

	logNodeStartupInfo(state, pubKey, logger, consensusLogger)

	// Make MempoolReactor
	mempool, mempoolReactor := createMempoolAndMempoolReactor(config, proxyApp, state, waitSync, memplMetrics, logger)

	// Make Evidence Reactor
	evidenceReactor, evidencePool, err := createEvidenceReactor(config, dbProvider, stateStore, blockStore, logger)
	if err != nil {
		return nil, err
	}

	pruner, err := createPruner(
		config,
		txIndexer,
		blockIndexer,
		stateStore,
		blockStore,
		smMetrics,
		logger.With("module", "state"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create pruner: %w", err)
	}

	// make block executor for consensus and blocksync reactors to execute blocks
	blockExec := sm.NewBlockExecutor(
		stateStore,
		logger.With("module", "state"),
		proxyApp.Consensus(),
		mempool,
		evidencePool,
		blockStore,
		sm.BlockExecutorWithPruner(pruner),
		sm.BlockExecutorWithMetrics(smMetrics),
	)

	offlineStateSyncHeight := int64(0)
	if blockStore.Height() == 0 {
		offlineStateSyncHeight, err = blockExec.Store().GetOfflineStateSyncHeight()
		if err != nil && err.Error() != "value empty" {
			panic(fmt.Sprintf("failed to retrieve statesynced height from store %s; expected state store height to be %v", err, state.LastBlockHeight))
		}
	}
	// Make BlocksyncReactor. Don't start block sync if we're doing a state sync first.
	bcReactor, err := createBlocksyncReactor(config, state, blockExec, blockStore, blockSync && !stateSync, logger, bsMetrics, offlineStateSyncHeight)
	if err != nil {
		return nil, fmt.Errorf("could not create blocksync reactor: %w", err)
	}

	// Make ConsensusReactor
	consensusReactor, consensusState := createConsensusReactor(
		config, state, blockExec, blockStore, mempool, evidencePool,
		privValidator, csMetrics, waitSync, eventBus, consensusLogger, offlineStateSyncHeight,
	)

	err = stateStore.SetOfflineStateSyncHeight(0)
	if err != nil {
		panic(fmt.Sprintf("failed to reset the offline state sync height %s", err))
	}
	// Set up state sync reactor, and schedule a sync if requested.
	// FIXME The way we do phased startups (e.g. replay -> block sync -> consensus) is very messy,
	// we should clean this whole thing up. See:
	// https://github.com/tendermint/tendermint/issues/4644
	stateSyncReactor := statesync.NewReactor(
		*config.StateSync,
		proxyApp.Snapshot(),
		proxyApp.Query(),
		ssMetrics,
	)
	stateSyncReactor.SetLogger(logger.With("module", "statesync"))

	nodeInfo, err := makeNodeInfo(config, nodeKey, txIndexer, genDoc, state)
	if err != nil {
		return nil, err
	}

	// Setup Transport.
	transport, peerFilters := createTransport(config, nodeInfo, nodeKey, proxyApp)

	// Setup Switch.
	p2pLogger := logger.With("module", "p2p")
	sw := createSwitch(
		config, transport, p2pMetrics, peerFilters, mempoolReactor, bcReactor,
		stateSyncReactor, consensusReactor, evidenceReactor, nodeInfo, nodeKey, p2pLogger,
	)

	err = sw.AddPersistentPeers(splitAndTrimEmpty(config.P2P.PersistentPeers, ",", " "))
	if err != nil {
		return nil, fmt.Errorf("could not add peers from persistent_peers field: %w", err)
	}

	err = sw.AddUnconditionalPeerIDs(splitAndTrimEmpty(config.P2P.UnconditionalPeerIDs, ",", " "))
	if err != nil {
		return nil, fmt.Errorf("could not add peer ids from unconditional_peer_ids field: %w", err)
	}

	addrBook, err := createAddrBookAndSetOnSwitch(config, sw, p2pLogger, nodeKey)
	if err != nil {
		return nil, fmt.Errorf("could not create addrbook: %w", err)
	}

	// Optionally, start the pex reactor
	//
	// TODO:
	//
	// We need to set Seeds and PersistentPeers on the switch,
	// since it needs to be able to use these (and their DNS names)
	// even if the PEX is off. We can include the DNS name in the NetAddress,
	// but it would still be nice to have a clear list of the current "PersistentPeers"
	// somewhere that we can return with net_info.
	//
	// If PEX is on, it should handle dialing the seeds. Otherwise the switch does it.
	// Note we currently use the addrBook regardless at least for AddOurAddress
	var pexReactor *pex.Reactor
	if config.P2P.PexReactor {
		pexReactor = createPEXReactorAndAddToSwitch(addrBook, config, sw, logger)
	}

	// Add private IDs to addrbook to block those peers being added
	addrBook.AddPrivateIDs(splitAndTrimEmpty(config.P2P.PrivatePeerIDs, ",", " "))

	node := &Node{
		config:        config,
		genesisDoc:    genDoc,
		privValidator: privValidator,

		transport: transport,
		sw:        sw,
		addrBook:  addrBook,
		nodeInfo:  nodeInfo,
		nodeKey:   nodeKey,

		stateStore:       stateStore,
		blockStore:       blockStore,
		pruner:           pruner,
		bcReactor:        bcReactor,
		mempoolReactor:   mempoolReactor,
		mempool:          mempool,
		consensusState:   consensusState,
		consensusReactor: consensusReactor,
		stateSyncReactor: stateSyncReactor,
		stateSync:        stateSync,
		stateSyncGenesis: state, // Shouldn't be necessary, but need a way to pass the genesis state
		pexReactor:       pexReactor,
		evidencePool:     evidencePool,
		proxyApp:         proxyApp,
		txIndexer:        txIndexer,
		indexerService:   indexerService,
		blockIndexer:     blockIndexer,
		eventBus:         eventBus,
	}
	node.BaseService = *service.NewBaseService(logger, "Node", node)

	for _, option := range options {
		option(node)
	}

	return node, nil
}
