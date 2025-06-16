package grpc

import (
	"context"
	"net"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/ports/primary"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/logging/ports" // Add this
	database "github.com/Anthony-Bible/password-exchange/app/pkg/pb/database"
	"github.com/Anthony-Bible/password-exchange/app/pkg/validation"
	// "github.com/rs/zerolog/log" // Remove this
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
)

// GRPCServer adapts the storage service to gRPC protocol
type GRPCServer struct {
	database.UnimplementedDbServiceServer
	storageService primary.StorageServicePort
	address        string
	logger         ports.Logger // Add this
}

// NewGRPCServer creates a new gRPC server adapter
func NewGRPCServer(storageService primary.StorageServicePort, address string, logger ports.Logger) *GRPCServer {
	return &GRPCServer{
		storageService: storageService,
		address:        address,
		logger:         logger, // Add this
	}
}

// Insert handles gRPC insert requests by delegating to the storage service
func (s *GRPCServer) Insert(ctx context.Context, request *database.InsertRequest) (*emptypb.Empty, error) {
	message := &domain.Message{
		Content:        request.GetContent(),
		UniqueID:       request.GetUuid(),
		Passphrase:     request.GetPassphrase(),
		RecipientEmail: request.GetRecipientEmail(),
		MaxViewCount:   int(request.GetMaxViewCount()),
	}

	err := s.storageService.StoreMessage(ctx, message)
	if err != nil {
		s.logger.Error(ctx, "Failed to insert message via gRPC", "uuid", request.GetUuid(), "error", err)
		return nil, err
	}

	s.logger.Info(ctx, "Message inserted successfully via gRPC", "uuid", request.GetUuid(), "maxViewCount", request.GetMaxViewCount(), "recipientEmail", validation.SanitizeEmailForLogging(request.GetRecipientEmail()))
	return &emptypb.Empty{}, nil
}

// Select handles gRPC select requests by delegating to the storage service
func (s *GRPCServer) Select(ctx context.Context, request *database.SelectRequest) (*database.SelectResponse, error) {
	message, err := s.storageService.RetrieveMessage(ctx, request.GetUuid())
	if err != nil {
		s.logger.Error(ctx, "Failed to select message via gRPC", "uuid", request.GetUuid(), "error", err)
		return nil, err
	}

	response := &database.SelectResponse{
		Content:      message.Content,
		Passphrase:   message.Passphrase,
		ViewCount:    int32(message.ViewCount),
		MaxViewCount: int32(message.MaxViewCount),
	}

	s.logger.Info(ctx, "Message selected successfully via gRPC", "uuid", request.GetUuid(), "viewCount", message.ViewCount)
	return response, nil
}

// GetMessage handles gRPC select requests without incrementing view count
func (s *GRPCServer) GetMessage(ctx context.Context, request *database.SelectRequest) (*database.SelectResponse, error) {
	message, err := s.storageService.GetMessage(ctx, request.GetUuid())
	if err != nil {
		s.logger.Error(ctx, "Failed to select message without incrementing view count via gRPC", "uuid", request.GetUuid(), "error", err)
		return nil, err
	}

	response := &database.SelectResponse{
		Content:      message.Content,
		Passphrase:   message.Passphrase,
		ViewCount:    int32(message.ViewCount),
		MaxViewCount: int32(message.MaxViewCount),
	}

	s.logger.Info(ctx, "Message selected successfully without incrementing view count via gRPC", "uuid", request.GetUuid(), "viewCount", message.ViewCount)
	return response, nil
}

// GetUnviewedMessagesForReminders handles gRPC requests for unviewed messages eligible for reminders
func (s *GRPCServer) GetUnviewedMessagesForReminders(ctx context.Context, request *database.GetUnviewedMessagesRequest) (*database.GetUnviewedMessagesResponse, error) {
	messages, err := s.storageService.GetUnviewedMessagesForReminders(ctx, int(request.GetOlderThanHours()), int(request.GetMaxReminders()), int(request.GetReminderIntervalHours()))
	if err != nil {
		s.logger.Error(ctx, "Failed to get unviewed messages for reminders via gRPC", "error", err)
		return nil, err
	}

	var unviewedMessages []*database.UnviewedMessage
	for _, msg := range messages {
		unviewedMessages = append(unviewedMessages, &database.UnviewedMessage{
			MessageId:      int32(msg.MessageID),
			UniqueId:       msg.UniqueID,
			RecipientEmail: msg.RecipientEmail,
			Created:        msg.Created.Format("2006-01-02 15:04:05"),
			DaysOld:        int32(msg.DaysOld),
		})
	}

	s.logger.Info(ctx, "Retrieved unviewed messages for reminders via gRPC", "count", len(unviewedMessages))
	return &database.GetUnviewedMessagesResponse{Messages: unviewedMessages}, nil
}

// LogReminderSent handles gRPC requests to log reminder attempts
func (s *GRPCServer) LogReminderSent(ctx context.Context, request *database.LogReminderRequest) (*emptypb.Empty, error) {
	err := s.storageService.LogReminderSent(ctx, int(request.GetMessageId()), request.GetEmailAddress())
	if err != nil {
		s.logger.Error(ctx, "Failed to log reminder sent via gRPC", "messageID", request.GetMessageId(), "emailAddress", validation.SanitizeEmailForLogging(request.GetEmailAddress()), "error", err)
		return nil, err
	}

	s.logger.Info(ctx, "Reminder sent logged successfully via gRPC", "messageID", request.GetMessageId(), "emailAddress", validation.SanitizeEmailForLogging(request.GetEmailAddress()))
	return &emptypb.Empty{}, nil
}

// GetReminderHistory handles gRPC requests for reminder history
func (s *GRPCServer) GetReminderHistory(ctx context.Context, request *database.GetReminderHistoryRequest) (*database.GetReminderHistoryResponse, error) {
	history, err := s.storageService.GetReminderHistory(ctx, int(request.GetMessageId()))
	if err != nil {
		s.logger.Error(ctx, "Failed to get reminder history via gRPC", "messageID", request.GetMessageId(), "error", err)
		return nil, err
	}

	var entries []*database.ReminderLogEntry
	for _, entry := range history {
		entries = append(entries, &database.ReminderLogEntry{
			MessageId:        int32(entry.MessageID),
			EmailAddress:     entry.EmailAddress,
			ReminderCount:    int32(entry.ReminderCount),
			LastReminderSent: entry.LastReminderSent.Format("2006-01-02 15:04:05"),
		})
	}

	s.logger.Info(ctx, "Retrieved reminder history via gRPC", "messageID", request.GetMessageId(), "count", len(entries))
	return &database.GetReminderHistoryResponse{Entries: entries}, nil
}

// Start starts the gRPC server
func (s *GRPCServer) Start() error {
	lis, err := net.Listen("tcp", s.address)
	if err != nil {
		s.logger.Error(context.Background(), "Failed to listen on gRPC address", "address", s.address, "error", err)
		return err
	}

	grpcServer := grpc.NewServer()
	database.RegisterDbServiceServer(grpcServer, s)
	reflection.Register(grpcServer)

	s.logger.Info(context.Background(), "Starting gRPC storage server", "address", s.address)

	if err := grpcServer.Serve(lis); err != nil {
		s.logger.Error(context.Background(), "Failed to serve gRPC storage server", "error", err)
		return err
	}

	return nil
}
