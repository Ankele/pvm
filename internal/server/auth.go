package server

import (
	"context"
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/ankele/pvm/internal/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type TokenAuth struct {
	token string
}

func NewTokenAuth(token string) *TokenAuth {
	return &TokenAuth{token: strings.TrimSpace(token)}
}

func (a *TokenAuth) ValidateHeader(value string) error {
	if strings.TrimSpace(a.token) == "" {
		return model.Errorf(model.ErrUnauthenticated, "api token is not configured")
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(value, prefix) {
		return model.Errorf(model.ErrUnauthenticated, "authorization header must use Bearer token")
	}
	if subtle.ConstantTimeCompare([]byte(strings.TrimSpace(value[len(prefix):])), []byte(a.token)) != 1 {
		return model.Errorf(model.ErrUnauthenticated, "invalid api token")
	}
	return nil
}

func (a *TokenAuth) UnaryInterceptor(exempt map[string]struct{}) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if _, ok := exempt[info.FullMethod]; ok {
			return handler(ctx, req)
		}
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, grpcError(model.Errorf(model.ErrUnauthenticated, "missing metadata"))
		}
		values := md.Get("authorization")
		if len(values) == 0 {
			return nil, grpcError(model.Errorf(model.ErrUnauthenticated, "missing authorization metadata"))
		}
		if err := a.ValidateHeader(values[0]); err != nil {
			return nil, grpcError(err)
		}
		return handler(ctx, req)
	}
}

func (a *TokenAuth) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}
		if err := a.ValidateHeader(r.Header.Get("Authorization")); err != nil {
			http.Error(w, model.MessageOf(err), httpStatus(err))
			return
		}
		next.ServeHTTP(w, r)
	})
}
