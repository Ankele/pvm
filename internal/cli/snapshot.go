package cli

import (
	"context"

	"github.com/ankele/pvm/internal/model"
	"github.com/spf13/cobra"
)

func newSnapshotCommand(deps Dependencies, opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{Use: "snapshot", Short: "Manage VM and volume snapshots"}

	vmCmd := &cobra.Command{Use: "vm", Short: "Manage VM snapshots"}
	vmCreate := &cobra.Command{
		Use:   "create",
		Short: "Create a VM snapshot",
		RunE: func(cmd *cobra.Command, _ []string) error {
			spec := model.VMSnapshotCreateSpec{
				VM:            mustGetString(cmd, "vm"),
				Name:          mustGetString(cmd, "name"),
				Description:   mustGetString(cmd, "description"),
				IncludeMemory: mustGetBool(cmd, "memory"),
				Quiesce:       mustGetBool(cmd, "quiesce"),
			}
			return runWithManager(cmd, opts, deps, func(ctx context.Context, _ *cobra.Command, print func(any) error, mgr *serviceLike) error {
				snap, err := mgr.CreateDomainSnapshot(ctx, spec)
				if err != nil {
					return err
				}
				return print(snap)
			})
		},
	}
	vmCreate.Flags().String("vm", "", "VM name")
	vmCreate.Flags().String("name", "", "snapshot name")
	vmCreate.Flags().String("description", "", "snapshot description")
	vmCreate.Flags().Bool("memory", false, "include memory state")
	vmCreate.Flags().Bool("quiesce", false, "quiesce guest filesystem")
	vmCmd.AddCommand(vmCreate)

	vmList := &cobra.Command{
		Use:   "list",
		Short: "List VM snapshots",
		RunE: func(cmd *cobra.Command, _ []string) error {
			vm := mustGetString(cmd, "vm")
			return runWithManager(cmd, opts, deps, func(ctx context.Context, _ *cobra.Command, print func(any) error, mgr *serviceLike) error {
				items, err := mgr.ListDomainSnapshots(ctx, vm)
				if err != nil {
					return err
				}
				return print(items)
			})
		},
	}
	vmList.Flags().String("vm", "", "VM name")
	vmCmd.AddCommand(vmList)

	vmDelete := &cobra.Command{
		Use:   "delete",
		Short: "Delete a VM snapshot",
		RunE: func(cmd *cobra.Command, _ []string) error {
			vm := mustGetString(cmd, "vm")
			name := mustGetString(cmd, "name")
			return runWithManager(cmd, opts, deps, func(ctx context.Context, _ *cobra.Command, print func(any) error, mgr *serviceLike) error {
				result, err := mgr.DeleteDomainSnapshot(ctx, vm, name)
				if err != nil {
					return err
				}
				return print(result)
			})
		},
	}
	vmDelete.Flags().String("vm", "", "VM name")
	vmDelete.Flags().String("name", "", "snapshot name")
	vmCmd.AddCommand(vmDelete)

	vmRevert := &cobra.Command{
		Use:   "revert",
		Short: "Revert a VM snapshot",
		RunE: func(cmd *cobra.Command, _ []string) error {
			vm := mustGetString(cmd, "vm")
			name := mustGetString(cmd, "name")
			return runWithManager(cmd, opts, deps, func(ctx context.Context, _ *cobra.Command, print func(any) error, mgr *serviceLike) error {
				snap, err := mgr.RevertDomainSnapshot(ctx, vm, name)
				if err != nil {
					return err
				}
				return print(snap)
			})
		},
	}
	vmRevert.Flags().String("vm", "", "VM name")
	vmRevert.Flags().String("name", "", "snapshot name")
	vmCmd.AddCommand(vmRevert)
	cmd.AddCommand(vmCmd)

	volumeCmd := &cobra.Command{Use: "volume", Short: "Manage volume snapshots"}
	volumeCreate := &cobra.Command{
		Use:   "create",
		Short: "Create a volume snapshot",
		RunE: func(cmd *cobra.Command, _ []string) error {
			spec := model.VolumeSnapshotCreateSpec{
				Pool:   mustGetString(cmd, "pool"),
				Volume: mustGetString(cmd, "volume"),
				Name:   mustGetString(cmd, "name"),
			}
			return runWithManager(cmd, opts, deps, func(ctx context.Context, _ *cobra.Command, print func(any) error, mgr *serviceLike) error {
				snap, err := mgr.CreateVolumeSnapshot(ctx, spec)
				if err != nil {
					return err
				}
				return print(snap)
			})
		},
	}
	volumeCreate.Flags().String("pool", "", "pool name")
	volumeCreate.Flags().String("volume", "", "volume name")
	volumeCreate.Flags().String("name", "", "snapshot name")
	volumeCmd.AddCommand(volumeCreate)

	volumeList := &cobra.Command{
		Use:   "list",
		Short: "List volume snapshots",
		RunE: func(cmd *cobra.Command, _ []string) error {
			pool := mustGetString(cmd, "pool")
			volume := mustGetString(cmd, "volume")
			return runWithManager(cmd, opts, deps, func(ctx context.Context, _ *cobra.Command, print func(any) error, mgr *serviceLike) error {
				items, err := mgr.ListVolumeSnapshots(ctx, pool, volume)
				if err != nil {
					return err
				}
				return print(items)
			})
		},
	}
	volumeList.Flags().String("pool", "", "pool name")
	volumeList.Flags().String("volume", "", "volume name")
	volumeCmd.AddCommand(volumeList)

	volumeDelete := &cobra.Command{
		Use:   "delete",
		Short: "Delete a volume snapshot",
		RunE: func(cmd *cobra.Command, _ []string) error {
			pool := mustGetString(cmd, "pool")
			volume := mustGetString(cmd, "volume")
			name := mustGetString(cmd, "name")
			return runWithManager(cmd, opts, deps, func(ctx context.Context, _ *cobra.Command, print func(any) error, mgr *serviceLike) error {
				result, err := mgr.DeleteVolumeSnapshot(ctx, pool, volume, name)
				if err != nil {
					return err
				}
				return print(result)
			})
		},
	}
	volumeDelete.Flags().String("pool", "", "pool name")
	volumeDelete.Flags().String("volume", "", "volume name")
	volumeDelete.Flags().String("name", "", "snapshot name")
	volumeCmd.AddCommand(volumeDelete)

	volumeRollback := &cobra.Command{
		Use:   "rollback",
		Short: "Rollback a volume snapshot",
		RunE: func(cmd *cobra.Command, _ []string) error {
			pool := mustGetString(cmd, "pool")
			volume := mustGetString(cmd, "volume")
			name := mustGetString(cmd, "name")
			return runWithManager(cmd, opts, deps, func(ctx context.Context, _ *cobra.Command, print func(any) error, mgr *serviceLike) error {
				snap, err := mgr.RollbackVolumeSnapshot(ctx, pool, volume, name)
				if err != nil {
					return err
				}
				return print(snap)
			})
		},
	}
	volumeRollback.Flags().String("pool", "", "pool name")
	volumeRollback.Flags().String("volume", "", "volume name")
	volumeRollback.Flags().String("name", "", "snapshot name")
	volumeCmd.AddCommand(volumeRollback)
	cmd.AddCommand(volumeCmd)

	return cmd
}
