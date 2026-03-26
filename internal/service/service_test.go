package service

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/ankele/pvm/internal/model"
	"github.com/ankele/pvm/internal/testutil"
)

func TestInstallVMValidation(t *testing.T) {
	t.Parallel()

	manager := NewManager(&testutil.FakeBackend{
		InstallVMFn: func(context.Context, model.VMInstallSpec) (*model.VMView, error) {
			t.Fatal("backend should not be called for invalid install spec")
			return nil, nil
		},
	})

	_, err := manager.InstallVM(context.Background(), model.VMInstallSpec{
		Name:      "demo",
		MemoryMiB: 2048,
		VCPU:      2,
		Pool:      "images",
		DiskName:  "demo-root",
	})
	if err == nil || model.CodeOf(err) != model.ErrInvalidArgument || !strings.Contains(err.Error(), "iso path is required") {
		t.Fatalf("expected invalid iso path error, got %v", err)
	}
}

func TestLaunchVMValidation(t *testing.T) {
	t.Parallel()

	manager := NewManager(&testutil.FakeBackend{
		LaunchVMFn: func(context.Context, model.VMLaunchSpec) (*model.VMView, error) {
			t.Fatal("backend should not be called for invalid launch spec")
			return nil, nil
		},
	})

	testCases := []struct {
		name string
		spec model.VMLaunchSpec
		want string
	}{
		{
			name: "requires explicit mode",
			spec: model.VMLaunchSpec{Name: "demo", MemoryMiB: 2048, VCPU: 2, ImagePath: "/tmp/demo.qcow2"},
			want: "launch mode must be clone or direct",
		},
		{
			name: "requires image or volume",
			spec: model.VMLaunchSpec{Name: "demo", Mode: model.LaunchModeDirect, MemoryMiB: 2048, VCPU: 2},
			want: "either image path or source volume is required",
		},
		{
			name: "rejects both image and volume",
			spec: model.VMLaunchSpec{
				Name:         "demo",
				Mode:         model.LaunchModeDirect,
				MemoryMiB:    2048,
				VCPU:         2,
				ImagePath:    "/tmp/demo.qcow2",
				SourceVolume: "images/demo",
			},
			want: "image path and source volume are mutually exclusive",
		},
		{
			name: "clone requires target",
			spec: model.VMLaunchSpec{
				Name:      "demo",
				Mode:      model.LaunchModeClone,
				MemoryMiB: 2048,
				VCPU:      2,
				ImagePath: "/tmp/demo.qcow2",
			},
			want: "clone mode requires target pool and target volume name",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := manager.LaunchVM(context.Background(), tc.spec)
			if err == nil || model.CodeOf(err) != model.ErrInvalidArgument || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected %q, got %v", tc.want, err)
			}
		})
	}
}

func TestLaunchVMPassesValidatedSpecToBackend(t *testing.T) {
	t.Parallel()

	var got model.VMLaunchSpec
	manager := NewManager(&testutil.FakeBackend{
		LaunchVMFn: func(_ context.Context, spec model.VMLaunchSpec) (*model.VMView, error) {
			got = spec
			return &model.VMView{Name: spec.Name, State: model.VMStateRunning, SourceMode: string(spec.Mode), SharedSource: spec.Mode == model.LaunchModeDirect}, nil
		},
	})

	spec := model.VMLaunchSpec{
		Name:         "demo",
		Mode:         model.LaunchModeDirect,
		MemoryMiB:    4096,
		VCPU:         4,
		SourceVolume: "images/demo-root",
		Networks:     []model.VMNICSpec{{Network: "default", Model: "virtio"}},
		Graphics:     []model.GraphicsSpec{{Type: "vnc", Listen: "127.0.0.1", AutoPort: true}},
	}

	vm, err := manager.LaunchVM(context.Background(), spec)
	if err != nil {
		t.Fatalf("LaunchVM() error = %v", err)
	}
	if !reflect.DeepEqual(got, spec) {
		t.Fatalf("LaunchVM() forwarded spec mismatch\nwant: %#v\ngot:  %#v", spec, got)
	}
	if vm.Name != "demo" || !vm.SharedSource || vm.SourceMode != "direct" {
		t.Fatalf("LaunchVM() response mismatch: %#v", vm)
	}
}
