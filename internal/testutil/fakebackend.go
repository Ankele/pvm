package testutil

import (
	"context"

	"github.com/ankele/pvm/internal/backend"
	"github.com/ankele/pvm/internal/model"
)

type FakeBackend struct {
	InfoFn                  func(context.Context) (*model.SystemView, error)
	ListVMsFn               func(context.Context) ([]model.VMView, error)
	GetVMFn                 func(context.Context, string) (*model.VMView, error)
	DefineVMFn              func(context.Context, model.DomainSpec) (*model.VMView, error)
	InstallVMFn             func(context.Context, model.VMInstallSpec) (*model.VMView, error)
	LaunchVMFn              func(context.Context, model.VMLaunchSpec) (*model.VMView, error)
	StartVMFn               func(context.Context, string) (*model.VMView, error)
	ShutdownVMFn            func(context.Context, string) (*model.VMView, error)
	DestroyVMFn             func(context.Context, string) (*model.VMView, error)
	RebootVMFn              func(context.Context, string) (*model.VMView, error)
	PauseVMFn               func(context.Context, string) (*model.VMView, error)
	ResumeVMFn              func(context.Context, string) (*model.VMView, error)
	UndefineVMFn            func(context.Context, string) (*model.ActionResult, error)
	AddVMNICFn              func(context.Context, model.VMNICSpec) (*model.VMView, error)
	UpdateVMNICFn           func(context.Context, model.VMNICUpdateSpec) (*model.VMView, error)
	RemoveVMNICFn           func(context.Context, string, string) (*model.VMView, error)
	GetVMGraphicsFn         func(context.Context, string) (*model.GraphicsInfo, error)
	CreateDomainSnapshotFn  func(context.Context, model.VMSnapshotCreateSpec) (*model.SnapshotView, error)
	ListDomainSnapshotsFn   func(context.Context, string) ([]model.SnapshotView, error)
	DeleteDomainSnapshotFn  func(context.Context, string, string) (*model.ActionResult, error)
	RevertDomainSnapshotFn  func(context.Context, string, string) (*model.SnapshotView, error)
	ListPoolsFn             func(context.Context) ([]model.StoragePoolView, error)
	GetPoolFn               func(context.Context, string) (*model.StoragePoolView, error)
	DefinePoolFn            func(context.Context, model.PoolDefineSpec) (*model.StoragePoolView, error)
	StartPoolFn             func(context.Context, string) (*model.StoragePoolView, error)
	DestroyPoolFn           func(context.Context, string) (*model.StoragePoolView, error)
	SetPoolAutostartFn      func(context.Context, string, bool) (*model.StoragePoolView, error)
	UndefinePoolFn          func(context.Context, string) (*model.ActionResult, error)
	RefreshPoolFn           func(context.Context, string) (*model.StoragePoolView, error)
	ListVolumesFn           func(context.Context, string) ([]model.VolumeView, error)
	GetVolumeFn             func(context.Context, string, string) (*model.VolumeView, error)
	CreateVolumeFn          func(context.Context, model.VolumeCreateSpec) (*model.VolumeView, error)
	DeleteVolumeFn          func(context.Context, string, string) (*model.ActionResult, error)
	ResizeVolumeFn          func(context.Context, model.VolumeResizeSpec) (*model.VolumeView, error)
	CreateVolumeSnapshotFn  func(context.Context, model.VolumeSnapshotCreateSpec) (*model.SnapshotView, error)
	ListVolumeSnapshotsFn   func(context.Context, string, string) ([]model.SnapshotView, error)
	DeleteVolumeSnapshotFn  func(context.Context, string, string, string) (*model.ActionResult, error)
	RollbackVolumeSnapshotFn func(context.Context, string, string, string) (*model.SnapshotView, error)
	ListNetworksFn          func(context.Context) ([]model.NetworkView, error)
	GetNetworkFn            func(context.Context, string) (*model.NetworkView, error)
	DefineNetworkFn         func(context.Context, model.NetworkDefineSpec) (*model.NetworkView, error)
	StartNetworkFn          func(context.Context, string) (*model.NetworkView, error)
	DestroyNetworkFn        func(context.Context, string) (*model.NetworkView, error)
	SetNetworkAutostartFn   func(context.Context, string, bool) (*model.NetworkView, error)
	UndefineNetworkFn       func(context.Context, string) (*model.ActionResult, error)
	ListInterfacesFn        func(context.Context) ([]model.InterfaceView, error)
	GetInterfaceFn          func(context.Context, string) (*model.InterfaceView, error)
	DefineInterfaceFn       func(context.Context, model.InterfaceDefineSpec) (*model.InterfaceView, error)
	StartInterfaceFn        func(context.Context, string) (*model.InterfaceView, error)
	DestroyInterfaceFn      func(context.Context, string) (*model.InterfaceView, error)
	UndefineInterfaceFn     func(context.Context, string) (*model.ActionResult, error)
}

func (f *FakeBackend) Close() error { return nil }

func (f *FakeBackend) Info(ctx context.Context) (*model.SystemView, error) {
	if f.InfoFn != nil {
		return f.InfoFn(ctx)
	}
	return &model.SystemView{Backend: "fake", URI: "test:///default", SupportsLibvirt: true}, nil
}

func (f *FakeBackend) ListVMs(ctx context.Context) ([]model.VMView, error) {
	if f.ListVMsFn != nil {
		return f.ListVMsFn(ctx)
	}
	return nil, nil
}

func (f *FakeBackend) GetVM(ctx context.Context, name string) (*model.VMView, error) {
	if f.GetVMFn != nil {
		return f.GetVMFn(ctx, name)
	}
	return nil, nil
}

func (f *FakeBackend) DefineVM(ctx context.Context, spec model.DomainSpec) (*model.VMView, error) {
	if f.DefineVMFn != nil {
		return f.DefineVMFn(ctx, spec)
	}
	return nil, nil
}

func (f *FakeBackend) InstallVM(ctx context.Context, spec model.VMInstallSpec) (*model.VMView, error) {
	if f.InstallVMFn != nil {
		return f.InstallVMFn(ctx, spec)
	}
	return nil, nil
}

func (f *FakeBackend) LaunchVM(ctx context.Context, spec model.VMLaunchSpec) (*model.VMView, error) {
	if f.LaunchVMFn != nil {
		return f.LaunchVMFn(ctx, spec)
	}
	return nil, nil
}

func (f *FakeBackend) StartVM(ctx context.Context, name string) (*model.VMView, error) {
	if f.StartVMFn != nil {
		return f.StartVMFn(ctx, name)
	}
	return nil, nil
}

func (f *FakeBackend) ShutdownVM(ctx context.Context, name string) (*model.VMView, error) {
	if f.ShutdownVMFn != nil {
		return f.ShutdownVMFn(ctx, name)
	}
	return nil, nil
}

func (f *FakeBackend) DestroyVM(ctx context.Context, name string) (*model.VMView, error) {
	if f.DestroyVMFn != nil {
		return f.DestroyVMFn(ctx, name)
	}
	return nil, nil
}

func (f *FakeBackend) RebootVM(ctx context.Context, name string) (*model.VMView, error) {
	if f.RebootVMFn != nil {
		return f.RebootVMFn(ctx, name)
	}
	return nil, nil
}

func (f *FakeBackend) PauseVM(ctx context.Context, name string) (*model.VMView, error) {
	if f.PauseVMFn != nil {
		return f.PauseVMFn(ctx, name)
	}
	return nil, nil
}

func (f *FakeBackend) ResumeVM(ctx context.Context, name string) (*model.VMView, error) {
	if f.ResumeVMFn != nil {
		return f.ResumeVMFn(ctx, name)
	}
	return nil, nil
}

func (f *FakeBackend) UndefineVM(ctx context.Context, name string) (*model.ActionResult, error) {
	if f.UndefineVMFn != nil {
		return f.UndefineVMFn(ctx, name)
	}
	return &model.ActionResult{Message: "ok"}, nil
}

func (f *FakeBackend) AddVMNIC(ctx context.Context, spec model.VMNICSpec) (*model.VMView, error) {
	if f.AddVMNICFn != nil {
		return f.AddVMNICFn(ctx, spec)
	}
	return nil, nil
}

func (f *FakeBackend) UpdateVMNIC(ctx context.Context, spec model.VMNICUpdateSpec) (*model.VMView, error) {
	if f.UpdateVMNICFn != nil {
		return f.UpdateVMNICFn(ctx, spec)
	}
	return nil, nil
}

func (f *FakeBackend) RemoveVMNIC(ctx context.Context, vm, alias string) (*model.VMView, error) {
	if f.RemoveVMNICFn != nil {
		return f.RemoveVMNICFn(ctx, vm, alias)
	}
	return nil, nil
}

func (f *FakeBackend) GetVMGraphics(ctx context.Context, vm string) (*model.GraphicsInfo, error) {
	if f.GetVMGraphicsFn != nil {
		return f.GetVMGraphicsFn(ctx, vm)
	}
	return nil, nil
}

func (f *FakeBackend) CreateDomainSnapshot(ctx context.Context, spec model.VMSnapshotCreateSpec) (*model.SnapshotView, error) {
	if f.CreateDomainSnapshotFn != nil {
		return f.CreateDomainSnapshotFn(ctx, spec)
	}
	return nil, nil
}

func (f *FakeBackend) ListDomainSnapshots(ctx context.Context, vm string) ([]model.SnapshotView, error) {
	if f.ListDomainSnapshotsFn != nil {
		return f.ListDomainSnapshotsFn(ctx, vm)
	}
	return nil, nil
}

func (f *FakeBackend) DeleteDomainSnapshot(ctx context.Context, vm, snapshot string) (*model.ActionResult, error) {
	if f.DeleteDomainSnapshotFn != nil {
		return f.DeleteDomainSnapshotFn(ctx, vm, snapshot)
	}
	return &model.ActionResult{Message: "ok"}, nil
}

func (f *FakeBackend) RevertDomainSnapshot(ctx context.Context, vm, snapshot string) (*model.SnapshotView, error) {
	if f.RevertDomainSnapshotFn != nil {
		return f.RevertDomainSnapshotFn(ctx, vm, snapshot)
	}
	return nil, nil
}

func (f *FakeBackend) ListPools(ctx context.Context) ([]model.StoragePoolView, error) {
	if f.ListPoolsFn != nil {
		return f.ListPoolsFn(ctx)
	}
	return nil, nil
}

func (f *FakeBackend) GetPool(ctx context.Context, name string) (*model.StoragePoolView, error) {
	if f.GetPoolFn != nil {
		return f.GetPoolFn(ctx, name)
	}
	return nil, nil
}

func (f *FakeBackend) DefinePool(ctx context.Context, spec model.PoolDefineSpec) (*model.StoragePoolView, error) {
	if f.DefinePoolFn != nil {
		return f.DefinePoolFn(ctx, spec)
	}
	return nil, nil
}

func (f *FakeBackend) StartPool(ctx context.Context, name string) (*model.StoragePoolView, error) {
	if f.StartPoolFn != nil {
		return f.StartPoolFn(ctx, name)
	}
	return nil, nil
}

func (f *FakeBackend) DestroyPool(ctx context.Context, name string) (*model.StoragePoolView, error) {
	if f.DestroyPoolFn != nil {
		return f.DestroyPoolFn(ctx, name)
	}
	return nil, nil
}

func (f *FakeBackend) SetPoolAutostart(ctx context.Context, name string, enabled bool) (*model.StoragePoolView, error) {
	if f.SetPoolAutostartFn != nil {
		return f.SetPoolAutostartFn(ctx, name, enabled)
	}
	return nil, nil
}

func (f *FakeBackend) UndefinePool(ctx context.Context, name string) (*model.ActionResult, error) {
	if f.UndefinePoolFn != nil {
		return f.UndefinePoolFn(ctx, name)
	}
	return &model.ActionResult{Message: "ok"}, nil
}

func (f *FakeBackend) RefreshPool(ctx context.Context, name string) (*model.StoragePoolView, error) {
	if f.RefreshPoolFn != nil {
		return f.RefreshPoolFn(ctx, name)
	}
	return nil, nil
}

func (f *FakeBackend) ListVolumes(ctx context.Context, pool string) ([]model.VolumeView, error) {
	if f.ListVolumesFn != nil {
		return f.ListVolumesFn(ctx, pool)
	}
	return nil, nil
}

func (f *FakeBackend) GetVolume(ctx context.Context, pool, name string) (*model.VolumeView, error) {
	if f.GetVolumeFn != nil {
		return f.GetVolumeFn(ctx, pool, name)
	}
	return nil, nil
}

func (f *FakeBackend) CreateVolume(ctx context.Context, spec model.VolumeCreateSpec) (*model.VolumeView, error) {
	if f.CreateVolumeFn != nil {
		return f.CreateVolumeFn(ctx, spec)
	}
	return nil, nil
}

func (f *FakeBackend) DeleteVolume(ctx context.Context, pool, name string) (*model.ActionResult, error) {
	if f.DeleteVolumeFn != nil {
		return f.DeleteVolumeFn(ctx, pool, name)
	}
	return &model.ActionResult{Message: "ok"}, nil
}

func (f *FakeBackend) ResizeVolume(ctx context.Context, spec model.VolumeResizeSpec) (*model.VolumeView, error) {
	if f.ResizeVolumeFn != nil {
		return f.ResizeVolumeFn(ctx, spec)
	}
	return nil, nil
}

func (f *FakeBackend) CreateVolumeSnapshot(ctx context.Context, spec model.VolumeSnapshotCreateSpec) (*model.SnapshotView, error) {
	if f.CreateVolumeSnapshotFn != nil {
		return f.CreateVolumeSnapshotFn(ctx, spec)
	}
	return nil, nil
}

func (f *FakeBackend) ListVolumeSnapshots(ctx context.Context, pool, volume string) ([]model.SnapshotView, error) {
	if f.ListVolumeSnapshotsFn != nil {
		return f.ListVolumeSnapshotsFn(ctx, pool, volume)
	}
	return nil, nil
}

func (f *FakeBackend) DeleteVolumeSnapshot(ctx context.Context, pool, volume, snapshot string) (*model.ActionResult, error) {
	if f.DeleteVolumeSnapshotFn != nil {
		return f.DeleteVolumeSnapshotFn(ctx, pool, volume, snapshot)
	}
	return &model.ActionResult{Message: "ok"}, nil
}

func (f *FakeBackend) RollbackVolumeSnapshot(ctx context.Context, pool, volume, snapshot string) (*model.SnapshotView, error) {
	if f.RollbackVolumeSnapshotFn != nil {
		return f.RollbackVolumeSnapshotFn(ctx, pool, volume, snapshot)
	}
	return nil, nil
}

func (f *FakeBackend) ListNetworks(ctx context.Context) ([]model.NetworkView, error) {
	if f.ListNetworksFn != nil {
		return f.ListNetworksFn(ctx)
	}
	return nil, nil
}

func (f *FakeBackend) GetNetwork(ctx context.Context, name string) (*model.NetworkView, error) {
	if f.GetNetworkFn != nil {
		return f.GetNetworkFn(ctx, name)
	}
	return nil, nil
}

func (f *FakeBackend) DefineNetwork(ctx context.Context, spec model.NetworkDefineSpec) (*model.NetworkView, error) {
	if f.DefineNetworkFn != nil {
		return f.DefineNetworkFn(ctx, spec)
	}
	return nil, nil
}

func (f *FakeBackend) StartNetwork(ctx context.Context, name string) (*model.NetworkView, error) {
	if f.StartNetworkFn != nil {
		return f.StartNetworkFn(ctx, name)
	}
	return nil, nil
}

func (f *FakeBackend) DestroyNetwork(ctx context.Context, name string) (*model.NetworkView, error) {
	if f.DestroyNetworkFn != nil {
		return f.DestroyNetworkFn(ctx, name)
	}
	return nil, nil
}

func (f *FakeBackend) SetNetworkAutostart(ctx context.Context, name string, enabled bool) (*model.NetworkView, error) {
	if f.SetNetworkAutostartFn != nil {
		return f.SetNetworkAutostartFn(ctx, name, enabled)
	}
	return nil, nil
}

func (f *FakeBackend) UndefineNetwork(ctx context.Context, name string) (*model.ActionResult, error) {
	if f.UndefineNetworkFn != nil {
		return f.UndefineNetworkFn(ctx, name)
	}
	return &model.ActionResult{Message: "ok"}, nil
}

func (f *FakeBackend) ListInterfaces(ctx context.Context) ([]model.InterfaceView, error) {
	if f.ListInterfacesFn != nil {
		return f.ListInterfacesFn(ctx)
	}
	return nil, nil
}

func (f *FakeBackend) GetInterface(ctx context.Context, name string) (*model.InterfaceView, error) {
	if f.GetInterfaceFn != nil {
		return f.GetInterfaceFn(ctx, name)
	}
	return nil, nil
}

func (f *FakeBackend) DefineInterface(ctx context.Context, spec model.InterfaceDefineSpec) (*model.InterfaceView, error) {
	if f.DefineInterfaceFn != nil {
		return f.DefineInterfaceFn(ctx, spec)
	}
	return nil, nil
}

func (f *FakeBackend) StartInterface(ctx context.Context, name string) (*model.InterfaceView, error) {
	if f.StartInterfaceFn != nil {
		return f.StartInterfaceFn(ctx, name)
	}
	return nil, nil
}

func (f *FakeBackend) DestroyInterface(ctx context.Context, name string) (*model.InterfaceView, error) {
	if f.DestroyInterfaceFn != nil {
		return f.DestroyInterfaceFn(ctx, name)
	}
	return nil, nil
}

func (f *FakeBackend) UndefineInterface(ctx context.Context, name string) (*model.ActionResult, error) {
	if f.UndefineInterfaceFn != nil {
		return f.UndefineInterfaceFn(ctx, name)
	}
	return &model.ActionResult{Message: "ok"}, nil
}

func Factory(fake backend.Backend) backend.Factory {
	return func(context.Context, backend.Options) (backend.Backend, error) {
		return fake, nil
	}
}

