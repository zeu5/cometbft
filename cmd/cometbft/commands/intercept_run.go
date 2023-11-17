package commands

import (
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"
	cmtos "github.com/zeu5/cometbft/libs/os"
	nm "github.com/zeu5/cometbft/node"
)

func AddInterceptFlags(cmd *cobra.Command) {
	cmd.Flags().String(
		"p2p.intercept_listen_addr",
		config.P2P.TestInterceptConfig.ListenAddr,
		"Address to listen to while intercepting messages",
	)

	cmd.Flags().String(
		"p2p.intercept_server_addr",
		config.P2P.TestInterceptConfig.ServerAddr,
		"Address to send intercepted messages to",
	)

	cmd.Flags().Int(
		"p2p.intercept_id",
		config.P2P.TestInterceptConfig.ID,
		"Id of the node used to register with the interceptor",
	)
}

func NewInterceptRunNodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "istart",
		Aliases: []string{"inode", "irun"},
		Short:   "Run the CometBFT node with a message interceptor",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(genesisHash) != 0 {
				config.Storage.GenesisHash = hex.EncodeToString(genesisHash)
			}
			config.P2P.TestIntercept = true

			n, err := nm.InterceptNode(config, logger)
			if err != nil {
				return fmt.Errorf("failed to create node: %w", err)
			}

			if err := n.Start(); err != nil {
				return fmt.Errorf("failed to start node: %w", err)
			}

			logger.Info("Started node", "nodeInfo", n.Switch().NodeInfo())

			// Stop upon receiving SIGTERM or CTRL-C.
			cmtos.TrapSignal(logger, func() {
				if n.IsRunning() {
					if err := n.Stop(); err != nil {
						logger.Error("unable to stop the node", "error", err)
					}
				}
			})

			// Run forever.
			select {}
		},
	}
	AddNodeFlags(cmd)
	AddInterceptFlags(cmd)
	return cmd
}
