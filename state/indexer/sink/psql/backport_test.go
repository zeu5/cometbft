package psql

import (
	"github.com/zeu5/cometbft/state/indexer"
	"github.com/zeu5/cometbft/state/txindex"
)

var (
	_ indexer.BlockIndexer = BackportBlockIndexer{}
	_ txindex.TxIndexer    = BackportTxIndexer{}
)
