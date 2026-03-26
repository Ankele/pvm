package server

import (
	"context"

	pvmv1 "github.com/ankele/pvm/api/gen/pvm/v1"
	"github.com/ankele/pvm/internal/model"
	"github.com/ankele/pvm/internal/service"
	"google.golang.org/protobuf/types/known/emptypb"
)

type APIServer struct {
	pvmv1.UnimplementedSystemServiceServer
	pvmv1.UnimplementedVMServiceServer
	pvmv1.UnimplementedStorageServiceServer
	pvmv1.UnimplementedNetworkServiceServer
	pvmv1.UnimplementedInterfaceServiceServer
	pvmv1.UnimplementedSnapshotServiceServer

	manager *service.Manager
}

func NewAPIServer(manager *service.Manager) *APIServer {
	return &APIServer{manager: manager}
}

func (s *APIServer) GetInfo(ctx context.Context, _ *emptypb.Empty) (*pvmv1.SystemInfoResponse, error) {
	info, err := s.manager.Info(ctx)
	if err != nil {
		return nil, grpcError(err)
	}
	return &pvmv1.SystemInfoResponse{Info: toProtoSystemInfo(info)}, nil
}

func (s *APIServer) Health(context.Context, *emptypb.Empty) (*pvmv1.HealthResponse, error) {
	return &pvmv1.HealthResponse{Status: "SERVING"}, nil
}

func (s *APIServer) ListVMs(ctx context.Context, _ *emptypb.Empty) (*pvmv1.ListVMsResponse, error) {
	items, err := s.manager.ListVMs(ctx)
	if err != nil {
		return nil, grpcError(err)
	}
	out := &pvmv1.ListVMsResponse{}
	for _, item := range items {
		vm := item
		out.Items = append(out.Items, toProtoVM(&vm))
	}
	return out, nil
}

func (s *APIServer) GetVM(ctx context.Context, req *pvmv1.NameRequest) (*pvmv1.VMResponse, error) {
	vm, err := s.manager.GetVM(ctx, req.GetName())
	if err != nil {
		return nil, grpcError(err)
	}
	return &pvmv1.VMResponse{Vm: toProtoVM(vm)}, nil
}

func (s *APIServer) DefineVM(ctx context.Context, req *pvmv1.DefineVMRequest) (*pvmv1.VMResponse, error) {
	vm, err := s.manager.DefineVM(ctx, domainSpecFromProto(req.GetSpec()))
	if err != nil {
		return nil, grpcError(err)
	}
	return &pvmv1.VMResponse{Vm: toProtoVM(vm)}, nil
}

func (s *APIServer) InstallVM(ctx context.Context, req *pvmv1.InstallVMRequest) (*pvmv1.VMResponse, error) {
	vm, err := s.manager.InstallVM(ctx, installSpecFromProto(req.GetSpec()))
	if err != nil {
		return nil, grpcError(err)
	}
	return &pvmv1.VMResponse{Vm: toProtoVM(vm)}, nil
}

func (s *APIServer) LaunchVM(ctx context.Context, req *pvmv1.LaunchVMRequest) (*pvmv1.VMResponse, error) {
	vm, err := s.manager.LaunchVM(ctx, launchSpecFromProto(req.GetSpec()))
	if err != nil {
		return nil, grpcError(err)
	}
	return &pvmv1.VMResponse{Vm: toProtoVM(vm)}, nil
}

func (s *APIServer) StartVM(ctx context.Context, req *pvmv1.NameRequest) (*pvmv1.VMResponse, error) {
	return s.wrapVM(ctx, req.GetName(), s.manager.StartVM)
}
func (s *APIServer) ShutdownVM(ctx context.Context, req *pvmv1.NameRequest) (*pvmv1.VMResponse, error) {
	return s.wrapVM(ctx, req.GetName(), s.manager.ShutdownVM)
}
func (s *APIServer) DestroyVM(ctx context.Context, req *pvmv1.NameRequest) (*pvmv1.VMResponse, error) {
	return s.wrapVM(ctx, req.GetName(), s.manager.DestroyVM)
}
func (s *APIServer) RebootVM(ctx context.Context, req *pvmv1.NameRequest) (*pvmv1.VMResponse, error) {
	return s.wrapVM(ctx, req.GetName(), s.manager.RebootVM)
}
func (s *APIServer) PauseVM(ctx context.Context, req *pvmv1.NameRequest) (*pvmv1.VMResponse, error) {
	return s.wrapVM(ctx, req.GetName(), s.manager.PauseVM)
}
func (s *APIServer) ResumeVM(ctx context.Context, req *pvmv1.NameRequest) (*pvmv1.VMResponse, error) {
	return s.wrapVM(ctx, req.GetName(), s.manager.ResumeVM)
}

func (s *APIServer) UndefineVM(ctx context.Context, req *pvmv1.NameRequest) (*pvmv1.ActionResponse, error) {
	res, err := s.manager.UndefineVM(ctx, req.GetName())
	if err != nil {
		return nil, grpcError(err)
	}
	return &pvmv1.ActionResponse{Message: res.Message}, nil
}

func (s *APIServer) AddNIC(ctx context.Context, req *pvmv1.AddNICRequest) (*pvmv1.VMResponse, error) {
	spec := nicSpecFromProto(req.GetSpec())
	spec.VM = req.GetVm()
	vm, err := s.manager.AddVMNIC(ctx, spec)
	if err != nil {
		return nil, grpcError(err)
	}
	return &pvmv1.VMResponse{Vm: toProtoVM(vm)}, nil
}

func (s *APIServer) UpdateNIC(ctx context.Context, req *pvmv1.UpdateNICRequest) (*pvmv1.VMResponse, error) {
	spec := model.VMNICUpdateSpec{VMNICSpec: nicSpecFromProto(req.GetSpec())}
	spec.VM = req.GetVm()
	spec.Alias = req.GetAlias()
	vm, err := s.manager.UpdateVMNIC(ctx, spec)
	if err != nil {
		return nil, grpcError(err)
	}
	return &pvmv1.VMResponse{Vm: toProtoVM(vm)}, nil
}

func (s *APIServer) RemoveNIC(ctx context.Context, req *pvmv1.RemoveNICRequest) (*pvmv1.VMResponse, error) {
	vm, err := s.manager.RemoveVMNIC(ctx, req.GetVm(), req.GetAlias())
	if err != nil {
		return nil, grpcError(err)
	}
	return &pvmv1.VMResponse{Vm: toProtoVM(vm)}, nil
}

func (s *APIServer) GetGraphics(ctx context.Context, req *pvmv1.NameRequest) (*pvmv1.GraphicsInfoResponse, error) {
	info, err := s.manager.GetVMGraphics(ctx, req.GetName())
	if err != nil {
		return nil, grpcError(err)
	}
	return &pvmv1.GraphicsInfoResponse{Info: toProtoGraphicsInfo(info)}, nil
}

func (s *APIServer) ListPools(ctx context.Context, _ *emptypb.Empty) (*pvmv1.ListPoolsResponse, error) {
	items, err := s.manager.ListPools(ctx)
	if err != nil {
		return nil, grpcError(err)
	}
	out := &pvmv1.ListPoolsResponse{}
	for _, item := range items {
		pool := item
		out.Items = append(out.Items, toProtoPool(&pool))
	}
	return out, nil
}

func (s *APIServer) GetPool(ctx context.Context, req *pvmv1.NameRequest) (*pvmv1.PoolResponse, error) {
	pool, err := s.manager.GetPool(ctx, req.GetName())
	if err != nil {
		return nil, grpcError(err)
	}
	return &pvmv1.PoolResponse{Pool: toProtoPool(pool)}, nil
}

func (s *APIServer) DefinePool(ctx context.Context, req *pvmv1.DefinePoolRequest) (*pvmv1.PoolResponse, error) {
	pool, err := s.manager.DefinePool(ctx, poolSpecFromProto(req.GetSpec()))
	if err != nil {
		return nil, grpcError(err)
	}
	return &pvmv1.PoolResponse{Pool: toProtoPool(pool)}, nil
}

func (s *APIServer) StartPool(ctx context.Context, req *pvmv1.NameRequest) (*pvmv1.PoolResponse, error) {
	return s.wrapPool(ctx, req.GetName(), s.manager.StartPool)
}
func (s *APIServer) DestroyPool(ctx context.Context, req *pvmv1.NameRequest) (*pvmv1.PoolResponse, error) {
	return s.wrapPool(ctx, req.GetName(), s.manager.DestroyPool)
}
func (s *APIServer) RefreshPool(ctx context.Context, req *pvmv1.NameRequest) (*pvmv1.PoolResponse, error) {
	return s.wrapPool(ctx, req.GetName(), s.manager.RefreshPool)
}

func (s *APIServer) SetPoolAutostart(ctx context.Context, req *pvmv1.SetAutostartRequest) (*pvmv1.PoolResponse, error) {
	pool, err := s.manager.SetPoolAutostart(ctx, req.GetName(), req.GetEnabled())
	if err != nil {
		return nil, grpcError(err)
	}
	return &pvmv1.PoolResponse{Pool: toProtoPool(pool)}, nil
}

func (s *APIServer) UndefinePool(ctx context.Context, req *pvmv1.NameRequest) (*pvmv1.ActionResponse, error) {
	res, err := s.manager.UndefinePool(ctx, req.GetName())
	if err != nil {
		return nil, grpcError(err)
	}
	return &pvmv1.ActionResponse{Message: res.Message}, nil
}

func (s *APIServer) ListVolumes(ctx context.Context, req *pvmv1.PoolNameRequest) (*pvmv1.ListVolumesResponse, error) {
	items, err := s.manager.ListVolumes(ctx, req.GetPool())
	if err != nil {
		return nil, grpcError(err)
	}
	out := &pvmv1.ListVolumesResponse{}
	for _, item := range items {
		vol := item
		out.Items = append(out.Items, toProtoVolume(&vol))
	}
	return out, nil
}

func (s *APIServer) GetVolume(ctx context.Context, req *pvmv1.GetVolumeRequest) (*pvmv1.VolumeResponse, error) {
	vol, err := s.manager.GetVolume(ctx, req.GetPool(), req.GetName())
	if err != nil {
		return nil, grpcError(err)
	}
	return &pvmv1.VolumeResponse{Volume: toProtoVolume(vol)}, nil
}

func (s *APIServer) CreateVolume(ctx context.Context, req *pvmv1.CreateVolumeRequest) (*pvmv1.VolumeResponse, error) {
	vol, err := s.manager.CreateVolume(ctx, volumeCreateSpecFromProto(req.GetSpec()))
	if err != nil {
		return nil, grpcError(err)
	}
	return &pvmv1.VolumeResponse{Volume: toProtoVolume(vol)}, nil
}

func (s *APIServer) DeleteVolume(ctx context.Context, req *pvmv1.GetVolumeRequest) (*pvmv1.ActionResponse, error) {
	res, err := s.manager.DeleteVolume(ctx, req.GetPool(), req.GetName())
	if err != nil {
		return nil, grpcError(err)
	}
	return &pvmv1.ActionResponse{Message: res.Message}, nil
}

func (s *APIServer) ResizeVolume(ctx context.Context, req *pvmv1.ResizeVolumeRequest) (*pvmv1.VolumeResponse, error) {
	vol, err := s.manager.ResizeVolume(ctx, volumeResizeSpecFromProto(req.GetSpec()))
	if err != nil {
		return nil, grpcError(err)
	}
	return &pvmv1.VolumeResponse{Volume: toProtoVolume(vol)}, nil
}

func (s *APIServer) ListNetworks(ctx context.Context, _ *emptypb.Empty) (*pvmv1.ListNetworksResponse, error) {
	items, err := s.manager.ListNetworks(ctx)
	if err != nil {
		return nil, grpcError(err)
	}
	out := &pvmv1.ListNetworksResponse{}
	for _, item := range items {
		net := item
		out.Items = append(out.Items, toProtoNetwork(&net))
	}
	return out, nil
}

func (s *APIServer) GetNetwork(ctx context.Context, req *pvmv1.NameRequest) (*pvmv1.NetworkResponse, error) {
	net, err := s.manager.GetNetwork(ctx, req.GetName())
	if err != nil {
		return nil, grpcError(err)
	}
	return &pvmv1.NetworkResponse{Network: toProtoNetwork(net)}, nil
}

func (s *APIServer) DefineNetwork(ctx context.Context, req *pvmv1.DefineNetworkRequest) (*pvmv1.NetworkResponse, error) {
	net, err := s.manager.DefineNetwork(ctx, networkSpecFromProto(req.GetSpec()))
	if err != nil {
		return nil, grpcError(err)
	}
	return &pvmv1.NetworkResponse{Network: toProtoNetwork(net)}, nil
}

func (s *APIServer) StartNetwork(ctx context.Context, req *pvmv1.NameRequest) (*pvmv1.NetworkResponse, error) {
	return s.wrapNetwork(ctx, req.GetName(), s.manager.StartNetwork)
}
func (s *APIServer) DestroyNetwork(ctx context.Context, req *pvmv1.NameRequest) (*pvmv1.NetworkResponse, error) {
	return s.wrapNetwork(ctx, req.GetName(), s.manager.DestroyNetwork)
}

func (s *APIServer) SetNetworkAutostart(ctx context.Context, req *pvmv1.SetAutostartRequest) (*pvmv1.NetworkResponse, error) {
	net, err := s.manager.SetNetworkAutostart(ctx, req.GetName(), req.GetEnabled())
	if err != nil {
		return nil, grpcError(err)
	}
	return &pvmv1.NetworkResponse{Network: toProtoNetwork(net)}, nil
}

func (s *APIServer) UndefineNetwork(ctx context.Context, req *pvmv1.NameRequest) (*pvmv1.ActionResponse, error) {
	res, err := s.manager.UndefineNetwork(ctx, req.GetName())
	if err != nil {
		return nil, grpcError(err)
	}
	return &pvmv1.ActionResponse{Message: res.Message}, nil
}

func (s *APIServer) ListInterfaces(ctx context.Context, _ *emptypb.Empty) (*pvmv1.ListInterfacesResponse, error) {
	items, err := s.manager.ListInterfaces(ctx)
	if err != nil {
		return nil, grpcError(err)
	}
	out := &pvmv1.ListInterfacesResponse{}
	for _, item := range items {
		iface := item
		out.Items = append(out.Items, toProtoInterface(&iface))
	}
	return out, nil
}

func (s *APIServer) GetInterface(ctx context.Context, req *pvmv1.NameRequest) (*pvmv1.InterfaceResponse, error) {
	iface, err := s.manager.GetInterface(ctx, req.GetName())
	if err != nil {
		return nil, grpcError(err)
	}
	return &pvmv1.InterfaceResponse{Iface: toProtoInterface(iface)}, nil
}

func (s *APIServer) DefineInterface(ctx context.Context, req *pvmv1.DefineInterfaceRequest) (*pvmv1.InterfaceResponse, error) {
	iface, err := s.manager.DefineInterface(ctx, interfaceSpecFromProto(req.GetSpec()))
	if err != nil {
		return nil, grpcError(err)
	}
	return &pvmv1.InterfaceResponse{Iface: toProtoInterface(iface)}, nil
}

func (s *APIServer) StartInterface(ctx context.Context, req *pvmv1.NameRequest) (*pvmv1.InterfaceResponse, error) {
	return s.wrapInterface(ctx, req.GetName(), s.manager.StartInterface)
}
func (s *APIServer) DestroyInterface(ctx context.Context, req *pvmv1.NameRequest) (*pvmv1.InterfaceResponse, error) {
	return s.wrapInterface(ctx, req.GetName(), s.manager.DestroyInterface)
}

func (s *APIServer) UndefineInterface(ctx context.Context, req *pvmv1.NameRequest) (*pvmv1.ActionResponse, error) {
	res, err := s.manager.UndefineInterface(ctx, req.GetName())
	if err != nil {
		return nil, grpcError(err)
	}
	return &pvmv1.ActionResponse{Message: res.Message}, nil
}

func (s *APIServer) CreateVMSnapshot(ctx context.Context, req *pvmv1.CreateVMSnapshotRequest) (*pvmv1.SnapshotResponse, error) {
	snap, err := s.manager.CreateDomainSnapshot(ctx, vmSnapshotSpecFromProto(req.GetSpec()))
	if err != nil {
		return nil, grpcError(err)
	}
	return &pvmv1.SnapshotResponse{Snapshot: toProtoSnapshot(snap)}, nil
}

func (s *APIServer) ListVMSnapshots(ctx context.Context, req *pvmv1.NameRequest) (*pvmv1.ListSnapshotsResponse, error) {
	items, err := s.manager.ListDomainSnapshots(ctx, req.GetName())
	if err != nil {
		return nil, grpcError(err)
	}
	out := &pvmv1.ListSnapshotsResponse{}
	for _, item := range items {
		snap := item
		out.Items = append(out.Items, toProtoSnapshot(&snap))
	}
	return out, nil
}

func (s *APIServer) DeleteVMSnapshot(ctx context.Context, req *pvmv1.DeleteVMSnapshotRequest) (*pvmv1.ActionResponse, error) {
	res, err := s.manager.DeleteDomainSnapshot(ctx, req.GetVm(), req.GetSnapshot())
	if err != nil {
		return nil, grpcError(err)
	}
	return &pvmv1.ActionResponse{Message: res.Message}, nil
}

func (s *APIServer) RevertVMSnapshot(ctx context.Context, req *pvmv1.DeleteVMSnapshotRequest) (*pvmv1.SnapshotResponse, error) {
	snap, err := s.manager.RevertDomainSnapshot(ctx, req.GetVm(), req.GetSnapshot())
	if err != nil {
		return nil, grpcError(err)
	}
	return &pvmv1.SnapshotResponse{Snapshot: toProtoSnapshot(snap)}, nil
}

func (s *APIServer) CreateVolumeSnapshot(ctx context.Context, req *pvmv1.CreateVolumeSnapshotRequest) (*pvmv1.SnapshotResponse, error) {
	snap, err := s.manager.CreateVolumeSnapshot(ctx, volumeSnapshotSpecFromProto(req.GetSpec()))
	if err != nil {
		return nil, grpcError(err)
	}
	return &pvmv1.SnapshotResponse{Snapshot: toProtoSnapshot(snap)}, nil
}

func (s *APIServer) ListVolumeSnapshots(ctx context.Context, req *pvmv1.GetVolumeRequest) (*pvmv1.ListSnapshotsResponse, error) {
	items, err := s.manager.ListVolumeSnapshots(ctx, req.GetPool(), req.GetName())
	if err != nil {
		return nil, grpcError(err)
	}
	out := &pvmv1.ListSnapshotsResponse{}
	for _, item := range items {
		snap := item
		out.Items = append(out.Items, toProtoSnapshot(&snap))
	}
	return out, nil
}

func (s *APIServer) DeleteVolumeSnapshot(ctx context.Context, req *pvmv1.DeleteVolumeSnapshotRequest) (*pvmv1.ActionResponse, error) {
	res, err := s.manager.DeleteVolumeSnapshot(ctx, req.GetPool(), req.GetVolume(), req.GetSnapshot())
	if err != nil {
		return nil, grpcError(err)
	}
	return &pvmv1.ActionResponse{Message: res.Message}, nil
}

func (s *APIServer) RollbackVolumeSnapshot(ctx context.Context, req *pvmv1.DeleteVolumeSnapshotRequest) (*pvmv1.SnapshotResponse, error) {
	snap, err := s.manager.RollbackVolumeSnapshot(ctx, req.GetPool(), req.GetVolume(), req.GetSnapshot())
	if err != nil {
		return nil, grpcError(err)
	}
	return &pvmv1.SnapshotResponse{Snapshot: toProtoSnapshot(snap)}, nil
}

func (s *APIServer) wrapVM(ctx context.Context, name string, fn func(context.Context, string) (*model.VMView, error)) (*pvmv1.VMResponse, error) {
	vm, err := fn(ctx, name)
	if err != nil {
		return nil, grpcError(err)
	}
	return &pvmv1.VMResponse{Vm: toProtoVM(vm)}, nil
}

func (s *APIServer) wrapPool(ctx context.Context, name string, fn func(context.Context, string) (*model.StoragePoolView, error)) (*pvmv1.PoolResponse, error) {
	pool, err := fn(ctx, name)
	if err != nil {
		return nil, grpcError(err)
	}
	return &pvmv1.PoolResponse{Pool: toProtoPool(pool)}, nil
}

func (s *APIServer) wrapNetwork(ctx context.Context, name string, fn func(context.Context, string) (*model.NetworkView, error)) (*pvmv1.NetworkResponse, error) {
	net, err := fn(ctx, name)
	if err != nil {
		return nil, grpcError(err)
	}
	return &pvmv1.NetworkResponse{Network: toProtoNetwork(net)}, nil
}

func (s *APIServer) wrapInterface(ctx context.Context, name string, fn func(context.Context, string) (*model.InterfaceView, error)) (*pvmv1.InterfaceResponse, error) {
	iface, err := fn(ctx, name)
	if err != nil {
		return nil, grpcError(err)
	}
	return &pvmv1.InterfaceResponse{Iface: toProtoInterface(iface)}, nil
}
