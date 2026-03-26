//go:build !linux || !cgo

package backend

import "context"

func New(ctx context.Context, opts Options) (Backend, error) {
	return NewUnsupported(ctx, opts)
}
