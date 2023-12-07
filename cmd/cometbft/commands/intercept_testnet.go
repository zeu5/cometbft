package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	cfg "github.com/zeu5/cometbft/config"
	cmtrand "github.com/zeu5/cometbft/libs/rand"
	"github.com/zeu5/cometbft/p2p"
	"github.com/zeu5/cometbft/privval"
	"github.com/zeu5/cometbft/types"
	cmttime "github.com/zeu5/cometbft/types/time"
)

var (
	interceptPort       int
	rpcPort             int
	interceptServerAddr string
	timeoutPropose      int
	timeoutPrevote      int
	timeoutPrecommit    int
	timeoutCommit       int
)

func init() {
	ITestnetFilesCmd.Flags().IntVar(&nValidators, "v", 4,
		"number of validators to initialize the testnet with")
	ITestnetFilesCmd.Flags().StringVar(&configFile, "config", "",
		"config file to use (note some options may be overwritten)")
	ITestnetFilesCmd.Flags().IntVar(&nNonValidators, "n", 0,
		"number of non-validators to initialize the testnet with")
	ITestnetFilesCmd.Flags().StringVar(&outputDir, "o", "./mytestnet",
		"directory to store initialization data for the testnet")
	ITestnetFilesCmd.Flags().StringVar(&nodeDirPrefix, "node-dir-prefix", "node",
		"prefix the directory name for each node with (node results in node0, node1, ...)")
	ITestnetFilesCmd.Flags().Int64Var(&initialHeight, "initial-height", 0,
		"initial height of the first block")

	ITestnetFilesCmd.Flags().BoolVar(&populatePersistentPeers, "populate-persistent-peers", true,
		"update config of each node with the list of persistent peers build using either"+
			" hostname-prefix or"+
			" starting-ip-address")
	ITestnetFilesCmd.Flags().StringVar(&hostnamePrefix, "hostname-prefix", "node",
		"hostname prefix (\"node\" results in persistent peers list ID0@node0:26656, ID1@node1:26656, ...)")
	ITestnetFilesCmd.Flags().StringVar(&hostnameSuffix, "hostname-suffix", "",
		"hostname suffix ("+
			"\".xyz.com\""+
			" results in persistent peers list ID0@node0.xyz.com:26656, ID1@node1.xyz.com:26656, ...)")
	ITestnetFilesCmd.Flags().StringVar(&startingIPAddress, "starting-ip-address", "",
		"starting IP address ("+
			"\"192.168.0.1\""+
			" results in persistent peers list ID0@192.168.0.1:26656, ID1@192.168.0.2:26656, ...)")
	ITestnetFilesCmd.Flags().StringArrayVar(&hostnames, "hostname", []string{},
		"manually override all hostnames of validators and non-validators (use --hostname multiple times for multiple hosts)")
	ITestnetFilesCmd.Flags().IntVar(&p2pPort, "p2p-port", 26656,
		"P2P Port")
	ITestnetFilesCmd.Flags().BoolVar(&randomMonikers, "random-monikers", false,
		"randomize the moniker for each generated node")
	ITestnetFilesCmd.Flags().IntVar(&interceptPort, "intercept-port", 2023,
		"Intercept port")
	ITestnetFilesCmd.Flags().IntVar(&rpcPort, "rpc-port", 26756,
		"Base RPC port")
	ITestnetFilesCmd.Flags().StringVar(&interceptServerAddr, "intercept-server-addr", "localhost:7074",
		"Intercept server address")
	ITestnetFilesCmd.Flags().IntVar(&timeoutPropose, "timeout-propose", 100,
		"Proposal Timeout")
	ITestnetFilesCmd.Flags().IntVar(&timeoutPrevote, "timeout-prevote", 50,
		"Prevote Timeout")
	ITestnetFilesCmd.Flags().IntVar(&timeoutPrecommit, "timeout-precommit", 50,
		"Precommit Timeout")
	ITestnetFilesCmd.Flags().IntVar(&timeoutCommit, "timeout-commit", 20,
		"Commit Timeout")
}

// TestnetFilesCmd allows initialisation of files for a CometBFT testnet.
var ITestnetFilesCmd = &cobra.Command{
	Use:   "itestnet",
	Short: "Initialize files for a CometBFT testnet with intercept transport",
	Long: `testnet will create "v" + "n" number of directories and populate each with
necessary files (private validator, genesis, config, etc.).

Note, strict routability for addresses is turned off in the config file.

Optionally, it will fill in persistent_peers list in config file using either hostnames or IPs.

Example:

	cometbft testnet --v 4 --o ./output --populate-persistent-peers --starting-ip-address 192.168.10.2
	`,
	RunE: interceptTestnetFiles,
}

func interceptTestnetFiles(*cobra.Command, []string) error {
	if len(hostnames) > 0 && len(hostnames) != (nValidators+nNonValidators) {
		return fmt.Errorf(
			"testnet needs precisely %d hostnames (number of validators plus non-validators) if --hostname parameter is used",
			nValidators+nNonValidators,
		)
	}

	config := cfg.DefaultConfig()

	config.Consensus.TimeoutProposeDelta = 10 * time.Millisecond
	config.Consensus.TimeoutPrevoteDelta = 10 * time.Millisecond
	config.Consensus.TimeoutPrecommitDelta = 10 * time.Millisecond
	config.Consensus.TimeoutPropose = time.Duration(timeoutPropose) * time.Millisecond
	config.Consensus.TimeoutPrevote = time.Duration(timeoutPrevote) * time.Millisecond
	config.Consensus.TimeoutPrecommit = time.Duration(timeoutPrecommit) * time.Millisecond
	config.Consensus.TimeoutCommit = time.Duration(timeoutCommit) * time.Millisecond
	config.Consensus.PeerGossipSleepDuration = 10 * time.Millisecond
	config.Consensus.PeerQueryMaj23SleepDuration = 100 * time.Millisecond

	// overwrite default config if set and valid
	if configFile != "" {
		viper.SetConfigFile(configFile)
		if err := viper.ReadInConfig(); err != nil {
			return err
		}
		if err := viper.Unmarshal(config); err != nil {
			return err
		}
		if err := config.ValidateBasic(); err != nil {
			return err
		}
	}

	genVals := make([]types.GenesisValidator, nValidators)

	for i := 0; i < nValidators; i++ {
		nodeDirName := fmt.Sprintf("%s%d", nodeDirPrefix, i+1)
		nodeDir := filepath.Join(outputDir, nodeDirName)
		config.SetRoot(nodeDir)

		err := os.MkdirAll(filepath.Join(nodeDir, "config"), nodeDirPerm)
		if err != nil {
			_ = os.RemoveAll(outputDir)
			return err
		}
		err = os.MkdirAll(filepath.Join(nodeDir, "data"), nodeDirPerm)
		if err != nil {
			_ = os.RemoveAll(outputDir)
			return err
		}

		if err := initFilesWithConfig(config); err != nil {
			return err
		}

		pvKeyFile := filepath.Join(nodeDir, config.BaseConfig.PrivValidatorKey)
		pvStateFile := filepath.Join(nodeDir, config.BaseConfig.PrivValidatorState)
		pv := privval.LoadFilePV(pvKeyFile, pvStateFile)

		pubKey, err := pv.GetPubKey()
		if err != nil {
			return fmt.Errorf("can't get pubkey: %w", err)
		}
		genVals[i] = types.GenesisValidator{
			Address: pubKey.Address(),
			PubKey:  pubKey,
			Power:   1,
			Name:    nodeDirName,
		}
	}

	for i := 0; i < nNonValidators; i++ {
		nodeDir := filepath.Join(outputDir, fmt.Sprintf("%s%d", nodeDirPrefix, i+nValidators+1))
		config.SetRoot(nodeDir)

		err := os.MkdirAll(filepath.Join(nodeDir, "config"), nodeDirPerm)
		if err != nil {
			_ = os.RemoveAll(outputDir)
			return err
		}

		err = os.MkdirAll(filepath.Join(nodeDir, "data"), nodeDirPerm)
		if err != nil {
			_ = os.RemoveAll(outputDir)
			return err
		}

		if err := initFilesWithConfig(config); err != nil {
			return err
		}
	}

	// Generate genesis doc from generated validators
	genDoc := &types.GenesisDoc{
		ChainID:         "chain-" + cmtrand.Str(6),
		ConsensusParams: types.DefaultConsensusParams(),
		GenesisTime:     cmttime.Now(),
		InitialHeight:   initialHeight,
		Validators:      genVals,
	}

	// Write genesis file.
	for i := 0; i < nValidators+nNonValidators; i++ {
		nodeDir := filepath.Join(outputDir, fmt.Sprintf("%s%d", nodeDirPrefix, i+1))
		if err := genDoc.SaveAs(filepath.Join(nodeDir, config.BaseConfig.Genesis)); err != nil {
			_ = os.RemoveAll(outputDir)
			return err
		}
	}

	// Gather persistent peer addresses.
	var (
		persistentPeers string
		err             error
	)
	if populatePersistentPeers {
		persistentPeers, err = persistentIPeersString(config)
		if err != nil {
			_ = os.RemoveAll(outputDir)
			return err
		}
	}

	// Overwrite default config.
	for i := 0; i < nValidators+nNonValidators; i++ {
		nodeDir := filepath.Join(outputDir, fmt.Sprintf("%s%d", nodeDirPrefix, i+1))
		config.SetRoot(nodeDir)

		config.ProxyApp = "kvstore"
		config.Consensus.CreateEmptyBlocks = false

		config.RPC.ListenAddress = fmt.Sprintf("tcp://localhost:%d", rpcPort+i)
		config.P2P.ListenAddress = fmt.Sprintf("tcp://localhost:%d", p2pPort+i)
		config.P2P.TestIntercept = true
		config.P2P.TestInterceptConfig.ListenAddr = fmt.Sprintf("localhost:%d", interceptPort+i)
		config.P2P.TestInterceptConfig.ServerAddr = interceptServerAddr
		config.P2P.TestInterceptConfig.ID = i + 1
		config.P2P.AddrBookStrict = false
		config.P2P.AllowDuplicateIP = true
		if populatePersistentPeers {
			config.P2P.PersistentPeers = persistentPeers
		}
		config.Moniker = moniker(i)

		cfg.WriteConfigFile(filepath.Join(nodeDir, "config", "config.toml"), config)
	}

	fmt.Printf("Successfully initialized %v node directories\n", nValidators+nNonValidators)
	return nil
}

func persistentIPeersString(config *cfg.Config) (string, error) {
	persistentPeers := make([]string, nValidators+nNonValidators)
	for i := 0; i < nValidators+nNonValidators; i++ {
		nodeDir := filepath.Join(outputDir, fmt.Sprintf("%s%d", nodeDirPrefix, i+1))
		config.SetRoot(nodeDir)
		nodeKey, err := p2p.LoadNodeKey(config.NodeKeyFile())
		if err != nil {
			return "", err
		}
		persistentPeers[i] = p2p.IDAddressString(nodeKey.ID(), fmt.Sprintf("localhost:%d", p2pPort+i))
	}
	return strings.Join(persistentPeers, ","), nil
}
