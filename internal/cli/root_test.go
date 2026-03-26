package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ankele/pvm/internal/model"
	"github.com/ankele/pvm/internal/testutil"
)

func TestRootCommandVMListJSON(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	cmd := NewRootCommand(Dependencies{
		NewBackend: testutil.Factory(&testutil.FakeBackend{
			ListVMsFn: func(context.Context) ([]model.VMView, error) {
				return []model.VMView{{
					Name:         "demo",
					State:        model.VMStateRunning,
					SharedSource: true,
					SourceMode:   "direct",
				}}, nil
			},
		}),
		Stdout: &stdout,
	})
	cmd.SetArgs([]string{"--output", "json", "vm", "list"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var got []model.VMView
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v\n%s", err, stdout.String())
	}
	if len(got) != 1 || got[0].Name != "demo" || !got[0].SharedSource || got[0].SourceMode != "direct" {
		t.Fatalf("unexpected vm list payload: %#v", got)
	}
}

func TestServeCommandRequiresToken(t *testing.T) {
	t.Setenv("PVM_API_TOKEN", "")

	cmd := NewRootCommand(Dependencies{
		NewBackend: testutil.Factory(&testutil.FakeBackend{}),
		Stdout:     &bytes.Buffer{},
		Stderr:     &bytes.Buffer{},
	})
	cmd.SetArgs([]string{"serve"})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "api token is required") {
		t.Fatalf("expected missing token error, got %v", err)
	}
}

func TestVMLaunchValidationFromCLI(t *testing.T) {
	t.Parallel()

	cmd := NewRootCommand(Dependencies{
		NewBackend: testutil.Factory(&testutil.FakeBackend{
			LaunchVMFn: func(context.Context, model.VMLaunchSpec) (*model.VMView, error) {
				t.Fatal("backend should not be called for invalid CLI input")
				return nil, nil
			},
		}),
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	})
	cmd.SetArgs([]string{"vm", "launch", "--name", "demo", "--mode", "direct"})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "either image path or source volume is required") {
		t.Fatalf("expected launch validation error, got %v", err)
	}
}
