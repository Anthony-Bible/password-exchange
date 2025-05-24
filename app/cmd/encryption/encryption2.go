package encryption

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"net"

	"github.com/Anthony-Bible/password-exchange/app/internal/shared/config"
	encryptionDomain "github.com/Anthony-Bible/password-exchange/app/internal/domains/encryption/domain"
	encryptionGRPC "github.com/Anthony-Bible/password-exchange/app/internal/domains/encryption/adapters/primary/grpc"
	memoryKeygen "github.com/Anthony-Bible/password-exchange/app/internal/domains/encryption/adapters/secondary/memory"
	"github.com/go-kit/kit/transport/amqp"
	"github.com/rs/xid"
	"github.com/rs/zerolog/log"

	// "password.exchange/message"
	// b "password.exchange/aws"
	pb "github.com/Anthony-Bible/password-exchange/app/pkg/pb/encryption"
	"google.golang.org/grpc/reflection"

	"google.golang.org/grpc"
)

type Config struct {
	config.PassConfig `mapstructure:",squash"`
	Channel           *amqp.Channel
}

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
type server struct {
	pb.UnimplementedMessageServiceServer
}

func GenerateRandomBytes(n int32) *[32]byte {
	key := [32]byte{}
	_, err := io.ReadFull(rand.Reader, key[:])
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("there's not enough randomness")
	}
	return &key
}

// GenerateRandomString returns a URL-safe, base64 encoded
// securely generated random string.
// func GenerateRandomString(s int) (*[32]byte, string) {

// }

func Generateid() string {
	guid := xid.New()
	return guid.String()
}

// func (*server) MessageEncrypt(plaintext []byte, key *[32]byte) (ciphertext string) {

// }

func (*server) DecryptMessage(ctx context.Context, request *pb.DecryptedMessageRequest) (*pb.DecryptedMessageResponse, error) {
	key := request.GetKey()
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	CipherText := request.GetCiphertext()
	response := &pb.DecryptedMessageResponse{}
	for i := range CipherText {
		decodecCipher, err := base64.URLEncoding.DecodeString(CipherText[i])
		if err != nil {
			log.Error().Err(err).Msg("Something went wrong with decoding ciphertext")
			return nil, err
		}

		ciphertext := []byte(decodecCipher)
		if len(ciphertext) < gcm.NonceSize() {
			log.Error().Err(err).Msg("Malformed Ciphertext")
			return nil, errors.New("malformed ciphertext")
		}
		plaintext, err := gcm.Open(nil,
			ciphertext[:gcm.NonceSize()],
			ciphertext[gcm.NonceSize():],
			nil,
		)
		if err != nil {
			return nil, err
		}

		response.Plaintext = append(response.Plaintext, string(base64.URLEncoding.EncodeToString(plaintext)))
	}
	return response, nil
}

func (*server) EncryptMessage(ctx context.Context, request *pb.EncryptedMessageRequest) (*pb.EncryptedMessageResponse, error) {
	key := []byte(request.GetKey())
	PlainText := request.GetPlainText()
	block, err := aes.NewCipher(key[:])
	if err != nil {
		log.Fatal().Err(err).Msg("something went wrong with NewCipher")
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Fatal().Err(err).Msg("something went wrong Creating new encryption key")
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		log.Fatal().Err(err).Msg("something went wrong with reading random")
		return nil, err
	}
	response := &pb.EncryptedMessageResponse{}
	for i := range PlainText {
		plaintext := PlainText[i]
		response.Ciphertext = append(response.Ciphertext, string(base64.URLEncoding.EncodeToString(gcm.Seal(nonce, nonce, []byte(plaintext), nil))))
	}
	return response, nil
}

func (*server) GenerateRandomString(ctx context.Context, request *pb.Randomrequest) (*pb.Randomresponse, error) {
	//todo add goroutines
	s := request.GetRandomLength()
	b := GenerateRandomBytes(s)
	return &pb.Randomresponse{Encryptionbytes: b[:], EncryptionString: base64.URLEncoding.EncodeToString((b[:]))}, nil
}

func (conf Config) startServer() {
	// Use hexagonal architecture
	conf.startHexagonalServer()
}

func (conf Config) startHexagonalServer() {
	address := "0.0.0.0:50051"
	
	// Create key generator (secondary adapter)
	keyGenerator := memoryKeygen.NewKeyGenerator()
	
	// Create encryption service (domain)
	encryptionService := encryptionDomain.NewEncryptionService(keyGenerator)
	
	// Create gRPC server (primary adapter)
	grpcServer := encryptionGRPC.NewGRPCServer(encryptionService, address)
	
	// Start the server
	log.Info().Str("address", address).Msg("Starting encryption service with hexagonal architecture")
	if err := grpcServer.Start(); err != nil {
		log.Fatal().Err(err).Msg("Failed to start hexagonal encryption server")
	}
}

// Legacy methods kept for backward compatibility
func (conf Config) startLegacyServer() {
	address := "0.0.0.0:50051"
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal().Err(err).Msg("Problem with starting grpc server")
	}
	s := grpc.NewServer()
	pb.RegisterMessageServiceServer(s, &server{})
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatal().Msgf("failed to serve: %v", err)
	}
}
