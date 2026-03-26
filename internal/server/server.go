package server

import (
	"context"
	"net"
	"net/http"
	"time"

	pvmv1 "github.com/ankele/pvm/api/gen/pvm/v1"
	"github.com/ankele/pvm/internal/service"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type Config struct {
	HTTPAddr string
	GRPCAddr string
	Token    string
}

type Server struct {
	cfg     Config
	manager *service.Manager
}

func New(cfg Config, manager *service.Manager) *Server {
	return &Server{cfg: cfg, manager: manager}
}

func (s *Server) Run(ctx context.Context) error {
	auth := NewTokenAuth(s.cfg.Token)
	api := NewAPIServer(s.manager)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(auth.UnaryInterceptor(map[string]struct{}{
			"/grpc.health.v1.Health/Check": {},
		})),
	)
	pvmv1.RegisterSystemServiceServer(grpcServer, api)
	pvmv1.RegisterVMServiceServer(grpcServer, api)
	pvmv1.RegisterStorageServiceServer(grpcServer, api)
	pvmv1.RegisterNetworkServiceServer(grpcServer, api)
	pvmv1.RegisterInterfaceServiceServer(grpcServer, api)
	pvmv1.RegisterSnapshotServiceServer(grpcServer, api)

	healthSrv := health.NewServer()
	healthSrv.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(grpcServer, healthSrv)

	grpcLn, err := net.Listen("tcp", s.cfg.GRPCAddr)
	if err != nil {
		return err
	}
	defer grpcLn.Close()

	httpMux, err := NewGatewayMux(api)
	if err != nil {
		return err
	}

	rootMux := http.NewServeMux()
	rootMux.Handle("/v1/", auth.HTTPMiddleware(httpMux))
	rootMux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	httpServer := &http.Server{
		Addr:              s.cfg.HTTPAddr,
		Handler:           rootMux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		if err := grpcServer.Serve(grpcLn); err != nil {
			return err
		}
		return nil
	})
	group.Go(func() error {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	})
	group.Go(func() error {
		<-groupCtx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		grpcServer.GracefulStop()
		return httpServer.Shutdown(shutdownCtx)
	})
	return group.Wait()
}
