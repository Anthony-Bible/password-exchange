package keymanager

import (
	"bytes"
	"context"
	"log"
	"net"

	"github.com/Anthony-Bible/password-exchange/app/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/tink/go/aead"
	"github.com/google/tink/go/keyset"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	// import other necessary packages
)

type Server struct {
	//...
}
type Config struct {
	db.UnimplementedDbServiceServer
	PassConfig config.PassConfig `mapstructure:",squash"`
}

// StartServer starts the grpc server
func (conf Config) startServer() {

	address := "0.0.0.0:50051"
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal().Err(err).Msg("Problem with starting grpc server")
	}

	s := grpc.NewServer()
	srv := Config{
		PassConfig: conf.PassConfig,
	}
	db.RegisterDbServiceServer(s, &srv)
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatal().Msgf("failed to serve: %v", err)
	}
}

func (s *Server) CreateKey(ctx context.Context, req *CreateKeyRequest) (*CreateKeyResponse, error) {
	// 1. Generate a new key using Tink
	kh, err := keyset.NewHandle(aead.AES256GCMKeyTemplate())
	if err != nil {
		// handle error
	}

	// 2. Serialize key to bytes
	buf := new(bytes.Buffer)
	w := keyset.NewBinaryWriter(buf)
	if err := kh.Write(w, nil); err != nil {
		// handle error
	}
	key := buf.Bytes()

	// 3. Encrypt the key using the master password
	encryptedKey, err := encryptMasterPassword(key, req.MasterPassword)
	if err != nil {
		// handle error
	}

	// 4. Save the encrypted key to an S3-compatible bucket
	err = putToS3(req.KeyId, encryptedKey)
	if err != nil {
		// handle error
	}

	// 5. Return response
	return &CreateKeyResponse{
		KeyId: req.KeyId,
	}, nil
}

func putToS3(keyId string, encryptedKey []byte) error {
	// Define your AWS session configuration here
	s, err := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
		// ... more configuration
	})
	if err != nil {
		// handle error
	}

	svc := s3.New(s)
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String("your-bucket"),
		Key:    aws.String(keyId),
		Body:   bytes.NewReader(encryptedKey),
	})
	return err
}
