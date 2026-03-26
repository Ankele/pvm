package zfs

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ankele/pvm/internal/model"
)

type Snapshot struct {
	Name      string
	Dataset   string
	CreatedAt time.Time
}

type Client struct {
	bin string
}

func New(bin string) *Client {
	if strings.TrimSpace(bin) == "" {
		bin = "zfs"
	}
	return &Client{bin: bin}
}

func (c *Client) Available(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, c.bin, "list", "-H")
	return cmd.Run() == nil
}

func DatasetForPath(path string) (string, error) {
	clean := filepath.Clean(path)
	if strings.HasPrefix(clean, "/dev/zvol/") {
		return strings.TrimPrefix(clean, "/dev/zvol/"), nil
	}
	return "", model.Errorf(model.ErrUnsupported, "path %q is not a ZFS zvol path", path)
}

func (c *Client) CreateSnapshot(ctx context.Context, dataset, name string) error {
	_, err := c.run(ctx, "snapshot", dataset+"@"+name)
	return err
}

func (c *Client) DestroySnapshot(ctx context.Context, dataset, name string) error {
	_, err := c.run(ctx, "destroy", dataset+"@"+name)
	return err
}

func (c *Client) RollbackSnapshot(ctx context.Context, dataset, name string) error {
	_, err := c.run(ctx, "rollback", "-r", dataset+"@"+name)
	return err
}

func (c *Client) ListSnapshots(ctx context.Context, dataset string) ([]Snapshot, error) {
	output, err := c.run(ctx, "list", "-H", "-p", "-t", "snapshot", "-o", "name,creation", "-s", "creation", "-r", dataset)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(output), "\n")
	prefix := dataset + "@"
	snapshots := make([]Snapshot, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 2 || !strings.HasPrefix(parts[0], prefix) {
			continue
		}
		createdUnix, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return nil, model.Wrap(model.ErrInternal, err, "parse zfs creation time")
		}
		snapshots = append(snapshots, Snapshot{
			Name:      strings.TrimPrefix(parts[0], prefix),
			Dataset:   dataset,
			CreatedAt: time.Unix(createdUnix, 0).UTC(),
		})
	}
	return snapshots, nil
}

func (c *Client) run(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, c.bin, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}
		return "", model.Wrap(model.ErrInternal, err, "zfs %s failed: %s", strings.Join(args, " "), message)
	}
	return strings.TrimSpace(stdout.String()), nil
}

func SnapshotRef(dataset, name string) string {
	return fmt.Sprintf("%s@%s", dataset, name)
}
