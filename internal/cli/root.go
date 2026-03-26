package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ankele/pvm/internal/backend"
	"github.com/ankele/pvm/internal/connection"
	"github.com/ankele/pvm/internal/model"
	"github.com/ankele/pvm/internal/render"
	"github.com/ankele/pvm/internal/service"
	"github.com/spf13/cobra"
)

type Dependencies struct {
	NewBackend backend.Factory
	Stdout     io.Writer
	Stderr     io.Writer
}

type rootOptions struct {
	uri           string
	username      string
	output        string
	passwordStdin bool
}

func NewRootCommand(deps Dependencies) *cobra.Command {
	if deps.NewBackend == nil {
		deps.NewBackend = backend.New
	}
	if deps.Stdout == nil {
		deps.Stdout = os.Stdout
	}
	if deps.Stderr == nil {
		deps.Stderr = os.Stderr
	}

	opts := &rootOptions{}
	cmd := &cobra.Command{
		Use:           "pvm",
		Short:         "Manage libvirt virtual machines, storage and networks",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.PersistentFlags().StringVar(&opts.uri, "uri", "", "libvirt connection URI")
	cmd.PersistentFlags().StringVar(&opts.username, "username", "", "libvirt username")
	cmd.PersistentFlags().StringVarP(&opts.output, "output", "o", string(model.OutputText), "output format: text|json")
	cmd.PersistentFlags().BoolVar(&opts.passwordStdin, "password-stdin", false, "read libvirt password from stdin")

	cmd.AddCommand(newServeCommand(deps, opts))
	cmd.AddCommand(newVMCommand(deps, opts))
	cmd.AddCommand(newPoolCommand(deps, opts))
	cmd.AddCommand(newVolumeCommand(deps, opts))
	cmd.AddCommand(newNetworkCommand(deps, opts))
	cmd.AddCommand(newInterfaceCommand(deps, opts))
	cmd.AddCommand(newSnapshotCommand(deps, opts))

	cmd.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
		return fmt.Errorf("%s: %w", c.CommandPath(), err)
	})

	cmd.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		switch model.OutputFormat(opts.output) {
		case model.OutputText, model.OutputJSON:
			return nil
		default:
			return fmt.Errorf("unsupported output format %q", opts.output)
		}
	}

	return cmd
}

func (o *rootOptions) newManager(ctx context.Context, factory backend.Factory) (*service.Manager, error) {
	password := ""
	if o.passwordStdin {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, err
		}
		password = strings.TrimSpace(string(data))
	}
	cfg := connection.Resolve(o.uri, o.username, password)
	b, err := factory(ctx, backend.Options{
		URI:      cfg.URI,
		Username: cfg.Username,
		Password: cfg.Password,
	})
	if err != nil {
		return nil, err
	}
	return service.NewManager(b), nil
}

func (o *rootOptions) renderer(stdout io.Writer) render.Renderer {
	return render.Renderer{
		Out:    stdout,
		Format: model.OutputFormat(o.output),
	}
}
