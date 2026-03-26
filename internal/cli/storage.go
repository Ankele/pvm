package cli

import (
	"context"

	"github.com/ankele/pvm/internal/model"
	"github.com/spf13/cobra"
)

func newPoolCommand(deps Dependencies, opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{Use: "pool", Short: "Manage storage pools"}

	cmd.AddCommand(simpleListCommand("list", "List storage pools", func(ctx context.Context, mgr *serviceLike) (any, error) {
		return mgr.ListPools(ctx)
	}, deps, opts))

	cmd.AddCommand(simpleNameCommand("get", "Get a storage pool", func(ctx context.Context, mgr *serviceLike, name string) (any, error) {
		return mgr.GetPool(ctx, name)
	}, deps, opts))
	cmd.AddCommand(simpleNameCommand("start", "Start a storage pool", func(ctx context.Context, mgr *serviceLike, name string) (any, error) {
		return mgr.StartPool(ctx, name)
	}, deps, opts))
	cmd.AddCommand(simpleNameCommand("destroy", "Destroy a storage pool", func(ctx context.Context, mgr *serviceLike, name string) (any, error) {
		return mgr.DestroyPool(ctx, name)
	}, deps, opts))
	cmd.AddCommand(simpleNameCommand("refresh", "Refresh a storage pool", func(ctx context.Context, mgr *serviceLike, name string) (any, error) {
		return mgr.RefreshPool(ctx, name)
	}, deps, opts))
	cmd.AddCommand(simpleNameCommand("undefine", "Undefine a storage pool", func(ctx context.Context, mgr *serviceLike, name string) (any, error) {
		return mgr.UndefinePool(ctx, name)
	}, deps, opts))

	defineCmd := &cobra.Command{
		Use:   "define",
		Short: "Define a storage pool",
		RunE: func(cmd *cobra.Command, _ []string) error {
			spec := model.PoolDefineSpec{
				Name:       mustGetString(cmd, "name"),
				Type:       mustGetString(cmd, "type"),
				SourceName: mustGetString(cmd, "source-name"),
				TargetPath: mustGetString(cmd, "target-path"),
				Autostart:  mustGetBool(cmd, "autostart"),
			}
			return runWithManager(cmd, opts, deps, func(ctx context.Context, _ *cobra.Command, print func(any) error, mgr *serviceLike) error {
				pool, err := mgr.DefinePool(ctx, spec)
				if err != nil {
					return err
				}
				return print(pool)
			})
		},
	}
	defineCmd.Flags().String("name", "", "pool name")
	defineCmd.Flags().String("type", "zfs", "pool type")
	defineCmd.Flags().String("source-name", "", "storage source name")
	defineCmd.Flags().String("target-path", "", "target path")
	defineCmd.Flags().Bool("autostart", false, "enable autostart")
	cmd.AddCommand(defineCmd)

	autostartCmd := &cobra.Command{
		Use:   "autostart <name>",
		Short: "Set pool autostart",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			enable := mustGetBool(cmd, "enable")
			return runWithManager(cmd, opts, deps, func(ctx context.Context, _ *cobra.Command, print func(any) error, mgr *serviceLike) error {
				pool, err := mgr.SetPoolAutostart(ctx, args[0], enable)
				if err != nil {
					return err
				}
				return print(pool)
			})
		},
	}
	autostartCmd.Flags().Bool("enable", true, "enable autostart")
	cmd.AddCommand(autostartCmd)

	return cmd
}

func newVolumeCommand(deps Dependencies, opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{Use: "volume", Short: "Manage storage volumes"}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List volumes in a pool",
		RunE: func(cmd *cobra.Command, _ []string) error {
			pool := mustGetString(cmd, "pool")
			return runWithManager(cmd, opts, deps, func(ctx context.Context, _ *cobra.Command, print func(any) error, mgr *serviceLike) error {
				items, err := mgr.ListVolumes(ctx, pool)
				if err != nil {
					return err
				}
				return print(items)
			})
		},
	}
	listCmd.Flags().String("pool", "", "pool name")
	cmd.AddCommand(listCmd)

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get a volume",
		RunE: func(cmd *cobra.Command, _ []string) error {
			pool := mustGetString(cmd, "pool")
			name := mustGetString(cmd, "name")
			return runWithManager(cmd, opts, deps, func(ctx context.Context, _ *cobra.Command, print func(any) error, mgr *serviceLike) error {
				vol, err := mgr.GetVolume(ctx, pool, name)
				if err != nil {
					return err
				}
				return print(vol)
			})
		},
	}
	getCmd.Flags().String("pool", "", "pool name")
	getCmd.Flags().String("name", "", "volume name")
	cmd.AddCommand(getCmd)

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a volume",
		RunE: func(cmd *cobra.Command, _ []string) error {
			spec := model.VolumeCreateSpec{
				Pool:          mustGetString(cmd, "pool"),
				Name:          mustGetString(cmd, "name"),
				CapacityBytes: mustGetUint64(cmd, "capacity-bytes"),
				Format:        mustGetString(cmd, "format"),
				Type:          mustGetString(cmd, "type"),
			}
			return runWithManager(cmd, opts, deps, func(ctx context.Context, _ *cobra.Command, print func(any) error, mgr *serviceLike) error {
				vol, err := mgr.CreateVolume(ctx, spec)
				if err != nil {
					return err
				}
				return print(vol)
			})
		},
	}
	createCmd.Flags().String("pool", "", "pool name")
	createCmd.Flags().String("name", "", "volume name")
	createCmd.Flags().Uint64("capacity-bytes", 0, "volume capacity in bytes")
	createCmd.Flags().String("format", "raw", "volume format")
	createCmd.Flags().String("type", "", "volume type")
	cmd.AddCommand(createCmd)

	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a volume",
		RunE: func(cmd *cobra.Command, _ []string) error {
			pool := mustGetString(cmd, "pool")
			name := mustGetString(cmd, "name")
			return runWithManager(cmd, opts, deps, func(ctx context.Context, _ *cobra.Command, print func(any) error, mgr *serviceLike) error {
				result, err := mgr.DeleteVolume(ctx, pool, name)
				if err != nil {
					return err
				}
				return print(result)
			})
		},
	}
	deleteCmd.Flags().String("pool", "", "pool name")
	deleteCmd.Flags().String("name", "", "volume name")
	cmd.AddCommand(deleteCmd)

	resizeCmd := &cobra.Command{
		Use:   "resize",
		Short: "Resize a volume",
		RunE: func(cmd *cobra.Command, _ []string) error {
			spec := model.VolumeResizeSpec{
				Pool:          mustGetString(cmd, "pool"),
				Name:          mustGetString(cmd, "name"),
				CapacityBytes: mustGetUint64(cmd, "capacity-bytes"),
			}
			return runWithManager(cmd, opts, deps, func(ctx context.Context, _ *cobra.Command, print func(any) error, mgr *serviceLike) error {
				vol, err := mgr.ResizeVolume(ctx, spec)
				if err != nil {
					return err
				}
				return print(vol)
			})
		},
	}
	resizeCmd.Flags().String("pool", "", "pool name")
	resizeCmd.Flags().String("name", "", "volume name")
	resizeCmd.Flags().Uint64("capacity-bytes", 0, "new capacity in bytes")
	cmd.AddCommand(resizeCmd)

	return cmd
}

func simpleListCommand(use, short string, fn func(context.Context, *serviceLike) (any, error), deps Dependencies, opts *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runWithManager(cmd, opts, deps, func(ctx context.Context, _ *cobra.Command, print func(any) error, mgr *serviceLike) error {
				value, err := fn(ctx, mgr)
				if err != nil {
					return err
				}
				return print(value)
			})
		},
	}
}

func simpleNameCommand(use, short string, fn func(context.Context, *serviceLike, string) (any, error), deps Dependencies, opts *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   use + " <name>",
		Short: short,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWithManager(cmd, opts, deps, func(ctx context.Context, _ *cobra.Command, print func(any) error, mgr *serviceLike) error {
				value, err := fn(ctx, mgr, args[0])
				if err != nil {
					return err
				}
				return print(value)
			})
		},
	}
}
