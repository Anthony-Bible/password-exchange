package domain

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/secondary"
)

const notificationSendTimeout = 2 * time.Second

// MessageService provides message sharing operations.
type MessageService struct {
	encryptionService   secondary.EncryptionServicePort
	storageService      secondary.StorageServicePort
	notificationService secondary.NotificationServicePort
	passwordHasher      secondary.PasswordHasherPort
	urlBuilder          secondary.URLBuilderPort
	turnstileValidator  secondary.TurnstileValidatorPort
	logger              secondary.LoggerPort
	config              secondary.ConfigPort
	validation          secondary.ValidationPort
}

// NewMessageService creates a new message service.
func NewMessageService(
	encryptionService secondary.EncryptionServicePort,
	storageService secondary.StorageServicePort,
	notificationService secondary.NotificationServicePort,
	passwordHasher secondary.PasswordHasherPort,
	urlBuilder secondary.URLBuilderPort,
	turnstileValidator secondary.TurnstileValidatorPort,
	logger secondary.LoggerPort,
	config secondary.ConfigPort,
	validation secondary.ValidationPort,
) *MessageService {
	return &MessageService{
		encryptionService:   encryptionService,
		storageService:      storageService,
		notificationService: notificationService,
		passwordHasher:      passwordHasher,
		urlBuilder:          urlBuilder,
		turnstileValidator:  turnstileValidator,
		logger:              logger,
		config:              config,
		validation:          validation,
	}
}

// SubmitMessage handles the submission of a new encrypted message.
func (s *MessageService) SubmitMessage(
	ctx context.Context,
	req MessageSubmissionRequest,
) (*MessageSubmissionResponse, error) {
	s.logger.Info().
		Str("senderEmail", s.validation.SanitizeEmailForLogging(req.SenderEmail)).
		Msg("Processing message submission")

	// Validate the request
	if err := s.validateSubmissionRequest(req); err != nil {
		s.logger.Error().Err(err).Msg("Invalid message submission request")
		return nil, fmt.Errorf("%w: %w", ErrInvalidMessageRequest, err)
	}

	// Validate Turnstile token only if sending email notifications
	if req.SendNotification {
		if strings.TrimSpace(req.TurnstileToken) == "" {
			s.logger.Error().Msg("Missing Turnstile token for email notification")
			return nil, fmt.Errorf("%w: missing Turnstile token", ErrInvalidMessageRequest)
		}

		// Extract remote IP from context if available
		remoteIP := ""
		if ip := ctx.Value("RemoteIP"); ip != nil {
			if ipStr, ok := ip.(string); ok {
				remoteIP = ipStr
			}
		}

		valid, err := s.turnstileValidator.ValidateToken(ctx, req.TurnstileToken, remoteIP)
		if err != nil {
			s.logger.Error().Err(err).Msg("Failed to validate Turnstile token")
			return nil, fmt.Errorf("%w: turnstile validation error: %w", ErrInvalidMessageRequest, err)
		}
		if !valid {
			s.logger.Warn().Msg("Turnstile token validation failed")
			return nil, fmt.Errorf("%w: turnstile validation failed", ErrInvalidMessageRequest)
		}
		s.logger.Debug().Msg("Turnstile token validated successfully")
	} else {
		s.logger.Debug().Msg("Skipping Turnstile validation - email notifications disabled")
	}
	encryptedString, encryptionKey, err := s.prepareStoredContent(ctx, req)
	if err != nil {
		return nil, err
	}

	// Generate unique ID
	messageID, err := s.encryptionService.GenerateID(ctx)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to generate message ID")
		return nil, fmt.Errorf("%w: %w", ErrGenerateIDFailed, err)
	}

	// Hash passphrase if provided
	hashedPassphrase := ""
	if strings.TrimSpace(req.Passphrase) != "" {
		hashedPassphrase, err = s.passwordHasher.Hash(ctx, req.Passphrase)
		if err != nil {
			s.logger.Error().Err(err).Msg("Failed to hash passphrase")
			return nil, fmt.Errorf("%w: %w", ErrPasswordHashFailed, err)
		}
	}

	// Build the decryption URL
	decryptURL := s.buildDecryptURL(messageID, req.IsClientEncrypted, encryptionKey, req.E2EKey)

	// Determine max view count (use request value or default from config)
	maxViewCount := req.MaxViewCount
	if maxViewCount <= 0 {
		maxViewCount = s.config.GetDefaultMaxViewCount()
		if maxViewCount <= 0 {
			maxViewCount = DefaultMaxViewCount
		}
	}

	// Determine TTL: use provided ExpirationHours or fall back to default
	ttl := DefaultMessageTTL
	if req.ExpirationHours > 0 {
		ttl = time.Duration(req.ExpirationHours) * time.Hour
	}
	expiresAt := time.Now().UTC().Add(ttl)

	// Store the encrypted message
	storeReq := MessageStorageRequest{
		MessageID:         messageID,
		Content:           encryptedString,
		IsClientEncrypted: req.IsClientEncrypted,
		Passphrase:        hashedPassphrase,
		MaxViewCount:      maxViewCount,
		ExpiresAt:         &expiresAt,
	}

	// Only store recipient email if email notifications are enabled
	if req.SendNotification {
		storeReq.RecipientEmail = req.RecipientEmail
	}

	err = s.storageService.StoreMessage(ctx, storeReq)
	if err != nil {
		s.logger.Error().Err(err).Str("messageId", messageID).Msg("Failed to store message")
		return nil, fmt.Errorf("%w: %w", ErrStorageFailed, err)
	}

	// Send notification if requested
	if req.SendNotification && strings.TrimSpace(req.RecipientEmail) != "" {
		notificationReq := MessageNotificationRequest{
			SenderName:     req.SenderName,
			SenderEmail:    req.SenderEmail,
			RecipientName:  req.RecipientName,
			RecipientEmail: req.RecipientEmail,
			MessageURL:     decryptURL,
			AdditionalInfo: req.AdditionalInfo,
		}

		go func(messageID string, notificationReq MessageNotificationRequest) {
			defer func() {
				if recovered := recover(); recovered != nil {
					s.logger.Error().
						Interface("panic", recovered).
						Str("messageId", messageID).
						Msg("Notification send panicked")
				}
			}()

			notificationCtx, cancel := context.WithTimeout(context.Background(), notificationSendTimeout)
			defer cancel()

			err := s.notificationService.SendMessageNotification(notificationCtx, notificationReq)
			if err != nil {
				s.logger.Error().Err(err).Str("messageId", messageID).Msg("Failed to send notification")
				// Don't fail the entire operation for notification errors
			}
		}(messageID, notificationReq)
	}

	response := &MessageSubmissionResponse{
		MessageID:         messageID,
		DecryptURL:        decryptURL,
		Key:               base64.URLEncoding.EncodeToString(encryptionKey),
		IsClientEncrypted: req.IsClientEncrypted,
		ExpiresAt:         &expiresAt,
		Success:           true,
	}
	if req.IsClientEncrypted {
		response.Key = ""
	}

	s.logger.Info().Str("messageId", messageID).Str("url", decryptURL).Msg("Message submitted successfully")
	return response, nil
}

// RetrieveMessage handles the retrieval and decryption of a stored message.
func (s *MessageService) RetrieveMessage(
	ctx context.Context,
	req MessageRetrievalRequest,
) (*MessageRetrievalResponse, error) {
	s.logger.Debug().Str("messageId", req.MessageID).Msg("Processing message retrieval")

	// First, get message metadata without incrementing view count to check passphrase
	storageReq := MessageRetrievalStorageRequest{
		MessageID: req.MessageID,
	}

	storedMessageMeta, err := s.storageService.GetMessage(ctx, storageReq)
	if err != nil {
		s.logger.Error().Err(err).Str("messageId", req.MessageID).Msg("Failed to get stored message metadata")
		return nil, fmt.Errorf("%w: %w", ErrMessageNotFound, err)
	}

	// Verify passphrase if required BEFORE retrieving full message and incrementing view count
	if storedMessageMeta.HasPassphrase {
		valid, err := s.passwordHasher.Verify(ctx, req.Passphrase, storedMessageMeta.HashedPassphrase)
		if err != nil {
			s.logger.Error().Err(err).Str("messageId", req.MessageID).Msg("Failed to verify passphrase")
			return nil, fmt.Errorf("%w: %w", ErrPasswordVerificationFailed, err)
		}
		if !valid {
			s.logger.Warn().Str("messageId", req.MessageID).Msg("Invalid passphrase provided")
			return nil, ErrInvalidPassphrase
		}
	}

	// Only after successful passphrase validation, retrieve full message and increment view count
	storedMessage, err := s.storageService.RetrieveMessage(ctx, storageReq)
	if err != nil {
		s.logger.Error().Err(err).Str("messageId", req.MessageID).Msg("Failed to retrieve stored message")
		return nil, fmt.Errorf("%w: %w", ErrMessageNotFound, err)
	}

	finalContent, err := s.resolveRetrievedContent(ctx, req.MessageID, storedMessage, req.DecryptionKey)
	if err != nil {
		return nil, err
	}

	response := &MessageRetrievalResponse{
		MessageID:         req.MessageID,
		Content:           finalContent,
		IsClientEncrypted: storedMessage.IsClientEncrypted,
		ViewCount:         storedMessage.ViewCount,
		MaxViewCount:      storedMessage.MaxViewCount,
		ExpiresAt:         storedMessage.ExpiresAt,
		Success:           true,
	}

	s.logger.Debug().
		Str("messageId", req.MessageID).
		Int("viewCount", storedMessage.ViewCount).
		Msg("Message retrieved successfully")
	return response, nil
}

// CheckMessageAccess checks if a message exists and whether it requires a passphrase.
func (s *MessageService) CheckMessageAccess(ctx context.Context, messageID string) (*MessageAccessInfo, error) {
	s.logger.Debug().Str("messageId", messageID).Msg("Checking message access")

	storageReq := MessageRetrievalStorageRequest{
		MessageID: messageID,
	}

	storedMessage, err := s.storageService.GetMessage(ctx, storageReq)
	if err != nil {
		s.logger.Error().Err(err).Str("messageId", messageID).Msg("Failed to check message access")
		return nil, fmt.Errorf("%w: %w", ErrMessageNotFound, err)
	}

	accessInfo := &MessageAccessInfo{
		MessageID:          messageID,
		RequiresPassphrase: storedMessage.HasPassphrase,
		IsClientEncrypted:  storedMessage.IsClientEncrypted,
		Exists:             true,
		ExpiresAt:          storedMessage.ExpiresAt,
	}

	s.logger.Debug().
		Str("messageId", messageID).
		Bool("requiresPassphrase", accessInfo.RequiresPassphrase).
		Msg("Message access checked")
	return accessInfo, nil
}

func (s *MessageService) prepareStoredContent(
	ctx context.Context,
	req MessageSubmissionRequest,
) (string, []byte, error) {
	if req.IsClientEncrypted {
		return req.Content, nil, nil
	}

	encryptionKey, err := s.encryptionService.GenerateKey(ctx, 32)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to generate encryption key")
		return "", nil, fmt.Errorf("%w: %w", ErrEncryptionFailed, err)
	}

	encryptedContent, encryptErr := s.encryptionService.Encrypt(ctx, []string{req.Content}, encryptionKey)
	if encryptErr != nil {
		s.logger.Error().Err(encryptErr).Msg("Failed to encrypt message content")
		return "", nil, fmt.Errorf("%w: %w", ErrEncryptionFailed, encryptErr)
	}

	return strings.Join(encryptedContent, ""), encryptionKey, nil
}

func (s *MessageService) buildDecryptURL(messageID string, isClientEncrypted bool, encryptionKey []byte, e2eKey string) string {
	if isClientEncrypted {
		return s.urlBuilder.BuildE2EDecryptURL(messageID, e2eKey)
	}
	return s.urlBuilder.BuildDecryptURL(messageID, encryptionKey)
}

func (s *MessageService) resolveRetrievedContent(
	ctx context.Context,
	messageID string,
	storedMessage *MessageStorageResponse,
	decryptionKey []byte,
) (string, error) {
	if storedMessage == nil {
		return "", fmt.Errorf("%w: empty storage response", ErrMessageNotFound)
	}

	if storedMessage.IsClientEncrypted {
		return storedMessage.EncryptedContent, nil
	}

	decryptedContent, decryptErr := s.encryptionService.Decrypt(
		ctx,
		[]string{storedMessage.EncryptedContent},
		decryptionKey,
	)
	if decryptErr != nil {
		s.logger.Error().Err(decryptErr).Str("messageId", messageID).Msg("Failed to decrypt message content")
		return "", fmt.Errorf("%w: %w", ErrDecryptionFailed, decryptErr)
	}

	if len(decryptedContent) == 0 {
		return "", nil
	}

	decodedBytes, decodeErr := base64.URLEncoding.DecodeString(decryptedContent[0])
	if decodeErr != nil {
		s.logger.Error().Err(decodeErr).Str("messageId", messageID).Msg("Failed to decode message content")
		return "", fmt.Errorf("%w: %w", ErrDecodingFailed, decodeErr)
	}

	return string(decodedBytes), nil
}

// NotifyMessage sends the email notification for an existing message using the
// caller-supplied ShareURL. This is called by the frontend after the optional
// file upload completes so the URL already contains all fragment params.
func (s *MessageService) NotifyMessage(ctx context.Context, req MessageNotifyRequest) error {
	s.logger.Info().
		Str("messageId", req.MessageID).
		Msg("Processing deferred message notification")

	if strings.TrimSpace(req.ShareURL) == "" {
		return fmt.Errorf("%w: shareURL is required", ErrInvalidMessageRequest)
	}
	if strings.TrimSpace(req.RecipientEmail) == "" {
		return fmt.Errorf("%w: recipientEmail is required", ErrInvalidMessageRequest)
	}
	if !strings.Contains(req.RecipientEmail, "@") {
		return fmt.Errorf("%w: %w", ErrInvalidMessageRequest, ErrInvalidEmailAddress)
	}

	// Validate Turnstile token
	if strings.TrimSpace(req.TurnstileToken) == "" {
		return fmt.Errorf("%w: missing Turnstile token", ErrInvalidMessageRequest)
	}
	remoteIP := ""
	if ip := ctx.Value("RemoteIP"); ip != nil {
		if ipStr, ok := ip.(string); ok {
			remoteIP = ipStr
		}
	}
	valid, err := s.turnstileValidator.ValidateToken(ctx, req.TurnstileToken, remoteIP)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to validate Turnstile token for notify")
		return fmt.Errorf("%w: turnstile validation error: %w", ErrInvalidMessageRequest, err)
	}
	if !valid {
		s.logger.Warn().Msg("Turnstile token validation failed for notify")
		return fmt.Errorf("%w: turnstile validation failed", ErrInvalidMessageRequest)
	}

	// Verify the message exists
	_, err = s.storageService.GetMessage(ctx, MessageRetrievalStorageRequest{MessageID: req.MessageID})
	if err != nil {
		s.logger.Error().Err(err).Str("messageId", req.MessageID).Msg("Message not found for notify")
		return fmt.Errorf("%w: %w", ErrMessageNotFound, err)
	}

	notificationReq := MessageNotificationRequest{
		SenderName:     req.SenderName,
		SenderEmail:    req.SenderEmail,
		RecipientName:  req.RecipientName,
		RecipientEmail: req.RecipientEmail,
		MessageURL:     req.ShareURL,
		AdditionalInfo: req.AdditionalInfo,
	}

	notificationCtx, cancel := context.WithTimeout(context.Background(), notificationSendTimeout)
	defer cancel()

	if err := s.notificationService.SendMessageNotification(notificationCtx, notificationReq); err != nil {
		s.logger.Error().Err(err).Str("messageId", req.MessageID).Msg("Failed to send deferred notification")
		return fmt.Errorf("notification send failed: %w", err)
	}

	s.logger.Info().Str("messageId", req.MessageID).Msg("Deferred notification sent successfully")
	return nil
}

// validateSubmissionRequest validates the message submission request.
func (s *MessageService) validateSubmissionRequest(req MessageSubmissionRequest) error {
	if strings.TrimSpace(req.Content) == "" {
		return fmt.Errorf("message content is required")
	}

	// Validate max view count if provided
	if req.MaxViewCount != 0 {
		if req.MaxViewCount < 1 || req.MaxViewCount > AbsoluteMaxViewCount {
			return fmt.Errorf("max view count must be between 1 and %d", AbsoluteMaxViewCount)
		}
	}

	// Validate expiration hours if provided (0 means use default TTL)
	if req.ExpirationHours != 0 {
		if req.ExpirationHours < MinExpirationHours || req.ExpirationHours > MaxExpirationHours {
			return fmt.Errorf(
				"expiration must be between %d and %d hours (%d days)",
				MinExpirationHours,
				MaxExpirationHours,
				MaxExpirationHours/24,
			)
		}
	}

	// Only validate sender and recipient information if email notifications are enabled
	if req.SendNotification {
		if strings.TrimSpace(req.SenderName) == "" {
			return fmt.Errorf("sender name is required")
		}

		if strings.TrimSpace(req.SenderEmail) == "" {
			return fmt.Errorf("sender email is required")
		}

		// Basic email validation for sender
		if !strings.Contains(req.SenderEmail, "@") {
			return ErrInvalidEmailAddress
		}

		if strings.TrimSpace(req.RecipientName) == "" {
			return fmt.Errorf("recipient name is required when sending notifications")
		}

		if strings.TrimSpace(req.RecipientEmail) == "" {
			return fmt.Errorf("recipient email is required when sending notifications")
		}

		// Basic email validation for recipient
		if !strings.Contains(req.RecipientEmail, "@") {
			return ErrInvalidEmailAddress
		}
	}

	return nil
}
