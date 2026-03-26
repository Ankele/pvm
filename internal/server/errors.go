package server

import (
	"net/http"

	"github.com/ankele/pvm/internal/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func grpcError(err error) error {
	if err == nil {
		return nil
	}
	switch model.CodeOf(err) {
	case model.ErrInvalidArgument:
		return status.Error(codes.InvalidArgument, model.MessageOf(err))
	case model.ErrNotFound:
		return status.Error(codes.NotFound, model.MessageOf(err))
	case model.ErrUnsupported:
		return status.Error(codes.Unimplemented, model.MessageOf(err))
	case model.ErrConflict:
		return status.Error(codes.AlreadyExists, model.MessageOf(err))
	case model.ErrPrecondition:
		return status.Error(codes.FailedPrecondition, model.MessageOf(err))
	case model.ErrUnauthenticated:
		return status.Error(codes.Unauthenticated, model.MessageOf(err))
	case model.ErrUnavailable:
		return status.Error(codes.Unavailable, model.MessageOf(err))
	default:
		return status.Error(codes.Internal, model.MessageOf(err))
	}
}

func httpStatus(err error) int {
	switch model.CodeOf(err) {
	case model.ErrInvalidArgument:
		return http.StatusBadRequest
	case model.ErrNotFound:
		return http.StatusNotFound
	case model.ErrUnsupported:
		return http.StatusNotImplemented
	case model.ErrConflict:
		return http.StatusConflict
	case model.ErrPrecondition:
		return http.StatusPreconditionFailed
	case model.ErrUnauthenticated:
		return http.StatusUnauthorized
	case model.ErrUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}
