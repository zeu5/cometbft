package http_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zeu5/cometbft/abci/example/kvstore"
	"github.com/zeu5/cometbft/light/provider"
	lighthttp "github.com/zeu5/cometbft/light/provider/http"
	rpcclient "github.com/zeu5/cometbft/rpc/client"
	rpchttp "github.com/zeu5/cometbft/rpc/client/http"
	rpctest "github.com/zeu5/cometbft/rpc/test"
	"github.com/zeu5/cometbft/types"
)

func TestNewProvider(t *testing.T) {
	c, err := lighthttp.New("chain-test", "192.168.0.1:26657")
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf("%s", c), "http{http://192.168.0.1:26657}")

	c, err = lighthttp.New("chain-test", "http://153.200.0.1:26657")
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf("%s", c), "http{http://153.200.0.1:26657}")

	c, err = lighthttp.New("chain-test", "153.200.0.1")
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf("%s", c), "http{http://153.200.0.1}")
}

func TestProvider(t *testing.T) {
	for _, path := range []string{"", "/", "/v1", "/v1/"} {
		app := kvstore.NewInMemoryApplication()
		app.RetainBlocks = 10
		node := rpctest.StartCometBFT(app, rpctest.RecreateConfig)

		cfg := rpctest.GetConfig()
		defer os.RemoveAll(cfg.RootDir)
		rpcAddr := cfg.RPC.ListenAddress
		genDoc, err := types.GenesisDocFromFile(cfg.GenesisFile())
		require.NoError(t, err)
		chainID := genDoc.ChainID

		c, err := rpchttp.New(rpcAddr + path)
		require.Nil(t, err)

		p := lighthttp.NewWithClient(chainID, c)
		require.NoError(t, err)
		require.NotNil(t, p)

		// let it produce some blocks
		err = rpcclient.WaitForHeight(c, 10, nil)
		require.NoError(t, err)

		// let's get the highest block
		lb, err := p.LightBlock(context.Background(), 0)
		require.NoError(t, err)
		require.NotNil(t, lb)
		assert.True(t, lb.Height < 1000)
		assert.True(t, lb.Height >= 10)

		// let's check this is valid somehow
		assert.Nil(t, lb.ValidateBasic(chainID))

		// historical queries now work :)
		lb, err = p.LightBlock(context.Background(), 0)
		require.NoError(t, err)
		require.NotNil(t, lb)
		lower := lb.Height - 3
		lb, err = p.LightBlock(context.Background(), lower)
		require.NoError(t, err)
		assert.Equal(t, lower, lb.Height)

		// fetching missing heights (both future and pruned) should return appropriate errors
		lb, err = p.LightBlock(context.Background(), 0)
		require.NoError(t, err)
		require.NotNil(t, lb)
		lb, err = p.LightBlock(context.Background(), lb.Height+1000)
		require.Error(t, err)
		require.Nil(t, lb)
		assert.Equal(t, provider.ErrHeightTooHigh, err)

		_, err = p.LightBlock(context.Background(), 1)
		require.Error(t, err)
		require.Nil(t, lb)
		assert.Equal(t, provider.ErrLightBlockNotFound, err)

		// stop the full node and check that a no response error is returned
		rpctest.StopCometBFT(node)
		time.Sleep(10 * time.Second)
		lb, err = p.LightBlock(context.Background(), lower+2)
		// we should see a connection refused
		require.Error(t, err)
		require.Contains(t, err.Error(), "connection refused")
		require.Nil(t, lb)
	}
}
