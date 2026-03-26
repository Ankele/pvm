package cli

import (
	"context"

	"github.com/ankele/pvm/internal/model"
	"github.com/spf13/cobra"
)

func newNetworkCommand(deps Dependencies, opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{Use: "network", Short: "Manage libvirt virtual networks"}

	cmd.AddCommand(simpleListCommand("list", "List networks", func(ctx context.Context, mgr *serviceLike) (any, error) {
		return mgr.ListNetworks(ctx)
	}, deps, opts))
	cmd.AddCommand(simpleNameCommand("get", "Get a network", func(ctx context.Context, mgr *serviceLike, name string) (any, error) {
		return mgr.GetNetwork(ctx, name)
	}, deps, opts))
	cmd.AddCommand(simpleNameCommand("start", "Start a network", func(ctx context.Context, mgr *serviceLike, name string) (any, error) {
		return mgr.StartNetwork(ctx, name)
	}, deps, opts))
	cmd.AddCommand(simpleNameCommand("destroy", "Destroy a network", func(ctx context.Context, mgr *serviceLike, name string) (any, error) {
		return mgr.DestroyNetwork(ctx, name)
	}, deps, opts))
	cmd.AddCommand(simpleNameCommand("undefine", "Undefine a network", func(ctx context.Context, mgr *serviceLike, name string) (any, error) {
		return mgr.UndefineNetwork(ctx, name)
	}, deps, opts))

	defineCmd := &cobra.Command{
		Use:   "define",
		Short: "Define a virtual network",
		RunE: func(cmd *cobra.Command, _ []string) error {
			spec := model.NetworkDefineSpec{
				Name:        mustGetString(cmd, "name"),
				Bridge:      mustGetString(cmd, "bridge"),
				ForwardMode: mustGetString(cmd, "forward-mode"),
				Domain:      mustGetString(cmd, "domain"),
				IPv4CIDR:    mustGetString(cmd, "ipv4-cidr"),
				DHCPStart:   mustGetString(cmd, "dhcp-start"),
				DHCPEnd:     mustGetString(cmd, "dhcp-end"),
				Autostart:   mustGetBool(cmd, "autostart"),
			}
			return runWithManager(cmd, opts, deps, func(ctx context.Context, _ *cobra.Command, print func(any) error, mgr *serviceLike) error {
				net, err := mgr.DefineNetwork(ctx, spec)
				if err != nil {
					return err
				}
				return print(net)
			})
		},
	}
	defineCmd.Flags().String("name", "", "network name")
	defineCmd.Flags().String("bridge", "", "bridge name")
	defineCmd.Flags().String("forward-mode", "nat", "forward mode")
	defineCmd.Flags().String("domain", "", "DNS domain")
	defineCmd.Flags().String("ipv4-cidr", "", "IPv4 CIDR")
	defineCmd.Flags().String("dhcp-start", "", "DHCP range start")
	defineCmd.Flags().String("dhcp-end", "", "DHCP range end")
	defineCmd.Flags().Bool("autostart", false, "enable autostart")
	cmd.AddCommand(defineCmd)

	autostartCmd := &cobra.Command{
		Use:   "autostart <name>",
		Short: "Set network autostart",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			enable := mustGetBool(cmd, "enable")
			return runWithManager(cmd, opts, deps, func(ctx context.Context, _ *cobra.Command, print func(any) error, mgr *serviceLike) error {
				net, err := mgr.SetNetworkAutostart(ctx, args[0], enable)
				if err != nil {
					return err
				}
				return print(net)
			})
		},
	}
	autostartCmd.Flags().Bool("enable", true, "enable autostart")
	cmd.AddCommand(autostartCmd)

	return cmd
}
