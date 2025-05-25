package grpc

import (
	"context"
	"net"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/ports/primary"
	db "github.com/Anthony-Bible/password-exchange/app/pkg/pb/database"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
)

// GRPCServer adapts the storage service to gRPC protocol
type GRPCServer struct {
	db.UnimplementedDbServiceServer
	storageService primary.StorageServicePort
	address        string
}

// NewGRPCServer creates a new gRPC server adapter
func NewGRPCServer(storageService primary.StorageServicePort, address string) *GRPCServer {
	return &GRPCServer{
		storageService: storageService,
		address:        address,
	}
}

// Insert handles gRPC insert requests by delegating to the storage service
func (s *GRPCServer) Insert(ctx context.Context, request *db.InsertRequest) (*emptypb.Empty, error) {
	err := s.storageService.StoreMessage(ctx, request.GetContent(), request.GetUuid(), request.GetPassphrase())
	if err != nil {
		log.Error().Err(err).Str("uuid", request.GetUuid()).Msg("Failed to insert message via gRPC")
		return nil, err
	}

	log.Info().Str("uuid", request.GetUuid()).Msg("Message inserted successfully via gRPC")
	return &emptypb.Empty{}, nil
}

// Select handles gRPC select requests by delegating to the storage service
func (s *GRPCServer) Select(ctx context.Context, request *db.SelectRequest) (*db.SelectResponse, error) {
	message, err := s.storageService.RetrieveMessage(ctx, request.GetUuid())
	if err != nil {
		log.Error().Err(err).Str("uuid", request.GetUuid()).Msg("Failed to select message via gRPC")
		return nil, err
	}

	response := &db.SelectResponse{
		Content:    message.Content,
		Passphrase: message.Passphrase,
		ViewCount:  int32(message.ViewCount),
	}

	log.Info().Str("uuid", request.GetUuid()).Int("viewCount", message.ViewCount).Msg("Message selected successfully via gRPC")
	return response, nil
}

// GetMessage handles gRPC select requests without incrementing view count
func (s *GRPCServer) GetMessage(ctx context.Context, request *db.SelectRequest) (*db.SelectResponse, error) {
	message, err := s.storageService.GetMessage(ctx, request.GetUuid())
	if err != nil {
		log.Error().Err(err).Str("uuid", request.GetUuid()).Msg("Failed to select message without incrementing view count via gRPC")
		return nil, err
	}

	response := &db.SelectResponse{
		Content:    message.Content,
		Passphrase: message.Passphrase,
		ViewCount:  int32(message.ViewCount),
	}

	log.Info().Str("uuid", request.GetUuid()).Int("viewCount", message.ViewCount).Msg("Message selected successfully without incrementing view count via gRPC")
	return response, nil
}

// Start starts the gRPC server
func (s *GRPCServer) Start() error {
	lis, err := net.Listen("tcp", s.address)
	if err != nil {
		log.Fatal().Err(err).Str("address", s.address).Msg("Failed to listen on gRPC address")
		return err
	}

	grpcServer := grpc.NewServer()
	db.RegisterDbServiceServer(grpcServer, s)
	reflection.Register(grpcServer)

	log.Info().Str("address", s.address).Msg("Starting gRPC storage server")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal().Err(err).Msg("Failed to serve gRPC storage server")
		return err
	}

	return nil
}
