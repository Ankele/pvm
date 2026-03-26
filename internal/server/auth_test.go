package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ankele/pvm/internal/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestTokenAuthValidateHeader(t *testing.T) {
	t.Parallel()

	auth := NewTokenAuth("secret-token")

	if err := auth.ValidateHeader("Bearer secret-token"); err != nil {
		t.Fatalf("ValidateHeader() unexpected error = %v", err)
	}

	for _, tc := range []struct {
		name   string
		header string
	}{
		{name: "missing bearer prefix", header: "secret-token"},
		{name: "wrong token", header: "Bearer wrong"},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := auth.ValidateHeader(tc.header)
			if err == nil || model.CodeOf(err) != model.ErrUnauthenticated {
				t.Fatalf("expected unauthenticated error, got %v", err)
			}
		})
	}
}

func TestTokenAuthHTTPMiddleware(t *testing.T) {
	t.Parallel()

	auth := NewTokenAuth("secret-token")
	protected := auth.HTTPMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	t.Run("healthz bypasses auth", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		rec := httptest.NewRecorder()
		protected.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected 204, got %d", rec.Code)
		}
	})

	t.Run("protected path requires token", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/v1/vms", nil)
		rec := httptest.NewRecorder()
		protected.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("protected path accepts valid token", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/v1/vms", nil)
		req.Header.Set("Authorization", "Bearer secret-token")
		rec := httptest.NewRecorder()
		protected.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected 204, got %d", rec.Code)
		}
	})
}

func TestTokenAuthUnaryInterceptor(t *testing.T) {
	t.Parallel()

	auth := NewTokenAuth("secret-token")
	interceptor := auth.UnaryInterceptor(map[string]struct{}{
		"/grpc.health.v1.Health/Check": {},
	})

	handler := func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	}

	resp, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{
		FullMethod: "/grpc.health.v1.Health/Check",
	}, handler)
	if err != nil || resp != "ok" {
		t.Fatalf("exempt interceptor call = (%v, %v), want (ok, nil)", resp, err)
	}

	_, err = interceptor(context.Background(), nil, &grpc.UnaryServerInfo{
		FullMethod: "/pvm.v1.VMService/ListVMs",
	}, handler)
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("missing auth code = %v, want %v", status.Code(err), codes.Unauthenticated)
	}

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer secret-token"))
	resp, err = interceptor(ctx, nil, &grpc.UnaryServerInfo{
		FullMethod: "/pvm.v1.VMService/ListVMs",
	}, handler)
	if err != nil || resp != "ok" {
		t.Fatalf("authorized interceptor call = (%v, %v), want (ok, nil)", resp, err)
	}
}
