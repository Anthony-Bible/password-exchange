package grpc_clients

import (
	"context"
	"testing"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/domain"
	database "github.com/Anthony-Bible/password-exchange/app/pkg/pb/database"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type storageDBStub struct {
	stubDBClientForUpload
	insertReq  *database.InsertRequest
	selectResp *database.SelectResponse
}

func (s *storageDBStub) Insert(
	_ context.Context,
	req *database.InsertRequest,
	_ ...grpc.CallOption,
) (*emptypb.Empty, error) {
	s.insertReq = req
	return &emptypb.Empty{}, nil
}

func (s *storageDBStub) Select(
	_ context.Context,
	_ *database.SelectRequest,
	_ ...grpc.CallOption,
) (*database.SelectResponse, error) {
	return s.selectResp, nil
}

func (s *storageDBStub) GetMessage(
	_ context.Context,
	_ *database.SelectRequest,
	_ ...grpc.CallOption,
) (*database.SelectResponse, error) {
	return s.selectResp, nil
}

func TestStorageClient_StoreMessage_PassesClientEncryptedFlagToProto(t *testing.T) {
	t.Parallel()

	stub := &storageDBStub{}
	client := &StorageClient{client: stub}

	req := domain.MessageStorageRequest{
		MessageID:         "msg-1",
		Content:           "ciphertext",
		IsClientEncrypted: true,
	}

	err := client.StoreMessage(context.Background(), req)
	if err != nil {
		t.Fatalf("StoreMessage returned error: %v", err)
	}

	if stub.insertReq == nil {
		t.Fatal("expected Insert to be called")
	}
	if !stub.insertReq.IsClientEncrypted {
		t.Fatal("expected IsClientEncrypted=true in proto request")
	}
}

func TestStorageClient_RetrieveMessage_MapsClientEncryptedFlagFromProto(t *testing.T) {
	t.Parallel()

	stub := &storageDBStub{
		selectResp: &database.SelectResponse{
			Uuid:              "msg-1",
			Content:           "ciphertext",
			IsClientEncrypted: true,
		},
	}

	client := &StorageClient{client: stub}
	got, err := client.RetrieveMessage(context.Background(), domain.MessageRetrievalStorageRequest{MessageID: "msg-1"})
	if err != nil {
		t.Fatalf("RetrieveMessage returned error: %v", err)
	}
	if !got.IsClientEncrypted {
		t.Fatal("expected IsClientEncrypted=true in storage response contract")
	}
}
