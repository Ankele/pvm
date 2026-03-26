package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ankele/pvm/internal/server"
	"github.com/spf13/cobra"
)

func newServeCommand(deps Dependencies, opts *rootOptions) *cobra.Command {
	var httpAddr string
	var grpcAddr string
	var tokenFile string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run HTTP and gRPC API servers",
		RunE: func(cmd *cobra.Command, _ []string) error {
			manager, err := opts.newManager(cmd.Context(), deps.NewBackend)
			if err != nil {
				return err
			}
			defer manager.Close()

			token := strings.TrimSpace(os.Getenv("PVM_API_TOKEN"))
			if token == "" && strings.TrimSpace(tokenFile) != "" {
				data, err := os.ReadFile(tokenFile)
				if err != nil {
					return fmt.Errorf("read token file: %w", err)
				}
				token = strings.TrimSpace(string(data))
			}
			if token == "" {
				return fmt.Errorf("api token is required via --token-file or PVM_API_TOKEN")
			}

			srv := server.New(server.Config{
				HTTPAddr: httpAddr,
				GRPCAddr: grpcAddr,
				Token:    token,
			}, manager)
			return srv.Run(context.WithoutCancel(cmd.Context()))
		},
	}

	cmd.Flags().StringVar(&httpAddr, "http-addr", "127.0.0.1:8080", "HTTP listen address")
	cmd.Flags().StringVar(&grpcAddr, "grpc-addr", "127.0.0.1:9090", "gRPC listen address")
	cmd.Flags().StringVar(&tokenFile, "token-file", "", "file containing API token")
	return cmd
}
