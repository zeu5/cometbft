package test

import "github.com/zeu5/cometbft/types"

func MakeNTxs(height, n int64) types.Txs {
	txs := make([]types.Tx, n)
	for i := range txs {
		txs[i] = types.Tx([]byte{byte(height), byte(i / 256), byte(i % 256)})
	}
	return txs
}
