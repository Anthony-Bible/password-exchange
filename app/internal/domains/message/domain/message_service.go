package domain

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
)

// MessageService provides message sharing operations
type MessageService struct {
	encryptionService   EncryptionService
	storageService      StorageService
	notificationService NotificationService
	passwordHasher      PasswordHasher
	urlBuilder          URLBuilder
}

// NewMessageService creates a new message service
func NewMessageService(
	encryptionService EncryptionService,
	storageService StorageService,
	notificationService NotificationService,
	passwordHasher PasswordHasher,
	urlBuilder URLBuilder,
) *MessageService {
	return &MessageService{
		encryptionService:   encryptionService,
		storageService:      storageService,
		notificationService: notificationService,
		passwordHasher:      passwordHasher,
		urlBuilder:          urlBuilder,
	}
}

// SubmitMessage handles the submission of a new encrypted message
func (s *MessageService) SubmitMessage(ctx context.Context, req MessageSubmissionRequest) (*MessageSubmissionResponse, error) {
	log.Info().Str("senderEmail", req.SenderEmail).Msg("Processing message submission")

	// Validate the request
	if err := s.validateSubmissionRequest(req); err != nil {
		log.Error().Err(err).Msg("Invalid message submission request")
		return nil, fmt.Errorf("%w: %v", ErrInvalidMessageRequest, err)
	}

	// Generate encryption key
	encryptionKey, err := s.encryptionService.GenerateKey(ctx, 32)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate encryption key")
		return nil, fmt.Errorf("%w: %v", ErrEncryptionFailed, err)
	}

	// Encrypt the message content
	encryptedContent, err := s.encryptionService.Encrypt(ctx, []string{req.Content}, encryptionKey)
	if err != nil {
		log.Error().Err(err).Msg("Failed to encrypt message content")
		return nil, fmt.Errorf("%w: %v", ErrEncryptionFailed, err)
	}

	// Generate unique ID
	messageID, err := s.encryptionService.GenerateID(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate message ID")
		return nil, fmt.Errorf("%w: %v", ErrGenerateIDFailed, err)
	}

	// Hash passphrase if provided
	hashedPassphrase := ""
	if strings.TrimSpace(req.Passphrase) != "" {
		hashedPassphrase, err = s.passwordHasher.Hash(ctx, req.Passphrase)
		if err != nil {
			log.Error().Err(err).Msg("Failed to hash passphrase")
			return nil, fmt.Errorf("%w: %v", ErrPasswordHashFailed, err)
		}
	}

	// Build the decryption URL
	decryptURL := s.urlBuilder.BuildDecryptURL(messageID, encryptionKey)

	// Store the encrypted message
	storeReq := MessageStorageRequest{
		MessageID:  messageID,
		Content:    strings.Join(encryptedContent, ""),
		Passphrase: hashedPassphrase,
	}

	err = s.storageService.StoreMessage(ctx, storeReq)
	if err != nil {
		log.Error().Err(err).Str("messageId", messageID).Msg("Failed to store message")
		return nil, fmt.Errorf("%w: %v", ErrStorageFailed, err)
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

		err = s.notificationService.SendMessageNotification(ctx, notificationReq)
		if err != nil {
			log.Error().Err(err).Str("messageId", messageID).Msg("Failed to send notification")
			// Don't fail the entire operation for notification errors
		}
	}

	response := &MessageSubmissionResponse{
		MessageID:  messageID,
		DecryptURL: decryptURL,
		Key:        base64.URLEncoding.EncodeToString(encryptionKey),
		Success:    true,
	}

	log.Info().Str("messageId", messageID).Str("url", decryptURL).Msg("Message submitted successfully")
	return response, nil
}

// RetrieveMessage handles the retrieval and decryption of a stored message
func (s *MessageService) RetrieveMessage(ctx context.Context, req MessageRetrievalRequest) (*MessageRetrievalResponse, error) {
	log.Debug().Str("messageId", req.MessageID).Msg("Processing message retrieval")

	// First, get message metadata without incrementing view count to check passphrase
	storageReq := MessageRetrievalStorageRequest{
		MessageID: req.MessageID,
	}

	storedMessageMeta, err := s.storageService.GetMessage(ctx, storageReq)
	if err != nil {
		log.Error().Err(err).Str("messageId", req.MessageID).Msg("Failed to get stored message metadata")
		return nil, fmt.Errorf("%w: %v", ErrMessageNotFound, err)
	}

	// Verify passphrase if required BEFORE retrieving full message and incrementing view count
	if storedMessageMeta.HasPassphrase {
		valid, err := s.passwordHasher.Verify(ctx, req.Passphrase, storedMessageMeta.HashedPassphrase)
		if err != nil {
			log.Error().Err(err).Str("messageId", req.MessageID).Msg("Failed to verify passphrase")
			return nil, fmt.Errorf("%w: %v", ErrPasswordVerificationFailed, err)
		}
		if !valid {
			log.Warn().Str("messageId", req.MessageID).Msg("Invalid passphrase provided")
			return nil, ErrInvalidPassphrase
		}
	}

	// Only after successful passphrase validation, retrieve full message and increment view count
	storedMessage, err := s.storageService.RetrieveMessage(ctx, storageReq)
	if err != nil {
		log.Error().Err(err).Str("messageId", req.MessageID).Msg("Failed to retrieve stored message")
		return nil, fmt.Errorf("%w: %v", ErrMessageNotFound, err)
	}

	// Decrypt the message content
	decryptedContent, err := s.encryptionService.Decrypt(ctx, []string{storedMessage.EncryptedContent}, req.DecryptionKey)
	if err != nil {
		log.Error().Err(err).Str("messageId", req.MessageID).Msg("Failed to decrypt message content")
		return nil, fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	// Decode the final content
	finalContent := ""
	if len(decryptedContent) > 0 {
		decodedBytes, err := base64.URLEncoding.DecodeString(decryptedContent[0])
		if err != nil {
			log.Error().Err(err).Str("messageId", req.MessageID).Msg("Failed to decode message content")
			return nil, fmt.Errorf("%w: %v", ErrDecodingFailed, err)
		}
		finalContent = string(decodedBytes)
	}

	response := &MessageRetrievalResponse{
		MessageID: req.MessageID,
		Content:   finalContent,
		ViewCount: storedMessage.ViewCount,
		Success:   true,
	}

	log.Debug().Str("messageId", req.MessageID).Int("viewCount", storedMessage.ViewCount).Msg("Message retrieved successfully")
	return response, nil
}

// CheckMessageAccess checks if a message exists and whether it requires a passphrase
func (s *MessageService) CheckMessageAccess(ctx context.Context, messageID string) (*MessageAccessInfo, error) {
	log.Debug().Str("messageId", messageID).Msg("Checking message access")

	storageReq := MessageRetrievalStorageRequest{
		MessageID: messageID,
	}

	storedMessage, err := s.storageService.GetMessage(ctx, storageReq)
	if err != nil {
		log.Error().Err(err).Str("messageId", messageID).Msg("Failed to check message access")
		return nil, fmt.Errorf("%w: %v", ErrMessageNotFound, err)
	}

	accessInfo := &MessageAccessInfo{
		MessageID:          messageID,
		RequiresPassphrase: storedMessage.HasPassphrase,
		Exists:             true,
	}

	log.Debug().Str("messageId", messageID).Bool("requiresPassphrase", accessInfo.RequiresPassphrase).Msg("Message access checked")
	return accessInfo, nil
}

// validateSubmissionRequest validates the message submission request
func (s *MessageService) validateSubmissionRequest(req MessageSubmissionRequest) error {
	if strings.TrimSpace(req.Content) == "" {
		return fmt.Errorf("message content is required")
	}

	// Only validate sender information if email notifications are enabled
	if !req.SkipEmail {
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
	}

	if req.SendNotification {
		if strings.TrimSpace(req.RecipientEmail) == "" {
			return fmt.Errorf("recipient email is required when sending notifications")
		}

		if strings.TrimSpace(req.RecipientName) == "" {
			return fmt.Errorf("recipient name is required when sending notifications")
		}

		// Basic email validation for recipient
		if !strings.Contains(req.RecipientEmail, "@") {
			return ErrInvalidEmailAddress
		}
	}

	return nil
}
