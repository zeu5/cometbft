package consensus

import (
	"context"

	abcicli "github.com/zeu5/cometbft/abci/client"
	abci "github.com/zeu5/cometbft/abci/types"
	"github.com/zeu5/cometbft/libs/clist"
	mempl "github.com/zeu5/cometbft/mempool"
	"github.com/zeu5/cometbft/proxy"
	"github.com/zeu5/cometbft/types"
)

//-----------------------------------------------------------------------------

type emptyMempool struct{}

var _ mempl.Mempool = emptyMempool{}

func (emptyMempool) Lock()            {}
func (emptyMempool) Unlock()          {}
func (emptyMempool) Size() int        { return 0 }
func (emptyMempool) SizeBytes() int64 { return 0 }
func (emptyMempool) CheckTx(types.Tx) (*abcicli.ReqRes, error) {
	return nil, nil
}

func (txmp emptyMempool) RemoveTxByKey(types.TxKey) error {
	return nil
}

func (emptyMempool) ReapMaxBytesMaxGas(int64, int64) types.Txs { return types.Txs{} }
func (emptyMempool) ReapMaxTxs(int) types.Txs                  { return types.Txs{} }
func (emptyMempool) Update(
	int64,
	types.Txs,
	[]*abci.ExecTxResult,
	mempl.PreCheckFunc,
	mempl.PostCheckFunc,
) error {
	return nil
}
func (emptyMempool) Flush()                                 {}
func (emptyMempool) FlushAppConn() error                    { return nil }
func (emptyMempool) TxsAvailable() <-chan struct{}          { return make(chan struct{}) }
func (emptyMempool) EnableTxsAvailable()                    {}
func (emptyMempool) SetTxRemovedCallback(func(types.TxKey)) {}
func (emptyMempool) TxsBytes() int64                        { return 0 }
func (emptyMempool) InMempool(types.TxKey) bool             { return false }

func (emptyMempool) TxsFront() *clist.CElement    { return nil }
func (emptyMempool) TxsWaitChan() <-chan struct{} { return nil }

func (emptyMempool) InitWAL() error { return nil }
func (emptyMempool) CloseWAL()      {}

//-----------------------------------------------------------------------------
// mockProxyApp uses ABCIResponses to give the right results.
//
// Useful because we don't want to call Commit() twice for the same block on
// the real app.

func newMockProxyApp(finalizeBlockResponse *abci.ResponseFinalizeBlock) proxy.AppConnConsensus {
	clientCreator := proxy.NewLocalClientCreator(&mockProxyApp{
		finalizeBlockResponse: finalizeBlockResponse,
	})
	cli, _ := clientCreator.NewABCIConsensusClient()
	err := cli.Start()
	if err != nil {
		panic(err)
	}
	return proxy.NewAppConnConsensus(cli, proxy.NopMetrics())
}

type mockProxyApp struct {
	abci.BaseApplication
	finalizeBlockResponse *abci.ResponseFinalizeBlock
}

func (mock *mockProxyApp) FinalizeBlock(context.Context, *abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error) {
	return mock.finalizeBlockResponse, nil
}
