package backend

import (
	"context"

	"github.com/ankele/pvm/internal/model"
)

type Options struct {
	URI      string
	Username string
	Password string
}

type Factory func(context.Context, Options) (Backend, error)

type Backend interface {
	Close() error
	Info(context.Context) (*model.SystemView, error)

	ListVMs(context.Context) ([]model.VMView, error)
	GetVM(context.Context, string) (*model.VMView, error)
	DefineVM(context.Context, model.DomainSpec) (*model.VMView, error)
	InstallVM(context.Context, model.VMInstallSpec) (*model.VMView, error)
	LaunchVM(context.Context, model.VMLaunchSpec) (*model.VMView, error)
	StartVM(context.Context, string) (*model.VMView, error)
	ShutdownVM(context.Context, string) (*model.VMView, error)
	DestroyVM(context.Context, string) (*model.VMView, error)
	RebootVM(context.Context, string) (*model.VMView, error)
	PauseVM(context.Context, string) (*model.VMView, error)
	ResumeVM(context.Context, string) (*model.VMView, error)
	UndefineVM(context.Context, string) (*model.ActionResult, error)
	AddVMNIC(context.Context, model.VMNICSpec) (*model.VMView, error)
	UpdateVMNIC(context.Context, model.VMNICUpdateSpec) (*model.VMView, error)
	RemoveVMNIC(context.Context, string, string) (*model.VMView, error)
	GetVMGraphics(context.Context, string) (*model.GraphicsInfo, error)

	CreateDomainSnapshot(context.Context, model.VMSnapshotCreateSpec) (*model.SnapshotView, error)
	ListDomainSnapshots(context.Context, string) ([]model.SnapshotView, error)
	DeleteDomainSnapshot(context.Context, string, string) (*model.ActionResult, error)
	RevertDomainSnapshot(context.Context, string, string) (*model.SnapshotView, error)

	ListPools(context.Context) ([]model.StoragePoolView, error)
	GetPool(context.Context, string) (*model.StoragePoolView, error)
	DefinePool(context.Context, model.PoolDefineSpec) (*model.StoragePoolView, error)
	StartPool(context.Context, string) (*model.StoragePoolView, error)
	DestroyPool(context.Context, string) (*model.StoragePoolView, error)
	SetPoolAutostart(context.Context, string, bool) (*model.StoragePoolView, error)
	UndefinePool(context.Context, string) (*model.ActionResult, error)
	RefreshPool(context.Context, string) (*model.StoragePoolView, error)

	ListVolumes(context.Context, string) ([]model.VolumeView, error)
	GetVolume(context.Context, string, string) (*model.VolumeView, error)
	CreateVolume(context.Context, model.VolumeCreateSpec) (*model.VolumeView, error)
	DeleteVolume(context.Context, string, string) (*model.ActionResult, error)
	ResizeVolume(context.Context, model.VolumeResizeSpec) (*model.VolumeView, error)

	CreateVolumeSnapshot(context.Context, model.VolumeSnapshotCreateSpec) (*model.SnapshotView, error)
	ListVolumeSnapshots(context.Context, string, string) ([]model.SnapshotView, error)
	DeleteVolumeSnapshot(context.Context, string, string, string) (*model.ActionResult, error)
	RollbackVolumeSnapshot(context.Context, string, string, string) (*model.SnapshotView, error)

	ListNetworks(context.Context) ([]model.NetworkView, error)
	GetNetwork(context.Context, string) (*model.NetworkView, error)
	DefineNetwork(context.Context, model.NetworkDefineSpec) (*model.NetworkView, error)
	StartNetwork(context.Context, string) (*model.NetworkView, error)
	DestroyNetwork(context.Context, string) (*model.NetworkView, error)
	SetNetworkAutostart(context.Context, string, bool) (*model.NetworkView, error)
	UndefineNetwork(context.Context, string) (*model.ActionResult, error)

	ListInterfaces(context.Context) ([]model.InterfaceView, error)
	GetInterface(context.Context, string) (*model.InterfaceView, error)
	DefineInterface(context.Context, model.InterfaceDefineSpec) (*model.InterfaceView, error)
	StartInterface(context.Context, string) (*model.InterfaceView, error)
	DestroyInterface(context.Context, string) (*model.InterfaceView, error)
	UndefineInterface(context.Context, string) (*model.ActionResult, error)
}
