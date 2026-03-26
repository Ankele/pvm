package backend

import (
	"context"
	"runtime"

	"github.com/ankele/pvm/internal/model"
)

type unsupportedBackend struct {
	opts Options
}

func NewUnsupported(_ context.Context, opts Options) (Backend, error) {
	return &unsupportedBackend{opts: opts}, nil
}

func (b *unsupportedBackend) Close() error { return nil }

func (b *unsupportedBackend) unsupported() error {
	return model.Errorf(model.ErrUnsupported, "libvirt backend is only available on Linux with cgo; current platform is %s/%s", runtime.GOOS, runtime.GOARCH)
}

func (b *unsupportedBackend) Info(context.Context) (*model.SystemView, error) {
	return &model.SystemView{
		Backend:         "unsupported",
		URI:             b.opts.URI,
		Platform:        runtime.GOOS + "/" + runtime.GOARCH,
		SupportsLibvirt: false,
		SupportsZFS:     false,
	}, nil
}

func (b *unsupportedBackend) ListVMs(context.Context) ([]model.VMView, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) GetVM(context.Context, string) (*model.VMView, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) DefineVM(context.Context, model.DomainSpec) (*model.VMView, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) InstallVM(context.Context, model.VMInstallSpec) (*model.VMView, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) LaunchVM(context.Context, model.VMLaunchSpec) (*model.VMView, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) StartVM(context.Context, string) (*model.VMView, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) ShutdownVM(context.Context, string) (*model.VMView, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) DestroyVM(context.Context, string) (*model.VMView, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) RebootVM(context.Context, string) (*model.VMView, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) PauseVM(context.Context, string) (*model.VMView, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) ResumeVM(context.Context, string) (*model.VMView, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) UndefineVM(context.Context, string) (*model.ActionResult, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) AddVMNIC(context.Context, model.VMNICSpec) (*model.VMView, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) UpdateVMNIC(context.Context, model.VMNICUpdateSpec) (*model.VMView, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) RemoveVMNIC(context.Context, string, string) (*model.VMView, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) GetVMGraphics(context.Context, string) (*model.GraphicsInfo, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) CreateDomainSnapshot(context.Context, model.VMSnapshotCreateSpec) (*model.SnapshotView, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) ListDomainSnapshots(context.Context, string) ([]model.SnapshotView, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) DeleteDomainSnapshot(context.Context, string, string) (*model.ActionResult, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) RevertDomainSnapshot(context.Context, string, string) (*model.SnapshotView, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) ListPools(context.Context) ([]model.StoragePoolView, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) GetPool(context.Context, string) (*model.StoragePoolView, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) DefinePool(context.Context, model.PoolDefineSpec) (*model.StoragePoolView, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) StartPool(context.Context, string) (*model.StoragePoolView, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) DestroyPool(context.Context, string) (*model.StoragePoolView, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) SetPoolAutostart(context.Context, string, bool) (*model.StoragePoolView, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) UndefinePool(context.Context, string) (*model.ActionResult, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) RefreshPool(context.Context, string) (*model.StoragePoolView, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) ListVolumes(context.Context, string) ([]model.VolumeView, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) GetVolume(context.Context, string, string) (*model.VolumeView, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) CreateVolume(context.Context, model.VolumeCreateSpec) (*model.VolumeView, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) DeleteVolume(context.Context, string, string) (*model.ActionResult, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) ResizeVolume(context.Context, model.VolumeResizeSpec) (*model.VolumeView, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) CreateVolumeSnapshot(context.Context, model.VolumeSnapshotCreateSpec) (*model.SnapshotView, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) ListVolumeSnapshots(context.Context, string, string) ([]model.SnapshotView, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) DeleteVolumeSnapshot(context.Context, string, string, string) (*model.ActionResult, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) RollbackVolumeSnapshot(context.Context, string, string, string) (*model.SnapshotView, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) ListNetworks(context.Context) ([]model.NetworkView, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) GetNetwork(context.Context, string) (*model.NetworkView, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) DefineNetwork(context.Context, model.NetworkDefineSpec) (*model.NetworkView, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) StartNetwork(context.Context, string) (*model.NetworkView, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) DestroyNetwork(context.Context, string) (*model.NetworkView, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) SetNetworkAutostart(context.Context, string, bool) (*model.NetworkView, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) UndefineNetwork(context.Context, string) (*model.ActionResult, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) ListInterfaces(context.Context) ([]model.InterfaceView, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) GetInterface(context.Context, string) (*model.InterfaceView, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) DefineInterface(context.Context, model.InterfaceDefineSpec) (*model.InterfaceView, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) StartInterface(context.Context, string) (*model.InterfaceView, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) DestroyInterface(context.Context, string) (*model.InterfaceView, error) {
	return nil, b.unsupported()
}
func (b *unsupportedBackend) UndefineInterface(context.Context, string) (*model.ActionResult, error) {
	return nil, b.unsupported()
}
