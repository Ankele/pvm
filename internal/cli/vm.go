package cli

import (
	"context"
	"fmt"

	"github.com/ankele/pvm/internal/model"
	"github.com/spf13/cobra"
)

func newVMCommand(deps Dependencies, opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vm",
		Short: "Manage virtual machines",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List virtual machines",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runWithManager(cmd, opts, deps, func(ctx context.Context, _ *cobra.Command, print func(any) error, mgr *serviceLike) error {
				items, err := mgr.ListVMs(ctx)
				if err != nil {
					return err
				}
				return print(items)
			})
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "get <name>",
		Short: "Get a virtual machine",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWithManager(cmd, opts, deps, func(ctx context.Context, _ *cobra.Command, print func(any) error, mgr *serviceLike) error {
				vm, err := mgr.GetVM(ctx, args[0])
				if err != nil {
					return err
				}
				return print(vm)
			})
		},
	})

	defineCmd := &cobra.Command{
		Use:   "define --file spec.yaml",
		Short: "Define a VM from YAML or JSON spec",
		RunE: func(cmd *cobra.Command, _ []string) error {
			specFile := mustGetString(cmd, "file")
			if specFile == "" {
				return fmt.Errorf("--file is required")
			}
			spec, err := loadDomainSpec(specFile)
			if err != nil {
				return err
			}
			return runWithManager(cmd, opts, deps, func(ctx context.Context, _ *cobra.Command, print func(any) error, mgr *serviceLike) error {
				vm, err := mgr.DefineVM(ctx, spec)
				if err != nil {
					return err
				}
				return print(vm)
			})
		},
	}
	defineCmd.Flags().String("file", "", "path to YAML or JSON domain spec")
	cmd.AddCommand(defineCmd)

	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Create a VM and boot it from an ISO",
		RunE: func(cmd *cobra.Command, _ []string) error {
			spec := model.VMInstallSpec{
				Name:        mustGetString(cmd, "name"),
				Description: mustGetString(cmd, "description"),
				MemoryMiB:   mustGetUint64(cmd, "memory-mib"),
				VCPU:        mustGetUint32(cmd, "vcpu"),
				Machine:     mustGetString(cmd, "machine"),
				Arch:        mustGetString(cmd, "arch"),
				Pool:        mustGetString(cmd, "pool"),
				DiskName:    mustGetString(cmd, "disk-name"),
				DiskSizeGiB: mustGetUint64(cmd, "disk-size-gib"),
				DiskFormat:  mustGetString(cmd, "disk-format"),
				ISOPath:     mustGetString(cmd, "iso-path"),
				Networks:    collectNetworks(cmd),
				Graphics:    collectGraphics(cmd),
				Autostart:   mustGetBool(cmd, "autostart"),
			}
			return runWithManager(cmd, opts, deps, func(ctx context.Context, _ *cobra.Command, print func(any) error, mgr *serviceLike) error {
				vm, err := mgr.InstallVM(ctx, spec)
				if err != nil {
					return err
				}
				return print(vm)
			})
		},
	}
	installCmd.Flags().String("name", "", "VM name")
	installCmd.Flags().String("description", "", "VM description")
	installCmd.Flags().Uint64("memory-mib", 2048, "memory size in MiB")
	installCmd.Flags().Uint32("vcpu", 2, "number of virtual CPUs")
	installCmd.Flags().String("machine", "", "machine type")
	installCmd.Flags().String("arch", "", "guest architecture")
	installCmd.Flags().String("pool", "", "target storage pool")
	installCmd.Flags().String("disk-name", "", "root disk volume name")
	installCmd.Flags().Uint64("disk-size-gib", 20, "root disk size in GiB")
	installCmd.Flags().String("disk-format", "raw", "disk format")
	installCmd.Flags().String("iso-path", "", "path to ISO image")
	installCmd.Flags().StringArray("network", []string{"default"}, "network to attach")
	installCmd.Flags().Bool("autostart", false, "enable autostart")
	defaultGraphicsFlags(installCmd)
	cmd.AddCommand(installCmd)

	launchCmd := &cobra.Command{
		Use:   "launch",
		Short: "Create a VM from an installed image",
		RunE: func(cmd *cobra.Command, _ []string) error {
			spec := model.VMLaunchSpec{
				Name:             mustGetString(cmd, "name"),
				Description:      mustGetString(cmd, "description"),
				Mode:             model.LaunchMode(mustGetString(cmd, "mode")),
				MemoryMiB:        mustGetUint64(cmd, "memory-mib"),
				VCPU:             mustGetUint32(cmd, "vcpu"),
				Machine:          mustGetString(cmd, "machine"),
				Arch:             mustGetString(cmd, "arch"),
				ImagePath:        mustGetString(cmd, "image-path"),
				SourceVolume:     mustGetString(cmd, "source-volume"),
				TargetPool:       mustGetString(cmd, "target-pool"),
				TargetVolumeName: mustGetString(cmd, "target-volume"),
				TargetFormat:     mustGetString(cmd, "target-format"),
				Networks:         collectNetworks(cmd),
				Graphics:         collectGraphics(cmd),
				Autostart:        mustGetBool(cmd, "autostart"),
			}
			return runWithManager(cmd, opts, deps, func(ctx context.Context, _ *cobra.Command, print func(any) error, mgr *serviceLike) error {
				vm, err := mgr.LaunchVM(ctx, spec)
				if err != nil {
					return err
				}
				return print(vm)
			})
		},
	}
	launchCmd.Flags().String("name", "", "VM name")
	launchCmd.Flags().String("description", "", "VM description")
	launchCmd.Flags().String("mode", "", "launch mode: clone|direct")
	launchCmd.Flags().Uint64("memory-mib", 2048, "memory size in MiB")
	launchCmd.Flags().Uint32("vcpu", 2, "number of virtual CPUs")
	launchCmd.Flags().String("machine", "", "machine type")
	launchCmd.Flags().String("arch", "", "guest architecture")
	launchCmd.Flags().String("image-path", "", "path to qcow2/raw image")
	launchCmd.Flags().String("source-volume", "", "source volume in pool/name form")
	launchCmd.Flags().String("target-pool", "", "target storage pool for clone mode")
	launchCmd.Flags().String("target-volume", "", "target volume name for clone mode")
	launchCmd.Flags().String("target-format", "raw", "target volume format")
	launchCmd.Flags().StringArray("network", []string{"default"}, "network to attach")
	launchCmd.Flags().Bool("autostart", false, "enable autostart")
	defaultGraphicsFlags(launchCmd)
	cmd.AddCommand(launchCmd)

	addNamedVMAction := func(use, short string, action func(context.Context, *serviceLike, string) (any, error)) {
		cmd.AddCommand(&cobra.Command{
			Use:   use + " <name>",
			Short: short,
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return runWithManager(cmd, opts, deps, func(ctx context.Context, _ *cobra.Command, print func(any) error, mgr *serviceLike) error {
					result, err := action(ctx, mgr, args[0])
					if err != nil {
						return err
					}
					return print(result)
				})
			},
		})
	}

	addNamedVMAction("start", "Start a VM", func(ctx context.Context, mgr *serviceLike, name string) (any, error) {
		return mgr.StartVM(ctx, name)
	})
	addNamedVMAction("shutdown", "Gracefully shut down a VM", func(ctx context.Context, mgr *serviceLike, name string) (any, error) {
		return mgr.ShutdownVM(ctx, name)
	})
	addNamedVMAction("destroy", "Force stop a VM", func(ctx context.Context, mgr *serviceLike, name string) (any, error) {
		return mgr.DestroyVM(ctx, name)
	})
	addNamedVMAction("reboot", "Reboot a VM", func(ctx context.Context, mgr *serviceLike, name string) (any, error) {
		return mgr.RebootVM(ctx, name)
	})
	addNamedVMAction("pause", "Pause a VM", func(ctx context.Context, mgr *serviceLike, name string) (any, error) {
		return mgr.PauseVM(ctx, name)
	})
	addNamedVMAction("resume", "Resume a VM", func(ctx context.Context, mgr *serviceLike, name string) (any, error) {
		return mgr.ResumeVM(ctx, name)
	})
	addNamedVMAction("undefine", "Remove a VM definition", func(ctx context.Context, mgr *serviceLike, name string) (any, error) {
		return mgr.UndefineVM(ctx, name)
	})
	addNamedVMAction("graphics", "Show graphics connection info", func(ctx context.Context, mgr *serviceLike, name string) (any, error) {
		return mgr.GetVMGraphics(ctx, name)
	})

	nicCmd := &cobra.Command{Use: "nic", Short: "Manage VM network interfaces"}
	nicAdd := &cobra.Command{
		Use:   "add <vm>",
		Short: "Add a VM network interface",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			spec := model.VMNICSpec{
				VM:      args[0],
				Alias:   mustGetString(cmd, "alias"),
				Network: mustGetString(cmd, "network"),
				Bridge:  mustGetString(cmd, "bridge"),
				MAC:     mustGetString(cmd, "mac"),
				Model:   mustGetString(cmd, "model"),
			}
			return runWithManager(cmd, opts, deps, func(ctx context.Context, _ *cobra.Command, print func(any) error, mgr *serviceLike) error {
				vm, err := mgr.AddVMNIC(ctx, spec)
				if err != nil {
					return err
				}
				return print(vm)
			})
		},
	}
	nicAdd.Flags().String("alias", "", "device alias")
	nicAdd.Flags().String("network", "", "libvirt network")
	nicAdd.Flags().String("bridge", "", "bridge name")
	nicAdd.Flags().String("mac", "", "MAC address")
	nicAdd.Flags().String("model", "virtio", "NIC model")
	nicCmd.AddCommand(nicAdd)

	nicUpdate := &cobra.Command{
		Use:   "update <vm>",
		Short: "Update a VM network interface",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			spec := model.VMNICUpdateSpec{
				VMNICSpec: model.VMNICSpec{
					VM:      args[0],
					Alias:   mustGetString(cmd, "alias"),
					Network: mustGetString(cmd, "network"),
					Bridge:  mustGetString(cmd, "bridge"),
					MAC:     mustGetString(cmd, "mac"),
					Model:   mustGetString(cmd, "model"),
				},
			}
			return runWithManager(cmd, opts, deps, func(ctx context.Context, _ *cobra.Command, print func(any) error, mgr *serviceLike) error {
				vm, err := mgr.UpdateVMNIC(ctx, spec)
				if err != nil {
					return err
				}
				return print(vm)
			})
		},
	}
	nicUpdate.Flags().String("alias", "", "device alias")
	nicUpdate.Flags().String("network", "", "libvirt network")
	nicUpdate.Flags().String("bridge", "", "bridge name")
	nicUpdate.Flags().String("mac", "", "MAC address")
	nicUpdate.Flags().String("model", "virtio", "NIC model")
	nicCmd.AddCommand(nicUpdate)

	nicRemove := &cobra.Command{
		Use:   "remove <vm>",
		Short: "Remove a VM network interface",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			alias := mustGetString(cmd, "alias")
			return runWithManager(cmd, opts, deps, func(ctx context.Context, _ *cobra.Command, print func(any) error, mgr *serviceLike) error {
				vm, err := mgr.RemoveVMNIC(ctx, args[0], alias)
				if err != nil {
					return err
				}
				return print(vm)
			})
		},
	}
	nicRemove.Flags().String("alias", "", "device alias")
	nicCmd.AddCommand(nicRemove)
	cmd.AddCommand(nicCmd)

	return cmd
}
