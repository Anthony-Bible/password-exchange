package grpc

import (
	"context"
	"net"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/encryption/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/encryption/ports/primary"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/logging/ports" // Add this
	pb "github.com/Anthony-Bible/password-exchange/app/pkg/pb/encryption"
	// "github.com/rs/zerolog/log" // Remove this
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// GRPCServer implements the gRPC encryption service
type GRPCServer struct {
	pb.UnimplementedMessageServiceServer
	encryptionService primary.EncryptionServicePort
	address           string
	logger            ports.Logger // Add this
}

// NewGRPCServer creates a new gRPC server for the encryption service
func NewGRPCServer(encryptionService primary.EncryptionServicePort, address string, logger ports.Logger) *GRPCServer {
	return &GRPCServer{
		encryptionService: encryptionService,
		address:           address,
		logger:            logger, // Add this
	}
}

// Start starts the gRPC server
func (s *GRPCServer) Start() error {
	lis, err := net.Listen("tcp", s.address)
	if err != nil {
		s.logger.Error(context.Background(), "Failed to listen on address", "address", s.address, "error", err)
		return err
	}

	grpcServer := grpc.NewServer()
	pb.RegisterMessageServiceServer(grpcServer, s)
	reflection.Register(grpcServer)

	s.logger.Info(context.Background(), "Starting encryption gRPC server", "address", s.address)
	if err := grpcServer.Serve(lis); err != nil {
		s.logger.Error(context.Background(), "Failed to serve gRPC", "error", err)
		return err
	}

	return nil
}

// EncryptMessage handles encryption requests
func (s *GRPCServer) EncryptMessage(ctx context.Context, request *pb.EncryptedMessageRequest) (*pb.EncryptedMessageResponse, error) {
	s.logger.Debug(ctx, "Received encryption request", "plaintextCount", len(request.GetPlainText()))

	domainRequest := domain.EncryptionRequest{
		Plaintext: request.GetPlainText(),
		Key:       request.GetKey(),
	}

	response, err := s.encryptionService.Encrypt(ctx, domainRequest)
	if err != nil {
		s.logger.Error(ctx, "Encryption failed", "error", err)
		return nil, err
	}

	pbResponse := &pb.EncryptedMessageResponse{
		Ciphertext: response.Ciphertext,
	}

	s.logger.Debug(ctx, "Successfully encrypted messages", "ciphertextCount", len(response.Ciphertext))
	return pbResponse, nil
}

// DecryptMessage handles decryption requests
func (s *GRPCServer) DecryptMessage(ctx context.Context, request *pb.DecryptedMessageRequest) (*pb.DecryptedMessageResponse, error) {
	s.logger.Debug(ctx, "Received decryption request", "ciphertextCount", len(request.GetCiphertext()))

	domainRequest := domain.DecryptionRequest{
		Ciphertext: request.GetCiphertext(),
		Key:        request.GetKey(),
	}

	response, err := s.encryptionService.Decrypt(ctx, domainRequest)
	if err != nil {
		s.logger.Error(ctx, "Decryption failed", "error", err)
		return nil, err
	}

	pbResponse := &pb.DecryptedMessageResponse{
		Plaintext: response.Plaintext,
	}

	s.logger.Debug(ctx, "Successfully decrypted messages", "plaintextCount", len(response.Plaintext))
	return pbResponse, nil
}

// GenerateRandomString handles random key generation requests
func (s *GRPCServer) GenerateRandomString(ctx context.Context, request *pb.Randomrequest) (*pb.Randomresponse, error) {
	s.logger.Debug(ctx, "Received random key generation request", "length", request.GetRandomLength())

	domainRequest := domain.RandomRequest{
		Length: request.GetRandomLength(),
	}

	response, err := s.encryptionService.GenerateRandomKey(ctx, domainRequest)
	if err != nil {
		s.logger.Error(ctx, "Random key generation failed", "error", err)
		return nil, err
	}

	pbResponse := &pb.Randomresponse{
		EncryptionBytes:  response.Key.Bytes(),
		EncryptionString: response.KeyString,
	}

	s.logger.Debug(ctx, "Successfully generated random key")
	return pbResponse, nil
}