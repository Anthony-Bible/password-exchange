package grpc

import (
	"context"
	"net"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/domain"
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
	message := &domain.Message{
		Content:        request.GetContent(),
		UniqueID:       request.GetUuid(),
		Passphrase:     request.GetPassphrase(),
		RecipientEmail: request.GetRecipientEmail(),
		MaxViewCount:   int(request.GetMaxViewCount()),
	}

	err := s.storageService.StoreMessage(ctx, message)
	if err != nil {
		log.Error().Err(err).Str("uuid", request.GetUuid()).Msg("Failed to insert message via gRPC")
		return nil, err
	}

	log.Info().Str("uuid", request.GetUuid()).Int32("maxViewCount", request.GetMaxViewCount()).Str("recipientEmail", request.GetRecipientEmail()).Msg("Message inserted successfully via gRPC")
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
		Content:      message.Content,
		Passphrase:   message.Passphrase,
		ViewCount:    int32(message.ViewCount),
		MaxViewCount: int32(message.MaxViewCount),
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
		Content:      message.Content,
		Passphrase:   message.Passphrase,
		ViewCount:    int32(message.ViewCount),
		MaxViewCount: int32(message.MaxViewCount),
	}

	log.Info().Str("uuid", request.GetUuid()).Int("viewCount", message.ViewCount).Msg("Message selected successfully without incrementing view count via gRPC")
	return response, nil
}

// GetUnviewedMessagesForReminders handles gRPC requests for unviewed messages eligible for reminders
func (s *GRPCServer) GetUnviewedMessagesForReminders(ctx context.Context, request *db.GetUnviewedMessagesRequest) (*db.GetUnviewedMessagesResponse, error) {
	messages, err := s.storageService.GetUnviewedMessagesForReminders(ctx, int(request.GetOlderThanHours()), int(request.GetMaxReminders()))
	if err != nil {
		log.Error().Err(err).Msg("Failed to get unviewed messages for reminders via gRPC")
		return nil, err
	}

	var unviewedMessages []*db.UnviewedMessage
	for _, msg := range messages {
		unviewedMessages = append(unviewedMessages, &db.UnviewedMessage{
			MessageId:      int32(msg.MessageID),
			UniqueId:       msg.UniqueID,
			RecipientEmail: msg.RecipientEmail,
			Created:        msg.Created.Format("2006-01-02 15:04:05"),
			DaysOld:        int32(msg.DaysOld),
		})
	}

	log.Info().Int("count", len(unviewedMessages)).Msg("Retrieved unviewed messages for reminders via gRPC")
	return &db.GetUnviewedMessagesResponse{Messages: unviewedMessages}, nil
}

// LogReminderSent handles gRPC requests to log reminder attempts
func (s *GRPCServer) LogReminderSent(ctx context.Context, request *db.LogReminderRequest) (*emptypb.Empty, error) {
	err := s.storageService.LogReminderSent(ctx, int(request.GetMessageId()), request.GetEmailAddress())
	if err != nil {
		log.Error().Err(err).Int32("messageID", request.GetMessageId()).Str("emailAddress", request.GetEmailAddress()).Msg("Failed to log reminder sent via gRPC")
		return nil, err
	}

	log.Info().Int32("messageID", request.GetMessageId()).Str("emailAddress", request.GetEmailAddress()).Msg("Reminder sent logged successfully via gRPC")
	return &emptypb.Empty{}, nil
}

// GetReminderHistory handles gRPC requests for reminder history
func (s *GRPCServer) GetReminderHistory(ctx context.Context, request *db.GetReminderHistoryRequest) (*db.GetReminderHistoryResponse, error) {
	history, err := s.storageService.GetReminderHistory(ctx, int(request.GetMessageId()))
	if err != nil {
		log.Error().Err(err).Int32("messageID", request.GetMessageId()).Msg("Failed to get reminder history via gRPC")
		return nil, err
	}

	var entries []*db.ReminderLogEntry
	for _, entry := range history {
		entries = append(entries, &db.ReminderLogEntry{
			MessageId:        int32(entry.MessageID),
			EmailAddress:     entry.EmailAddress,
			ReminderCount:    int32(entry.ReminderCount),
			LastReminderSent: entry.LastReminderSent.Format("2006-01-02 15:04:05"),
		})
	}

	log.Info().Int32("messageID", request.GetMessageId()).Int("count", len(entries)).Msg("Retrieved reminder history via gRPC")
	return &db.GetReminderHistoryResponse{Entries: entries}, nil
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
