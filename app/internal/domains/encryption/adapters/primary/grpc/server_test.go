package grpc

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/encryption/ports/contracts"
	pb "github.com/Anthony-Bible/password-exchange/app/pkg/pb/encryption"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

// --- Stub LoggerPort ---

type stubLogEvent struct{}

func (stubLogEvent) Err(error) contracts.LogEvent             { return stubLogEvent{} }
func (stubLogEvent) Str(string, string) contracts.LogEvent    { return stubLogEvent{} }
func (stubLogEvent) Int(string, int) contracts.LogEvent       { return stubLogEvent{} }
func (stubLogEvent) Int32(string, int32) contracts.LogEvent   { return stubLogEvent{} }
func (stubLogEvent) Bool(string, bool) contracts.LogEvent     { return stubLogEvent{} }
func (stubLogEvent) Dur(string, time.Duration) contracts.LogEvent { return stubLogEvent{} }
func (stubLogEvent) Float64(string, float64) contracts.LogEvent   { return stubLogEvent{} }
func (stubLogEvent) Msg(string)                                {}

type stubLogger struct{}

func (stubLogger) Debug() contracts.LogEvent { return stubLogEvent{} }
func (stubLogger) Info() contracts.LogEvent  { return stubLogEvent{} }
func (stubLogger) Warn() contracts.LogEvent  { return stubLogEvent{} }
func (stubLogger) Error() contracts.LogEvent { return stubLogEvent{} }
func (stubLogger) Fatal() contracts.LogEvent { return stubLogEvent{} }

// --- Stub EncryptionServicePort ---

type stubService struct {
	encryptResp *contracts.EncryptionResponse
	encryptErr  error
	decryptResp *contracts.DecryptionResponse
	decryptErr  error
	randomResp  *contracts.RandomResponse
	randomErr   error

	lastEncryptReq contracts.EncryptionRequest
	lastDecryptReq contracts.DecryptionRequest
	lastRandomReq  contracts.RandomRequest
}

func (s *stubService) Encrypt(_ context.Context, req contracts.EncryptionRequest) (*contracts.EncryptionResponse, error) {
	s.lastEncryptReq = req
	return s.encryptResp, s.encryptErr
}

func (s *stubService) Decrypt(_ context.Context, req contracts.DecryptionRequest) (*contracts.DecryptionResponse, error) {
	s.lastDecryptReq = req
	return s.decryptResp, s.decryptErr
}

func (s *stubService) GenerateRandomKey(_ context.Context, req contracts.RandomRequest) (*contracts.RandomResponse, error) {
	s.lastRandomReq = req
	return s.randomResp, s.randomErr
}

func (s *stubService) GenerateID(_ context.Context) string { return "stub-id" }

// --- Test harness ---

func startTestServer(t *testing.T, svc *stubService) (pb.MessageServiceClient, func()) {
	t.Helper()
	lis := bufconn.Listen(1024 * 1024)
	srv := grpc.NewServer()
	adapter := NewGRPCServer(svc, "", stubLogger{})
	pb.RegisterMessageServiceServer(srv, adapter)

	serveErrCh := make(chan error, 1)
	go func() {
		serveErrCh <- srv.Serve(lis)
	}()

	conn, err := grpc.NewClient(
		"passthrough://bufnet",
		grpc.WithContextDialer(func(_ context.Context, _ string) (net.Conn, error) { return lis.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	cleanup := func() {
		_ = conn.Close()
		srv.Stop()
		_ = lis.Close()
		if serveErr := <-serveErrCh; serveErr != nil && !errors.Is(serveErr, grpc.ErrServerStopped) {
			t.Errorf("gRPC Serve returned unexpected error: %v", serveErr)
		}
	}
	return pb.NewMessageServiceClient(conn), cleanup
}

func TestEncryptMessage_Success(t *testing.T) {
	svc := &stubService{
		encryptResp: &contracts.EncryptionResponse{Ciphertext: []string{"cipher1", "cipher2"}},
	}
	client, cleanup := startTestServer(t, svc)
	defer cleanup()

	resp, err := client.EncryptMessage(context.Background(), &pb.EncryptedMessageRequest{
		PlainText: []string{"hello", "world"},
		Key:       []byte("key-bytes"),
	})

	require.NoError(t, err)
	assert.Equal(t, []string{"cipher1", "cipher2"}, resp.GetCiphertext())
	assert.Equal(t, []string{"hello", "world"}, svc.lastEncryptReq.Plaintext)
	assert.Equal(t, []byte("key-bytes"), svc.lastEncryptReq.Key)
}

func TestEncryptMessage_Error(t *testing.T) {
	svc := &stubService{encryptErr: errors.New("boom")}
	client, cleanup := startTestServer(t, svc)
	defer cleanup()

	_, err := client.EncryptMessage(context.Background(), &pb.EncryptedMessageRequest{
		PlainText: []string{"hi"},
	})
	assert.Error(t, err)
}

func TestDecryptMessage_Success(t *testing.T) {
	svc := &stubService{
		decryptResp: &contracts.DecryptionResponse{Plaintext: []string{"plain1"}},
	}
	client, cleanup := startTestServer(t, svc)
	defer cleanup()

	resp, err := client.DecryptMessage(context.Background(), &pb.DecryptedMessageRequest{
		Ciphertext: []string{"cipher1"},
		Key:        []byte("k"),
	})

	require.NoError(t, err)
	assert.Equal(t, []string{"plain1"}, resp.GetPlaintext())
	assert.Equal(t, []string{"cipher1"}, svc.lastDecryptReq.Ciphertext)
	assert.Equal(t, []byte("k"), svc.lastDecryptReq.Key)
}

func TestDecryptMessage_Error(t *testing.T) {
	svc := &stubService{decryptErr: errors.New("nope")}
	client, cleanup := startTestServer(t, svc)
	defer cleanup()

	_, err := client.DecryptMessage(context.Background(), &pb.DecryptedMessageRequest{
		Ciphertext: []string{"c"},
	})
	assert.Error(t, err)
}

func TestGenerateRandomString_Success(t *testing.T) {
	var key contracts.EncryptionKey
	for i := range key {
		key[i] = byte(i)
	}
	svc := &stubService{
		randomResp: &contracts.RandomResponse{Key: key, KeyString: "encoded-key"},
	}
	client, cleanup := startTestServer(t, svc)
	defer cleanup()

	resp, err := client.GenerateRandomString(context.Background(), &pb.Randomrequest{RandomLength: 32})

	require.NoError(t, err)
	assert.Equal(t, key.Bytes(), resp.GetEncryptionBytes())
	assert.Equal(t, "encoded-key", resp.GetEncryptionString())
	assert.Equal(t, int32(32), svc.lastRandomReq.Length)
}

func TestGenerateRandomString_Error(t *testing.T) {
	svc := &stubService{randomErr: errors.New("rng failure")}
	client, cleanup := startTestServer(t, svc)
	defer cleanup()

	_, err := client.GenerateRandomString(context.Background(), &pb.Randomrequest{RandomLength: 16})
	assert.Error(t, err)
}
