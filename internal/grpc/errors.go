package grpc

import (
	"errors"

	"github.com/al-tokarev/shortener/internal/repository/urlrepository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// mapError превращает доменную ошибку в gRPC-статус.
func mapError(err error) error {
	switch {
	case errors.Is(err, urlrepository.ErrURLNotFound):
		return status.Error(codes.NotFound, "not found")
	case errors.Is(err, urlrepository.ErrOriginalURLAlreadyExists):
		return status.Error(codes.AlreadyExists, "already exists")
	case errors.Is(err, urlrepository.ErrShortURLAlreadyExists):
		return status.Error(codes.InvalidArgument, "already exists")
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
