package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ankele/pvm/internal/model"
	"github.com/ankele/pvm/internal/service"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func runWithManager(cmd *cobra.Command, opts *rootOptions, deps Dependencies, fn func(context.Context, *cobra.Command, func(any) error, *serviceLike) error) error {
	ctx := cmd.Context()
	manager, err := opts.newManager(ctx, deps.NewBackend)
	if err != nil {
		return err
	}
	defer manager.Close()

	printer := func(v any) error {
		return opts.renderer(deps.Stdout).Print(v)
	}
	return fn(ctx, cmd, printer, &serviceLike{manager})
}

type serviceLike struct {
	*service.Manager
}

func mustGetString(cmd *cobra.Command, name string) string {
	value, _ := cmd.Flags().GetString(name)
	return strings.TrimSpace(value)
}

func mustGetStringSlice(cmd *cobra.Command, name string) []string {
	values, _ := cmd.Flags().GetStringArray(name)
	return values
}

func mustGetUint32(cmd *cobra.Command, name string) uint32 {
	value, _ := cmd.Flags().GetUint32(name)
	return value
}

func mustGetUint64(cmd *cobra.Command, name string) uint64 {
	value, _ := cmd.Flags().GetUint64(name)
	return value
}

func mustGetBool(cmd *cobra.Command, name string) bool {
	value, _ := cmd.Flags().GetBool(name)
	return value
}

func loadDomainSpec(path string) (model.DomainSpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return model.DomainSpec{}, err
	}
	var spec model.DomainSpec
	switch {
	case strings.HasSuffix(path, ".json"):
		err = json.Unmarshal(data, &spec)
	default:
		err = yaml.Unmarshal(data, &spec)
	}
	if err != nil {
		return model.DomainSpec{}, fmt.Errorf("decode %s: %w", path, err)
	}
	return spec, nil
}

func defaultGraphicsFlags(cmd *cobra.Command) {
	cmd.Flags().String("graphics-type", "vnc", "graphics type: vnc|spice")
	cmd.Flags().String("graphics-listen", "127.0.0.1", "graphics listen address")
	cmd.Flags().Int32("graphics-port", 0, "graphics port, 0 means auto")
	cmd.Flags().Int32("graphics-tls-port", 0, "graphics TLS port")
}

func collectGraphics(cmd *cobra.Command) []model.GraphicsSpec {
	typ := mustGetString(cmd, "graphics-type")
	if typ == "" {
		return nil
	}
	port, _ := cmd.Flags().GetInt32("graphics-port")
	tlsPort, _ := cmd.Flags().GetInt32("graphics-tls-port")
	return []model.GraphicsSpec{{
		Type:     typ,
		Listen:   mustGetString(cmd, "graphics-listen"),
		AutoPort: port == 0,
		Port:     port,
		TLSPort:  tlsPort,
	}}
}

func collectNetworks(cmd *cobra.Command) []model.VMNICSpec {
	values := mustGetStringSlice(cmd, "network")
	out := make([]model.VMNICSpec, 0, len(values))
	for _, name := range values {
		if strings.TrimSpace(name) == "" {
			continue
		}
		out = append(out, model.VMNICSpec{
			Network: name,
			Model:   "virtio",
		})
	}
	return out
}
