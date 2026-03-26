package service

import (
	"context"
	"strings"

	"github.com/ankele/pvm/internal/backend"
	"github.com/ankele/pvm/internal/model"
)

type Manager struct {
	backend backend.Backend
}

func NewManager(b backend.Backend) *Manager {
	return &Manager{backend: b}
}

func (m *Manager) Close() error {
	if m == nil || m.backend == nil {
		return nil
	}
	return m.backend.Close()
}

func (m *Manager) Info(ctx context.Context) (*model.SystemView, error) {
	return m.backend.Info(ctx)
}

func (m *Manager) ListVMs(ctx context.Context) ([]model.VMView, error) {
	return m.backend.ListVMs(ctx)
}

func (m *Manager) GetVM(ctx context.Context, name string) (*model.VMView, error) {
	if strings.TrimSpace(name) == "" {
		return nil, model.Errorf(model.ErrInvalidArgument, "vm name is required")
	}
	return m.backend.GetVM(ctx, name)
}

func (m *Manager) DefineVM(ctx context.Context, spec model.DomainSpec) (*model.VMView, error) {
	if err := validateDomainSpec(spec); err != nil {
		return nil, err
	}
	return m.backend.DefineVM(ctx, spec)
}

func (m *Manager) InstallVM(ctx context.Context, spec model.VMInstallSpec) (*model.VMView, error) {
	if strings.TrimSpace(spec.Name) == "" {
		return nil, model.Errorf(model.ErrInvalidArgument, "vm name is required")
	}
	if strings.TrimSpace(spec.Pool) == "" || strings.TrimSpace(spec.DiskName) == "" {
		return nil, model.Errorf(model.ErrInvalidArgument, "pool and disk name are required")
	}
	if strings.TrimSpace(spec.ISOPath) == "" {
		return nil, model.Errorf(model.ErrInvalidArgument, "iso path is required")
	}
	if spec.MemoryMiB == 0 || spec.VCPU == 0 || spec.DiskSizeGiB == 0 {
		return nil, model.Errorf(model.ErrInvalidArgument, "memory, vcpu and disk size must be greater than zero")
	}
	return m.backend.InstallVM(ctx, spec)
}

func (m *Manager) LaunchVM(ctx context.Context, spec model.VMLaunchSpec) (*model.VMView, error) {
	if strings.TrimSpace(spec.Name) == "" {
		return nil, model.Errorf(model.ErrInvalidArgument, "vm name is required")
	}
	if spec.Mode != model.LaunchModeClone && spec.Mode != model.LaunchModeDirect {
		return nil, model.Errorf(model.ErrInvalidArgument, "launch mode must be clone or direct")
	}
	if strings.TrimSpace(spec.ImagePath) == "" && strings.TrimSpace(spec.SourceVolume) == "" {
		return nil, model.Errorf(model.ErrInvalidArgument, "either image path or source volume is required")
	}
	if strings.TrimSpace(spec.ImagePath) != "" && strings.TrimSpace(spec.SourceVolume) != "" {
		return nil, model.Errorf(model.ErrInvalidArgument, "image path and source volume are mutually exclusive")
	}
	if spec.Mode == model.LaunchModeClone {
		if strings.TrimSpace(spec.TargetPool) == "" || strings.TrimSpace(spec.TargetVolumeName) == "" {
			return nil, model.Errorf(model.ErrInvalidArgument, "clone mode requires target pool and target volume name")
		}
	}
	if spec.MemoryMiB == 0 || spec.VCPU == 0 {
		return nil, model.Errorf(model.ErrInvalidArgument, "memory and vcpu must be greater than zero")
	}
	return m.backend.LaunchVM(ctx, spec)
}

func (m *Manager) StartVM(ctx context.Context, name string) (*model.VMView, error) {
	return m.simpleVMAction(ctx, name, m.backend.StartVM)
}
func (m *Manager) ShutdownVM(ctx context.Context, name string) (*model.VMView, error) {
	return m.simpleVMAction(ctx, name, m.backend.ShutdownVM)
}
func (m *Manager) DestroyVM(ctx context.Context, name string) (*model.VMView, error) {
	return m.simpleVMAction(ctx, name, m.backend.DestroyVM)
}
func (m *Manager) RebootVM(ctx context.Context, name string) (*model.VMView, error) {
	return m.simpleVMAction(ctx, name, m.backend.RebootVM)
}
func (m *Manager) PauseVM(ctx context.Context, name string) (*model.VMView, error) {
	return m.simpleVMAction(ctx, name, m.backend.PauseVM)
}
func (m *Manager) ResumeVM(ctx context.Context, name string) (*model.VMView, error) {
	return m.simpleVMAction(ctx, name, m.backend.ResumeVM)
}
func (m *Manager) UndefineVM(ctx context.Context, name string) (*model.ActionResult, error) {
	return m.simpleVMDelete(ctx, name, m.backend.UndefineVM)
}
func (m *Manager) GetVMGraphics(ctx context.Context, name string) (*model.GraphicsInfo, error) {
	if strings.TrimSpace(name) == "" {
		return nil, model.Errorf(model.ErrInvalidArgument, "vm name is required")
	}
	return m.backend.GetVMGraphics(ctx, name)
}

func (m *Manager) AddVMNIC(ctx context.Context, spec model.VMNICSpec) (*model.VMView, error) {
	if strings.TrimSpace(spec.VM) == "" {
		return nil, model.Errorf(model.ErrInvalidArgument, "vm name is required")
	}
	if strings.TrimSpace(spec.Network) == "" && strings.TrimSpace(spec.Bridge) == "" {
		return nil, model.Errorf(model.ErrInvalidArgument, "network or bridge is required")
	}
	return m.backend.AddVMNIC(ctx, spec)
}

func (m *Manager) UpdateVMNIC(ctx context.Context, spec model.VMNICUpdateSpec) (*model.VMView, error) {
	if strings.TrimSpace(spec.VM) == "" || strings.TrimSpace(spec.Alias) == "" {
		return nil, model.Errorf(model.ErrInvalidArgument, "vm and alias are required")
	}
	return m.backend.UpdateVMNIC(ctx, spec)
}

func (m *Manager) RemoveVMNIC(ctx context.Context, vmName, alias string) (*model.VMView, error) {
	if strings.TrimSpace(vmName) == "" || strings.TrimSpace(alias) == "" {
		return nil, model.Errorf(model.ErrInvalidArgument, "vm and alias are required")
	}
	return m.backend.RemoveVMNIC(ctx, vmName, alias)
}

func (m *Manager) CreateDomainSnapshot(ctx context.Context, spec model.VMSnapshotCreateSpec) (*model.SnapshotView, error) {
	if strings.TrimSpace(spec.VM) == "" || strings.TrimSpace(spec.Name) == "" {
		return nil, model.Errorf(model.ErrInvalidArgument, "vm and snapshot name are required")
	}
	return m.backend.CreateDomainSnapshot(ctx, spec)
}

func (m *Manager) ListDomainSnapshots(ctx context.Context, vm string) ([]model.SnapshotView, error) {
	if strings.TrimSpace(vm) == "" {
		return nil, model.Errorf(model.ErrInvalidArgument, "vm name is required")
	}
	return m.backend.ListDomainSnapshots(ctx, vm)
}

func (m *Manager) DeleteDomainSnapshot(ctx context.Context, vm, name string) (*model.ActionResult, error) {
	if strings.TrimSpace(vm) == "" || strings.TrimSpace(name) == "" {
		return nil, model.Errorf(model.ErrInvalidArgument, "vm and snapshot name are required")
	}
	return m.backend.DeleteDomainSnapshot(ctx, vm, name)
}

func (m *Manager) RevertDomainSnapshot(ctx context.Context, vm, name string) (*model.SnapshotView, error) {
	if strings.TrimSpace(vm) == "" || strings.TrimSpace(name) == "" {
		return nil, model.Errorf(model.ErrInvalidArgument, "vm and snapshot name are required")
	}
	return m.backend.RevertDomainSnapshot(ctx, vm, name)
}

func (m *Manager) ListPools(ctx context.Context) ([]model.StoragePoolView, error) {
	return m.backend.ListPools(ctx)
}
func (m *Manager) ListNetworks(ctx context.Context) ([]model.NetworkView, error) {
	return m.backend.ListNetworks(ctx)
}
func (m *Manager) ListInterfaces(ctx context.Context) ([]model.InterfaceView, error) {
	return m.backend.ListInterfaces(ctx)
}
func (m *Manager) GetPool(ctx context.Context, name string) (*model.StoragePoolView, error) {
	return m.namedPool(ctx, name, m.backend.GetPool)
}
func (m *Manager) StartPool(ctx context.Context, name string) (*model.StoragePoolView, error) {
	return m.namedPool(ctx, name, m.backend.StartPool)
}
func (m *Manager) DestroyPool(ctx context.Context, name string) (*model.StoragePoolView, error) {
	return m.namedPool(ctx, name, m.backend.DestroyPool)
}
func (m *Manager) RefreshPool(ctx context.Context, name string) (*model.StoragePoolView, error) {
	return m.namedPool(ctx, name, m.backend.RefreshPool)
}
func (m *Manager) UndefinePool(ctx context.Context, name string) (*model.ActionResult, error) {
	return m.namedPoolDelete(ctx, name, m.backend.UndefinePool)
}
func (m *Manager) GetNetwork(ctx context.Context, name string) (*model.NetworkView, error) {
	return m.namedNetwork(ctx, name, m.backend.GetNetwork)
}
func (m *Manager) StartNetwork(ctx context.Context, name string) (*model.NetworkView, error) {
	return m.namedNetwork(ctx, name, m.backend.StartNetwork)
}
func (m *Manager) DestroyNetwork(ctx context.Context, name string) (*model.NetworkView, error) {
	return m.namedNetwork(ctx, name, m.backend.DestroyNetwork)
}
func (m *Manager) UndefineNetwork(ctx context.Context, name string) (*model.ActionResult, error) {
	return m.namedNetworkDelete(ctx, name, m.backend.UndefineNetwork)
}
func (m *Manager) GetInterface(ctx context.Context, name string) (*model.InterfaceView, error) {
	return m.namedInterface(ctx, name, m.backend.GetInterface)
}
func (m *Manager) StartInterface(ctx context.Context, name string) (*model.InterfaceView, error) {
	return m.namedInterface(ctx, name, m.backend.StartInterface)
}
func (m *Manager) DestroyInterface(ctx context.Context, name string) (*model.InterfaceView, error) {
	return m.namedInterface(ctx, name, m.backend.DestroyInterface)
}
func (m *Manager) UndefineInterface(ctx context.Context, name string) (*model.ActionResult, error) {
	return m.namedInterfaceDelete(ctx, name, m.backend.UndefineInterface)
}
func (m *Manager) SetPoolAutostart(ctx context.Context, name string, autostart bool) (*model.StoragePoolView, error) {
	if strings.TrimSpace(name) == "" {
		return nil, model.Errorf(model.ErrInvalidArgument, "pool name is required")
	}
	return m.backend.SetPoolAutostart(ctx, name, autostart)
}
func (m *Manager) SetNetworkAutostart(ctx context.Context, name string, autostart bool) (*model.NetworkView, error) {
	if strings.TrimSpace(name) == "" {
		return nil, model.Errorf(model.ErrInvalidArgument, "network name is required")
	}
	return m.backend.SetNetworkAutostart(ctx, name, autostart)
}

func (m *Manager) DefinePool(ctx context.Context, spec model.PoolDefineSpec) (*model.StoragePoolView, error) {
	if strings.TrimSpace(spec.Name) == "" || strings.TrimSpace(spec.Type) == "" {
		return nil, model.Errorf(model.ErrInvalidArgument, "pool name and type are required")
	}
	return m.backend.DefinePool(ctx, spec)
}

func (m *Manager) ListVolumes(ctx context.Context, pool string) ([]model.VolumeView, error) {
	if strings.TrimSpace(pool) == "" {
		return nil, model.Errorf(model.ErrInvalidArgument, "pool name is required")
	}
	return m.backend.ListVolumes(ctx, pool)
}

func (m *Manager) GetVolume(ctx context.Context, pool, name string) (*model.VolumeView, error) {
	if strings.TrimSpace(pool) == "" || strings.TrimSpace(name) == "" {
		return nil, model.Errorf(model.ErrInvalidArgument, "pool and volume are required")
	}
	return m.backend.GetVolume(ctx, pool, name)
}

func (m *Manager) CreateVolume(ctx context.Context, spec model.VolumeCreateSpec) (*model.VolumeView, error) {
	if strings.TrimSpace(spec.Pool) == "" || strings.TrimSpace(spec.Name) == "" || spec.CapacityBytes == 0 {
		return nil, model.Errorf(model.ErrInvalidArgument, "pool, volume name and capacity are required")
	}
	return m.backend.CreateVolume(ctx, spec)
}

func (m *Manager) DeleteVolume(ctx context.Context, pool, name string) (*model.ActionResult, error) {
	if strings.TrimSpace(pool) == "" || strings.TrimSpace(name) == "" {
		return nil, model.Errorf(model.ErrInvalidArgument, "pool and volume are required")
	}
	return m.backend.DeleteVolume(ctx, pool, name)
}

func (m *Manager) ResizeVolume(ctx context.Context, spec model.VolumeResizeSpec) (*model.VolumeView, error) {
	if strings.TrimSpace(spec.Pool) == "" || strings.TrimSpace(spec.Name) == "" || spec.CapacityBytes == 0 {
		return nil, model.Errorf(model.ErrInvalidArgument, "pool, volume and capacity are required")
	}
	return m.backend.ResizeVolume(ctx, spec)
}

func (m *Manager) CreateVolumeSnapshot(ctx context.Context, spec model.VolumeSnapshotCreateSpec) (*model.SnapshotView, error) {
	if strings.TrimSpace(spec.Pool) == "" || strings.TrimSpace(spec.Volume) == "" || strings.TrimSpace(spec.Name) == "" {
		return nil, model.Errorf(model.ErrInvalidArgument, "pool, volume and snapshot name are required")
	}
	return m.backend.CreateVolumeSnapshot(ctx, spec)
}

func (m *Manager) ListVolumeSnapshots(ctx context.Context, pool, volume string) ([]model.SnapshotView, error) {
	if strings.TrimSpace(pool) == "" || strings.TrimSpace(volume) == "" {
		return nil, model.Errorf(model.ErrInvalidArgument, "pool and volume are required")
	}
	return m.backend.ListVolumeSnapshots(ctx, pool, volume)
}

func (m *Manager) DeleteVolumeSnapshot(ctx context.Context, pool, volume, name string) (*model.ActionResult, error) {
	if strings.TrimSpace(pool) == "" || strings.TrimSpace(volume) == "" || strings.TrimSpace(name) == "" {
		return nil, model.Errorf(model.ErrInvalidArgument, "pool, volume and snapshot name are required")
	}
	return m.backend.DeleteVolumeSnapshot(ctx, pool, volume, name)
}

func (m *Manager) RollbackVolumeSnapshot(ctx context.Context, pool, volume, name string) (*model.SnapshotView, error) {
	if strings.TrimSpace(pool) == "" || strings.TrimSpace(volume) == "" || strings.TrimSpace(name) == "" {
		return nil, model.Errorf(model.ErrInvalidArgument, "pool, volume and snapshot name are required")
	}
	return m.backend.RollbackVolumeSnapshot(ctx, pool, volume, name)
}

func (m *Manager) DefineNetwork(ctx context.Context, spec model.NetworkDefineSpec) (*model.NetworkView, error) {
	if strings.TrimSpace(spec.Name) == "" {
		return nil, model.Errorf(model.ErrInvalidArgument, "network name is required")
	}
	return m.backend.DefineNetwork(ctx, spec)
}

func (m *Manager) DefineInterface(ctx context.Context, spec model.InterfaceDefineSpec) (*model.InterfaceView, error) {
	if strings.TrimSpace(spec.Name) == "" || strings.TrimSpace(spec.Type) == "" {
		return nil, model.Errorf(model.ErrInvalidArgument, "interface name and type are required")
	}
	return m.backend.DefineInterface(ctx, spec)
}

func validateDomainSpec(spec model.DomainSpec) error {
	if strings.TrimSpace(spec.Name) == "" {
		return model.Errorf(model.ErrInvalidArgument, "vm name is required")
	}
	if spec.MemoryMiB == 0 || spec.VCPU == 0 {
		return model.Errorf(model.ErrInvalidArgument, "memory and vcpu must be greater than zero")
	}
	if len(spec.Disks) == 0 {
		return model.Errorf(model.ErrInvalidArgument, "at least one disk is required")
	}
	return nil
}

func (m *Manager) simpleVMAction(ctx context.Context, name string, fn func(context.Context, string) (*model.VMView, error)) (*model.VMView, error) {
	if strings.TrimSpace(name) == "" {
		return nil, model.Errorf(model.ErrInvalidArgument, "vm name is required")
	}
	return fn(ctx, name)
}

func (m *Manager) simpleVMDelete(ctx context.Context, name string, fn func(context.Context, string) (*model.ActionResult, error)) (*model.ActionResult, error) {
	if strings.TrimSpace(name) == "" {
		return nil, model.Errorf(model.ErrInvalidArgument, "vm name is required")
	}
	return fn(ctx, name)
}

func (m *Manager) namedPool(ctx context.Context, name string, fn func(context.Context, string) (*model.StoragePoolView, error)) (*model.StoragePoolView, error) {
	if strings.TrimSpace(name) == "" {
		return nil, model.Errorf(model.ErrInvalidArgument, "pool name is required")
	}
	return fn(ctx, name)
}

func (m *Manager) namedPoolDelete(ctx context.Context, name string, fn func(context.Context, string) (*model.ActionResult, error)) (*model.ActionResult, error) {
	if strings.TrimSpace(name) == "" {
		return nil, model.Errorf(model.ErrInvalidArgument, "pool name is required")
	}
	return fn(ctx, name)
}

func (m *Manager) namedNetwork(ctx context.Context, name string, fn func(context.Context, string) (*model.NetworkView, error)) (*model.NetworkView, error) {
	if strings.TrimSpace(name) == "" {
		return nil, model.Errorf(model.ErrInvalidArgument, "network name is required")
	}
	return fn(ctx, name)
}

func (m *Manager) namedNetworkDelete(ctx context.Context, name string, fn func(context.Context, string) (*model.ActionResult, error)) (*model.ActionResult, error) {
	if strings.TrimSpace(name) == "" {
		return nil, model.Errorf(model.ErrInvalidArgument, "network name is required")
	}
	return fn(ctx, name)
}

func (m *Manager) namedInterface(ctx context.Context, name string, fn func(context.Context, string) (*model.InterfaceView, error)) (*model.InterfaceView, error) {
	if strings.TrimSpace(name) == "" {
		return nil, model.Errorf(model.ErrInvalidArgument, "interface name is required")
	}
	return fn(ctx, name)
}

func (m *Manager) namedInterfaceDelete(ctx context.Context, name string, fn func(context.Context, string) (*model.ActionResult, error)) (*model.ActionResult, error) {
	if strings.TrimSpace(name) == "" {
		return nil, model.Errorf(model.ErrInvalidArgument, "interface name is required")
	}
	return fn(ctx, name)
}
