package grpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/ports/contracts"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/ports/primary"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/ports/secondary"
	database "github.com/Anthony-Bible/password-exchange/app/pkg/pb/database"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// maxExpirationDuration is the maximum allowed expiration time (90 days).
const maxExpirationDuration = 2160 * time.Hour

// domainErrorToStatus maps storage-domain sentinel errors to typed gRPC
// status codes. Without this, validation failures bubble up as codes.Unknown
// and clients that retry on Unknown (a common default) loop forever on
// deterministic input errors.
func domainErrorToStatus(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, domain.ErrNilMessage),
		errors.Is(err, domain.ErrEmptyContent),
		errors.Is(err, domain.ErrEmptyUniqueID),
		errors.Is(err, domain.ErrEmptyEmailAddress),
		errors.Is(err, domain.ErrInvalidParameter),
		errors.Is(err, domain.ErrInvalidMaxViewCount):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrMessageNotFound):
		return status.Error(codes.NotFound, err.Error())
	}
	return err
}

// GRPCServer adapts the storage service to gRPC protocol.
//
// All logging and email-sanitization flow through injected secondary ports so
// the adapter stays decoupled from the shared logging/validation packages and
// can be exercised in tests without touching globals.
type GRPCServer struct {
	database.UnimplementedDbServiceServer
	storageService primary.StorageServicePort
	logger         secondary.LoggerPort
	validator      secondary.ValidationPort
	address        string
}

// NewGRPCServer creates a new gRPC server adapter wired with the storage
// service plus the logger and validator secondary ports.
func NewGRPCServer(
	storageService primary.StorageServicePort,
	address string,
	logger secondary.LoggerPort,
	validator secondary.ValidationPort,
) *GRPCServer {
	if storageService == nil {
		panic("storage/grpc: NewGRPCServer requires a non-nil StorageServicePort")
	}
	if logger == nil {
		panic("storage/grpc: NewGRPCServer requires a non-nil LoggerPort")
	}
	if validator == nil {
		panic("storage/grpc: NewGRPCServer requires a non-nil ValidationPort")
	}
	return &GRPCServer{
		storageService: storageService,
		logger:         logger,
		validator:      validator,
		address:        address,
	}
}

// Insert handles gRPC insert requests by delegating to the storage service
func (s *GRPCServer) Insert(ctx context.Context, request *database.InsertRequest) (*emptypb.Empty, error) {
	expiresAt, err := parseExpiresAt(request.GetExpiresAt())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid expires_at: %v", err)
	}
	if expiresAt != nil {
		remaining := time.Until(*expiresAt)
		if remaining <= 0 {
			return nil, status.Errorf(codes.InvalidArgument, "expires_at must be in the future")
		}
		if remaining > maxExpirationDuration {
			return nil, status.Errorf(codes.InvalidArgument, "expires_at exceeds maximum allowed expiration")
		}
	}

	message := &contracts.Message{
		Content:        request.GetContent(),
		UniqueID:       request.GetUuid(),
		Passphrase:     request.GetPassphrase(),
		RecipientEmail: request.GetRecipientEmail(),
		MaxViewCount:   int(request.GetMaxViewCount()),
		ExpiresAt:      expiresAt,
	}

	err = s.storageService.StoreMessage(ctx, message)
	if err != nil {
		s.logger.Error().Err(err).Str("uuid", request.GetUuid()).Msg("Failed to insert message via gRPC")
		return nil, domainErrorToStatus(err)
	}

	s.logger.Info().
		Str("uuid", request.GetUuid()).
		Int32("maxViewCount", request.GetMaxViewCount()).
		Str("recipientEmail", s.validator.SanitizeEmailForLogging(request.GetRecipientEmail())).
		Msg("Message inserted successfully via gRPC")
	return &emptypb.Empty{}, nil
}

// parseExpiresAt parses an RFC3339 timestamp string into a *time.Time.
// Returns nil, nil for empty string (use default TTL).
// Returns nil, err for non-empty but un-parseable strings.
func parseExpiresAt(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil, fmt.Errorf("parse RFC3339: %w", err)
	}
	return &t, nil
}

// formatTime formats a *time.Time as RFC3339, returning empty string for nil.
func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

// Select handles gRPC select requests by delegating to the storage service
func (s *GRPCServer) Select(ctx context.Context, request *database.SelectRequest) (*database.SelectResponse, error) {
	message, err := s.storageService.RetrieveMessage(ctx, request.GetUuid())
	if err != nil {
		s.logger.Error().Err(err).Str("uuid", request.GetUuid()).Msg("Failed to select message via gRPC")
		return nil, err
	}

	response := &database.SelectResponse{
		Content:      message.Content,
		Passphrase:   message.Passphrase,
		ViewCount:    int32(message.ViewCount),
		MaxViewCount: int32(message.MaxViewCount),
		ExpiresAt:    formatTime(message.ExpiresAt),
	}

	s.logger.Info().
		Str("uuid", request.GetUuid()).
		Int("viewCount", message.ViewCount).
		Msg("Message selected successfully via gRPC")
	return response, nil
}

// GetMessage handles gRPC select requests without incrementing view count
func (s *GRPCServer) GetMessage(
	ctx context.Context,
	request *database.SelectRequest,
) (*database.SelectResponse, error) {
	message, err := s.storageService.GetMessage(ctx, request.GetUuid())
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("uuid", request.GetUuid()).
			Msg("Failed to select message without incrementing view count via gRPC")
		return nil, err
	}

	response := &database.SelectResponse{
		Content:      message.Content,
		Passphrase:   message.Passphrase,
		ViewCount:    int32(message.ViewCount),
		MaxViewCount: int32(message.MaxViewCount),
		ExpiresAt:    formatTime(message.ExpiresAt),
	}

	s.logger.Info().
		Str("uuid", request.GetUuid()).
		Int("viewCount", message.ViewCount).
		Msg("Message selected successfully without incrementing view count via gRPC")
	return response, nil
}

// GetUnviewedMessagesForReminders handles gRPC requests for unviewed messages eligible for reminders
func (s *GRPCServer) GetUnviewedMessagesForReminders(
	ctx context.Context,
	request *database.GetUnviewedMessagesRequest,
) (*database.GetUnviewedMessagesResponse, error) {
	messages, err := s.storageService.GetUnviewedMessagesForReminders(
		ctx,
		int(request.GetOlderThanHours()),
		int(request.GetMaxReminders()),
		int(request.GetReminderIntervalHours()),
	)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get unviewed messages for reminders via gRPC")
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

	s.logger.Info().Int("count", len(unviewedMessages)).Msg("Retrieved unviewed messages for reminders via gRPC")
	return &database.GetUnviewedMessagesResponse{Messages: unviewedMessages}, nil
}

// LogReminderSent handles gRPC requests to log reminder attempts
func (s *GRPCServer) LogReminderSent(
	ctx context.Context,
	request *database.LogReminderRequest,
) (*emptypb.Empty, error) {
	err := s.storageService.LogReminderSent(ctx, int(request.GetMessageId()), request.GetEmailAddress())
	if err != nil {
		s.logger.Error().
			Err(err).
			Int32("messageID", request.GetMessageId()).
			Str("emailAddress", s.validator.SanitizeEmailForLogging(request.GetEmailAddress())).
			Msg("Failed to log reminder sent via gRPC")
		return nil, err
	}

	s.logger.Info().
		Int32("messageID", request.GetMessageId()).
		Str("emailAddress", s.validator.SanitizeEmailForLogging(request.GetEmailAddress())).
		Msg("Reminder sent logged successfully via gRPC")
	return &emptypb.Empty{}, nil
}

// GetReminderHistory handles gRPC requests for reminder history
func (s *GRPCServer) GetReminderHistory(
	ctx context.Context,
	request *database.GetReminderHistoryRequest,
) (*database.GetReminderHistoryResponse, error) {
	history, err := s.storageService.GetReminderHistory(ctx, int(request.GetMessageId()))
	if err != nil {
		s.logger.Error().Err(err).Int32("messageID", request.GetMessageId()).Msg("Failed to get reminder history via gRPC")
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

	s.logger.Info().
		Int32("messageID", request.GetMessageId()).
		Int("count", len(entries)).
		Msg("Retrieved reminder history via gRPC")
	return &database.GetReminderHistoryResponse{Entries: entries}, nil
}

// runExpiredMessageCleanup runs a background loop that periodically deletes expired messages.
// It exits when ctx is cancelled (i.e., when the server shuts down).
func (s *GRPCServer) runExpiredMessageCleanup(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.storageService.CleanupExpiredMessages(ctx); err != nil {
				s.logger.Error().Err(err).Msg("Failed to cleanup expired messages")
			}
		}
	}
}

// registerHealthServer wires the standard gRPC health service into the given
// server with the overall service ("") marked SERVING. k8s grpc: probes need a
// SERVING response on this contract to mark the pod Ready; deeper per-component
// health (e.g. DB connectivity) is intentionally out of scope here.
func (s *GRPCServer) registerHealthServer(grpcServer *grpc.Server) {
	healthSrv := health.NewServer()
	healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthSrv)
}

// Start starts the gRPC server. Fatal lifecycle decisions (process exit) are
// left to the caller so the adapter remains a pure transport layer.
func (s *GRPCServer) Start() error {
	lis, err := net.Listen("tcp", s.address)
	if err != nil {
		s.logger.Error().Err(err).Str("address", s.address).Msg("Failed to listen on gRPC address")
		return err
	}

	grpcServer := grpc.NewServer()
	database.RegisterDbServiceServer(grpcServer, s)
	s.registerHealthServer(grpcServer)
	reflection.Register(grpcServer)

	// Run expired message cleanup in the background; cancel it when Start returns.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go s.runExpiredMessageCleanup(ctx)

	s.logger.Info().Str("address", s.address).Msg("Starting gRPC storage server")

	if err := grpcServer.Serve(lis); err != nil {
		s.logger.Error().Err(err).Msg("Failed to serve gRPC storage server")
		return err
	}

	return nil
}
