package grpc

import (
	"context"
	"strings"

	"github.com/al-tokarev/shortener/internal/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor проверяет metadata authorization и кладёт userID в ctx.
// Ожидает: authorization: <userID>|<signature>
// (то же значение, что в cookie user_id).
func AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "no metadata")
		}

		values := md.Get("authorization")
		if len(values) == 0 {
			return nil, status.Error(codes.Unauthenticated, "no authorization header")
		}

		// На случай, если клиент оборачивает в "Bearer " — снимаем
		raw := strings.TrimPrefix(values[0], "Bearer ")

		userID, err := auth.ValidateCookieValue(raw)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization")
		}

		return handler(withUserID(ctx, userID), req)
	}
}
