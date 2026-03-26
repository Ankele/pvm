package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ankele/pvm/internal/model"
	"github.com/ankele/pvm/internal/service"
	"github.com/ankele/pvm/internal/testutil"
)

func TestGatewayListVMsUsesSnakeCaseJSON(t *testing.T) {
	t.Parallel()

	api := NewAPIServer(service.NewManager(&testutil.FakeBackend{
		ListVMsFn: func(context.Context) ([]model.VMView, error) {
			return []model.VMView{{
				Name:         "demo",
				State:        model.VMStateRunning,
				SharedSource: true,
				SourceMode:   "direct",
				MemoryMiB:    4096,
				VCPU:         4,
			}}, nil
		},
	}))

	mux, err := NewGatewayMux(api)
	if err != nil {
		t.Fatalf("NewGatewayMux() error = %v", err)
	}

	handler := NewTokenAuth("secret-token").HTTPMiddleware(mux)
	req := httptest.NewRequest(http.MethodGet, "/v1/vms", nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{`"items"`, `"shared_source":true`, `"source_mode":"direct"`, `"memory_mib":"4096"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("response body missing %q\n%s", want, body)
		}
	}

	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
}

func TestGatewayPreservesNotFoundStatus(t *testing.T) {
	t.Parallel()

	api := NewAPIServer(service.NewManager(&testutil.FakeBackend{
		GetVMFn: func(context.Context, string) (*model.VMView, error) {
			return nil, model.Errorf(model.ErrNotFound, "vm not found")
		},
	}))

	mux, err := NewGatewayMux(api)
	if err != nil {
		t.Fatalf("NewGatewayMux() error = %v", err)
	}

	handler := NewTokenAuth("secret-token").HTTPMiddleware(mux)
	req := httptest.NewRequest(http.MethodGet, "/v1/vms/missing", nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "vm not found") {
		t.Fatalf("expected not found message, got %s", rec.Body.String())
	}
}
