package grpc

import (
	"context"

	pb "github.com/al-tokarev/shortener/internal/grpc/grpcgen"
	"github.com/al-tokarev/shortener/internal/service/urlservices"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ShortenerServer struct {
	pb.ShortenerServiceServer
	service urlservices.URLServiceInterface
	logger  *zap.SugaredLogger
}

func NewServer(s urlservices.URLServiceInterface, l *zap.SugaredLogger) *ShortenerServer {
	return &ShortenerServer{
		service: s,
		logger:  l,
	}
}

func (s *ShortenerServer) ShortenURL(ctx context.Context, req *pb.URLShortenRequest) (*pb.URLShortenResponse, error) {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	id, err := s.service.SetURL(req.GetUrl(), userID)
	if err != nil {
		s.logger.Errorw("grpc shorten failed", "error", err)
		return nil, mapError(err)
	}

	return &pb.URLShortenResponse{Result: &id.ShortURL}, nil
}

func (s *ShortenerServer) ExpandURL(ctx context.Context, req *pb.URLExpandRequest) (*pb.URLExpandResponse, error) {
	url, err := s.service.GetFullURL(req.GetId())
	if err != nil {
		s.logger.Errorw("grpc expand failed", "error", err)
		return nil, mapError(err)
	}

	return &pb.URLExpandResponse{Result: &url}, nil
}

func (s *ShortenerServer) ListUserURLs(ctx context.Context, _ *emptypb.Empty) (*pb.UserURLsResponse, error) {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	urls, err := s.service.GetUserURLs(userID)
	if err != nil {
		s.logger.Errorw("grpc list failed", "error", err)
		return nil, mapError(err)
	}

	resp := &pb.UserURLsResponse{
		Url: make([]*pb.URLData, 0, len(*urls)),
	}
	for _, u := range *urls {
		resp.Url = append(resp.Url, &pb.URLData{
			ShortUrl:    &u.ShortURL,
			OriginalUrl: &u.OriginalURL,
		})
	}
	return resp, nil
}
