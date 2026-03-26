package cli

import (
	"context"

	"github.com/ankele/pvm/internal/model"
	"github.com/spf13/cobra"
)

func newInterfaceCommand(deps Dependencies, opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{Use: "iface", Short: "Manage host interfaces"}

	cmd.AddCommand(simpleListCommand("list", "List interfaces", func(ctx context.Context, mgr *serviceLike) (any, error) {
		return mgr.ListInterfaces(ctx)
	}, deps, opts))
	cmd.AddCommand(simpleNameCommand("get", "Get an interface", func(ctx context.Context, mgr *serviceLike, name string) (any, error) {
		return mgr.GetInterface(ctx, name)
	}, deps, opts))
	cmd.AddCommand(simpleNameCommand("start", "Start an interface", func(ctx context.Context, mgr *serviceLike, name string) (any, error) {
		return mgr.StartInterface(ctx, name)
	}, deps, opts))
	cmd.AddCommand(simpleNameCommand("destroy", "Destroy an interface", func(ctx context.Context, mgr *serviceLike, name string) (any, error) {
		return mgr.DestroyInterface(ctx, name)
	}, deps, opts))
	cmd.AddCommand(simpleNameCommand("undefine", "Undefine an interface", func(ctx context.Context, mgr *serviceLike, name string) (any, error) {
		return mgr.UndefineInterface(ctx, name)
	}, deps, opts))

	defineCmd := &cobra.Command{
		Use:   "define",
		Short: "Define a host interface",
		RunE: func(cmd *cobra.Command, _ []string) error {
			spec := model.InterfaceDefineSpec{
				Name:          mustGetString(cmd, "name"),
				Type:          mustGetString(cmd, "type"),
				MAC:           mustGetString(cmd, "mac"),
				StartMode:     mustGetString(cmd, "start-mode"),
				MTU:           mustGetUint32(cmd, "mtu"),
				BridgeMembers: mustGetStringSlice(cmd, "member"),
			}
			return runWithManager(cmd, opts, deps, func(ctx context.Context, _ *cobra.Command, print func(any) error, mgr *serviceLike) error {
				iface, err := mgr.DefineInterface(ctx, spec)
				if err != nil {
					return err
				}
				return print(iface)
			})
		},
	}
	defineCmd.Flags().String("name", "", "interface name")
	defineCmd.Flags().String("type", "bridge", "interface type")
	defineCmd.Flags().String("mac", "", "interface MAC address")
	defineCmd.Flags().String("start-mode", "onboot", "start mode")
	defineCmd.Flags().Uint32("mtu", 0, "MTU")
	defineCmd.Flags().StringArray("member", nil, "bridge or bond member")
	cmd.AddCommand(defineCmd)

	return cmd
}
