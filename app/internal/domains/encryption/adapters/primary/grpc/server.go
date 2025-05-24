package grpc

import (
	"context"
	"net"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/encryption/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/encryption/ports/primary"
	pb "github.com/Anthony-Bible/password-exchange/app/pkg/pb/encryption"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// GRPCServer implements the gRPC encryption service
type GRPCServer struct {
	pb.UnimplementedMessageServiceServer
	encryptionService primary.EncryptionServicePort
	address           string
}

// NewGRPCServer creates a new gRPC server for the encryption service
func NewGRPCServer(encryptionService primary.EncryptionServicePort, address string) *GRPCServer {
	return &GRPCServer{
		encryptionService: encryptionService,
		address:           address,
	}
}

// Start starts the gRPC server
func (s *GRPCServer) Start() error {
	lis, err := net.Listen("tcp", s.address)
	if err != nil {
		log.Error().Err(err).Str("address", s.address).Msg("Failed to listen on address")
		return err
	}

	grpcServer := grpc.NewServer()
	pb.RegisterMessageServiceServer(grpcServer, s)
	reflection.Register(grpcServer)

	log.Info().Str("address", s.address).Msg("Starting encryption gRPC server")
	if err := grpcServer.Serve(lis); err != nil {
		log.Error().Err(err).Msg("Failed to serve gRPC")
		return err
	}

	return nil
}

// EncryptMessage handles encryption requests
func (s *GRPCServer) EncryptMessage(ctx context.Context, request *pb.EncryptedMessageRequest) (*pb.EncryptedMessageResponse, error) {
	log.Debug().Int("plaintextCount", len(request.GetPlainText())).Msg("Received encryption request")

	domainRequest := domain.EncryptionRequest{
		Plaintext: request.GetPlainText(),
		Key:       request.GetKey(),
	}

	response, err := s.encryptionService.Encrypt(ctx, domainRequest)
	if err != nil {
		log.Error().Err(err).Msg("Encryption failed")
		return nil, err
	}

	pbResponse := &pb.EncryptedMessageResponse{
		Ciphertext: response.Ciphertext,
	}

	log.Debug().Int("ciphertextCount", len(response.Ciphertext)).Msg("Successfully encrypted messages")
	return pbResponse, nil
}

// DecryptMessage handles decryption requests
func (s *GRPCServer) DecryptMessage(ctx context.Context, request *pb.DecryptedMessageRequest) (*pb.DecryptedMessageResponse, error) {
	log.Debug().Int("ciphertextCount", len(request.GetCiphertext())).Msg("Received decryption request")

	domainRequest := domain.DecryptionRequest{
		Ciphertext: request.GetCiphertext(),
		Key:        request.GetKey(),
	}

	response, err := s.encryptionService.Decrypt(ctx, domainRequest)
	if err != nil {
		log.Error().Err(err).Msg("Decryption failed")
		return nil, err
	}

	pbResponse := &pb.DecryptedMessageResponse{
		Plaintext: response.Plaintext,
	}

	log.Debug().Int("plaintextCount", len(response.Plaintext)).Msg("Successfully decrypted messages")
	return pbResponse, nil
}

// GenerateRandomString handles random key generation requests
func (s *GRPCServer) GenerateRandomString(ctx context.Context, request *pb.Randomrequest) (*pb.Randomresponse, error) {
	log.Debug().Int32("length", request.GetRandomLength()).Msg("Received random key generation request")

	domainRequest := domain.RandomRequest{
		Length: request.GetRandomLength(),
	}

	response, err := s.encryptionService.GenerateRandomKey(ctx, domainRequest)
	if err != nil {
		log.Error().Err(err).Msg("Random key generation failed")
		return nil, err
	}

	pbResponse := &pb.Randomresponse{
		Encryptionbytes:  response.Key.Bytes(),
		EncryptionString: response.KeyString,
	}

	log.Debug().Msg("Successfully generated random key")
	return pbResponse, nil
}