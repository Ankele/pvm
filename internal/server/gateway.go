package server

import (
	"context"
	"net/http"

	pvmv1 "github.com/ankele/pvm/api/gen/pvm/v1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/emptypb"
)

func NewGatewayMux(api *APIServer) (*runtime.ServeMux, error) {
	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
	)
	register := func(method, pattern string, handler func(context.Context, *http.Request, map[string]string) (proto.Message, error)) error {
		return mux.HandlePath(method, pattern, func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
			ctx := r.Context()
			resp, err := handler(ctx, r, pathParams)
			_, outbound := runtime.MarshalerForRequest(mux, r)
			if err != nil {
				runtime.HTTPError(ctx, mux, outbound, w, r, err)
				return
			}
			runtime.ForwardResponseMessage(ctx, mux, outbound, w, r, resp)
		})
	}

	decode := func(r *http.Request, msg protoreflect.ProtoMessage) error {
		if r.Body == nil || r.ContentLength == 0 {
			return nil
		}
		inbound, _ := runtime.MarshalerForRequest(mux, r)
		return inbound.NewDecoder(r.Body).Decode(msg)
	}

	for _, route := range []struct {
		method  string
		pattern string
		handler func(context.Context, *http.Request, map[string]string) (proto.Message, error)
	}{
		{"GET", "/v1/system", func(ctx context.Context, _ *http.Request, _ map[string]string) (proto.Message, error) {
			return api.GetInfo(ctx, &emptypb.Empty{})
		}},
		{"GET", "/v1/health", func(ctx context.Context, _ *http.Request, _ map[string]string) (proto.Message, error) {
			return api.Health(ctx, &emptypb.Empty{})
		}},
		{"GET", "/v1/vms", func(ctx context.Context, _ *http.Request, _ map[string]string) (proto.Message, error) {
			return api.ListVMs(ctx, &emptypb.Empty{})
		}},
		{"GET", "/v1/vms/{name}", func(ctx context.Context, _ *http.Request, p map[string]string) (proto.Message, error) {
			return api.GetVM(ctx, &pvmv1.NameRequest{Name: p["name"]})
		}},
		{"POST", "/v1/vms:define", func(ctx context.Context, r *http.Request, _ map[string]string) (proto.Message, error) {
			req := &pvmv1.DefineVMRequest{}
			if err := decode(r, req); err != nil {
				return nil, err
			}
			return api.DefineVM(ctx, req)
		}},
		{"POST", "/v1/vms:install", func(ctx context.Context, r *http.Request, _ map[string]string) (proto.Message, error) {
			req := &pvmv1.InstallVMRequest{}
			if err := decode(r, req); err != nil {
				return nil, err
			}
			return api.InstallVM(ctx, req)
		}},
		{"POST", "/v1/vms:launch", func(ctx context.Context, r *http.Request, _ map[string]string) (proto.Message, error) {
			req := &pvmv1.LaunchVMRequest{}
			if err := decode(r, req); err != nil {
				return nil, err
			}
			return api.LaunchVM(ctx, req)
		}},
		{"POST", "/v1/vms/{name}:start", func(ctx context.Context, _ *http.Request, p map[string]string) (proto.Message, error) {
			return api.StartVM(ctx, &pvmv1.NameRequest{Name: p["name"]})
		}},
		{"POST", "/v1/vms/{name}:shutdown", func(ctx context.Context, _ *http.Request, p map[string]string) (proto.Message, error) {
			return api.ShutdownVM(ctx, &pvmv1.NameRequest{Name: p["name"]})
		}},
		{"POST", "/v1/vms/{name}:destroy", func(ctx context.Context, _ *http.Request, p map[string]string) (proto.Message, error) {
			return api.DestroyVM(ctx, &pvmv1.NameRequest{Name: p["name"]})
		}},
		{"POST", "/v1/vms/{name}:reboot", func(ctx context.Context, _ *http.Request, p map[string]string) (proto.Message, error) {
			return api.RebootVM(ctx, &pvmv1.NameRequest{Name: p["name"]})
		}},
		{"POST", "/v1/vms/{name}:pause", func(ctx context.Context, _ *http.Request, p map[string]string) (proto.Message, error) {
			return api.PauseVM(ctx, &pvmv1.NameRequest{Name: p["name"]})
		}},
		{"POST", "/v1/vms/{name}:resume", func(ctx context.Context, _ *http.Request, p map[string]string) (proto.Message, error) {
			return api.ResumeVM(ctx, &pvmv1.NameRequest{Name: p["name"]})
		}},
		{"DELETE", "/v1/vms/{name}", func(ctx context.Context, _ *http.Request, p map[string]string) (proto.Message, error) {
			return api.UndefineVM(ctx, &pvmv1.NameRequest{Name: p["name"]})
		}},
		{"POST", "/v1/vms/{vm}/nics", func(ctx context.Context, r *http.Request, p map[string]string) (proto.Message, error) {
			req := &pvmv1.AddNICRequest{Vm: p["vm"]}
			if err := decode(r, req); err != nil {
				return nil, err
			}
			req.Vm = p["vm"]
			return api.AddNIC(ctx, req)
		}},
		{"PATCH", "/v1/vms/{vm}/nics/{alias}", func(ctx context.Context, r *http.Request, p map[string]string) (proto.Message, error) {
			req := &pvmv1.UpdateNICRequest{Vm: p["vm"], Alias: p["alias"]}
			if err := decode(r, req); err != nil {
				return nil, err
			}
			req.Vm, req.Alias = p["vm"], p["alias"]
			return api.UpdateNIC(ctx, req)
		}},
		{"DELETE", "/v1/vms/{vm}/nics/{alias}", func(ctx context.Context, _ *http.Request, p map[string]string) (proto.Message, error) {
			return api.RemoveNIC(ctx, &pvmv1.RemoveNICRequest{Vm: p["vm"], Alias: p["alias"]})
		}},
		{"GET", "/v1/vms/{name}/graphics", func(ctx context.Context, _ *http.Request, p map[string]string) (proto.Message, error) {
			return api.GetGraphics(ctx, &pvmv1.NameRequest{Name: p["name"]})
		}},
		{"GET", "/v1/pools", func(ctx context.Context, _ *http.Request, _ map[string]string) (proto.Message, error) {
			return api.ListPools(ctx, &emptypb.Empty{})
		}},
		{"GET", "/v1/pools/{name}", func(ctx context.Context, _ *http.Request, p map[string]string) (proto.Message, error) {
			return api.GetPool(ctx, &pvmv1.NameRequest{Name: p["name"]})
		}},
		{"POST", "/v1/pools", func(ctx context.Context, r *http.Request, _ map[string]string) (proto.Message, error) {
			req := &pvmv1.DefinePoolRequest{}
			if err := decode(r, req); err != nil {
				return nil, err
			}
			return api.DefinePool(ctx, req)
		}},
		{"POST", "/v1/pools/{name}:start", func(ctx context.Context, _ *http.Request, p map[string]string) (proto.Message, error) {
			return api.StartPool(ctx, &pvmv1.NameRequest{Name: p["name"]})
		}},
		{"POST", "/v1/pools/{name}:destroy", func(ctx context.Context, _ *http.Request, p map[string]string) (proto.Message, error) {
			return api.DestroyPool(ctx, &pvmv1.NameRequest{Name: p["name"]})
		}},
		{"POST", "/v1/pools/{name}:autostart", func(ctx context.Context, r *http.Request, p map[string]string) (proto.Message, error) {
			req := &pvmv1.SetAutostartRequest{Name: p["name"]}
			if err := decode(r, req); err != nil {
				return nil, err
			}
			req.Name = p["name"]
			return api.SetPoolAutostart(ctx, req)
		}},
		{"DELETE", "/v1/pools/{name}", func(ctx context.Context, _ *http.Request, p map[string]string) (proto.Message, error) {
			return api.UndefinePool(ctx, &pvmv1.NameRequest{Name: p["name"]})
		}},
		{"POST", "/v1/pools/{name}:refresh", func(ctx context.Context, _ *http.Request, p map[string]string) (proto.Message, error) {
			return api.RefreshPool(ctx, &pvmv1.NameRequest{Name: p["name"]})
		}},
		{"GET", "/v1/pools/{pool}/volumes", func(ctx context.Context, _ *http.Request, p map[string]string) (proto.Message, error) {
			return api.ListVolumes(ctx, &pvmv1.PoolNameRequest{Pool: p["pool"]})
		}},
		{"GET", "/v1/pools/{pool}/volumes/{name}", func(ctx context.Context, _ *http.Request, p map[string]string) (proto.Message, error) {
			return api.GetVolume(ctx, &pvmv1.GetVolumeRequest{Pool: p["pool"], Name: p["name"]})
		}},
		{"POST", "/v1/pools/{pool}/volumes", func(ctx context.Context, r *http.Request, p map[string]string) (proto.Message, error) {
			req := &pvmv1.CreateVolumeRequest{Spec: &pvmv1.VolumeCreateSpec{Pool: p["pool"]}}
			if err := decode(r, req); err != nil {
				return nil, err
			}
			if req.Spec == nil {
				req.Spec = &pvmv1.VolumeCreateSpec{}
			}
			req.Spec.Pool = p["pool"]
			return api.CreateVolume(ctx, req)
		}},
		{"DELETE", "/v1/pools/{pool}/volumes/{name}", func(ctx context.Context, _ *http.Request, p map[string]string) (proto.Message, error) {
			return api.DeleteVolume(ctx, &pvmv1.GetVolumeRequest{Pool: p["pool"], Name: p["name"]})
		}},
		{"POST", "/v1/pools/{pool}/volumes/{name}:resize", func(ctx context.Context, r *http.Request, p map[string]string) (proto.Message, error) {
			req := &pvmv1.ResizeVolumeRequest{Spec: &pvmv1.VolumeResizeSpec{Pool: p["pool"], Name: p["name"]}}
			if err := decode(r, req); err != nil {
				return nil, err
			}
			if req.Spec == nil {
				req.Spec = &pvmv1.VolumeResizeSpec{}
			}
			req.Spec.Pool, req.Spec.Name = p["pool"], p["name"]
			return api.ResizeVolume(ctx, req)
		}},
		{"GET", "/v1/networks", func(ctx context.Context, _ *http.Request, _ map[string]string) (proto.Message, error) {
			return api.ListNetworks(ctx, &emptypb.Empty{})
		}},
		{"GET", "/v1/networks/{name}", func(ctx context.Context, _ *http.Request, p map[string]string) (proto.Message, error) {
			return api.GetNetwork(ctx, &pvmv1.NameRequest{Name: p["name"]})
		}},
		{"POST", "/v1/networks", func(ctx context.Context, r *http.Request, _ map[string]string) (proto.Message, error) {
			req := &pvmv1.DefineNetworkRequest{}
			if err := decode(r, req); err != nil {
				return nil, err
			}
			return api.DefineNetwork(ctx, req)
		}},
		{"POST", "/v1/networks/{name}:start", func(ctx context.Context, _ *http.Request, p map[string]string) (proto.Message, error) {
			return api.StartNetwork(ctx, &pvmv1.NameRequest{Name: p["name"]})
		}},
		{"POST", "/v1/networks/{name}:destroy", func(ctx context.Context, _ *http.Request, p map[string]string) (proto.Message, error) {
			return api.DestroyNetwork(ctx, &pvmv1.NameRequest{Name: p["name"]})
		}},
		{"POST", "/v1/networks/{name}:autostart", func(ctx context.Context, r *http.Request, p map[string]string) (proto.Message, error) {
			req := &pvmv1.SetAutostartRequest{Name: p["name"]}
			if err := decode(r, req); err != nil {
				return nil, err
			}
			req.Name = p["name"]
			return api.SetNetworkAutostart(ctx, req)
		}},
		{"DELETE", "/v1/networks/{name}", func(ctx context.Context, _ *http.Request, p map[string]string) (proto.Message, error) {
			return api.UndefineNetwork(ctx, &pvmv1.NameRequest{Name: p["name"]})
		}},
		{"GET", "/v1/interfaces", func(ctx context.Context, _ *http.Request, _ map[string]string) (proto.Message, error) {
			return api.ListInterfaces(ctx, &emptypb.Empty{})
		}},
		{"GET", "/v1/interfaces/{name}", func(ctx context.Context, _ *http.Request, p map[string]string) (proto.Message, error) {
			return api.GetInterface(ctx, &pvmv1.NameRequest{Name: p["name"]})
		}},
		{"POST", "/v1/interfaces", func(ctx context.Context, r *http.Request, _ map[string]string) (proto.Message, error) {
			req := &pvmv1.DefineInterfaceRequest{}
			if err := decode(r, req); err != nil {
				return nil, err
			}
			return api.DefineInterface(ctx, req)
		}},
		{"POST", "/v1/interfaces/{name}:start", func(ctx context.Context, _ *http.Request, p map[string]string) (proto.Message, error) {
			return api.StartInterface(ctx, &pvmv1.NameRequest{Name: p["name"]})
		}},
		{"POST", "/v1/interfaces/{name}:destroy", func(ctx context.Context, _ *http.Request, p map[string]string) (proto.Message, error) {
			return api.DestroyInterface(ctx, &pvmv1.NameRequest{Name: p["name"]})
		}},
		{"DELETE", "/v1/interfaces/{name}", func(ctx context.Context, _ *http.Request, p map[string]string) (proto.Message, error) {
			return api.UndefineInterface(ctx, &pvmv1.NameRequest{Name: p["name"]})
		}},
		{"POST", "/v1/vms/{vm}/snapshots", func(ctx context.Context, r *http.Request, p map[string]string) (proto.Message, error) {
			req := &pvmv1.CreateVMSnapshotRequest{Spec: &pvmv1.VMSnapshotSpec{Vm: p["vm"]}}
			if err := decode(r, req); err != nil {
				return nil, err
			}
			if req.Spec == nil {
				req.Spec = &pvmv1.VMSnapshotSpec{}
			}
			req.Spec.Vm = p["vm"]
			return api.CreateVMSnapshot(ctx, req)
		}},
		{"GET", "/v1/vms/{name}/snapshots", func(ctx context.Context, _ *http.Request, p map[string]string) (proto.Message, error) {
			return api.ListVMSnapshots(ctx, &pvmv1.NameRequest{Name: p["name"]})
		}},
		{"DELETE", "/v1/vms/{vm}/snapshots/{snapshot}", func(ctx context.Context, _ *http.Request, p map[string]string) (proto.Message, error) {
			return api.DeleteVMSnapshot(ctx, &pvmv1.DeleteVMSnapshotRequest{Vm: p["vm"], Snapshot: p["snapshot"]})
		}},
		{"POST", "/v1/vms/{vm}/snapshots/{snapshot}:revert", func(ctx context.Context, _ *http.Request, p map[string]string) (proto.Message, error) {
			return api.RevertVMSnapshot(ctx, &pvmv1.DeleteVMSnapshotRequest{Vm: p["vm"], Snapshot: p["snapshot"]})
		}},
		{"POST", "/v1/pools/{pool}/volumes/{volume}/snapshots", func(ctx context.Context, r *http.Request, p map[string]string) (proto.Message, error) {
			req := &pvmv1.CreateVolumeSnapshotRequest{Spec: &pvmv1.VolumeSnapshotSpec{Pool: p["pool"], Volume: p["volume"]}}
			if err := decode(r, req); err != nil {
				return nil, err
			}
			if req.Spec == nil {
				req.Spec = &pvmv1.VolumeSnapshotSpec{}
			}
			req.Spec.Pool, req.Spec.Volume = p["pool"], p["volume"]
			return api.CreateVolumeSnapshot(ctx, req)
		}},
		{"GET", "/v1/pools/{pool}/volumes/{name}/snapshots", func(ctx context.Context, _ *http.Request, p map[string]string) (proto.Message, error) {
			return api.ListVolumeSnapshots(ctx, &pvmv1.GetVolumeRequest{Pool: p["pool"], Name: p["name"]})
		}},
		{"DELETE", "/v1/pools/{pool}/volumes/{volume}/snapshots/{snapshot}", func(ctx context.Context, _ *http.Request, p map[string]string) (proto.Message, error) {
			return api.DeleteVolumeSnapshot(ctx, &pvmv1.DeleteVolumeSnapshotRequest{Pool: p["pool"], Volume: p["volume"], Snapshot: p["snapshot"]})
		}},
		{"POST", "/v1/pools/{pool}/volumes/{volume}/snapshots/{snapshot}:rollback", func(ctx context.Context, _ *http.Request, p map[string]string) (proto.Message, error) {
			return api.RollbackVolumeSnapshot(ctx, &pvmv1.DeleteVolumeSnapshotRequest{Pool: p["pool"], Volume: p["volume"], Snapshot: p["snapshot"]})
		}},
	} {
		if err := register(route.method, route.pattern, route.handler); err != nil {
			return nil, err
		}
	}

	return mux, nil
}
