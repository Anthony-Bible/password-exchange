package web

import (
	"context"
	"fmt"
	"time"

	apiAdapter "github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/primary/api"
	webAdapter "github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/primary/web"
	bcryptAdapter "github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/secondary/bcrypt"
	messageConfigAdapter "github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/secondary/config"
	cryptoAdapter "github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/secondary/crypto"
	grpcClients "github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/secondary/grpc_clients"
	httpAdapter "github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/secondary/http"
	memoryAdapter "github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/secondary/memory"
	rabbitMQAdapter "github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/secondary/rabbitmq"
	s3Adapter "github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/secondary/s3"
	urlAdapter "github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/secondary/url"
	messageValidationAdapter "github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/secondary/validation"
	messageDomain "github.com/Anthony-Bible/password-exchange/app/internal/domains/message/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/secondary"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/config"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/logging"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/validation"
)

// defaultFileUploadMaxSize is the default maximum permitted file upload size (100 MiB).
const defaultFileUploadMaxSize int64 = 100 << 20

type Config struct {
	config.PassConfig `mapstructure:",squash"`
}

func (conf Config) StartServer() {
	// Use hexagonal architecture
	conf.startHexagonalServer()
}

func (conf Config) startHexagonalServer() {
	// Get service endpoints
	encryptionServiceName, dbServiceName := conf.getServiceNames()

	// Create secondary adapters (clients to other services)
	encryptionClient, err := grpcClients.NewEncryptionClient(encryptionServiceName)
	if err != nil {
		logging.Fatal().Err(err).Msg("Failed to create encryption client")
	}
	defer encryptionClient.Close()

	storageClient, err := grpcClients.NewStorageClient(dbServiceName)
	if err != nil {
		logging.Fatal().Err(err).Msg("Failed to create storage client")
	}
	defer storageClient.Close()

	// Create notification publisher
	notificationConfig := rabbitMQAdapter.NotificationConfig{
		Host:      conf.RabHost,
		Port:      conf.RabPort,
		User:      conf.RabUser,
		Password:  conf.RabPass,
		QueueName: conf.RabQName,
	}

	notificationPublisher, err := rabbitMQAdapter.NewNotificationPublisher(notificationConfig)
	if err != nil {
		logging.Fatal().Err(err).Msg("Failed to create notification publisher")
	}
	defer notificationPublisher.Close()

	// Create other secondary adapters
	passwordHasher := bcryptAdapter.NewPasswordHasher(11)

	environment := conf.RunningEnvironment
	siteHost, err := validation.GetViperVariable(environment + "Host")
	if err != nil {
		logging.Fatal().Err(err).Msg("Failed to get site host")
	}
	urlBuilder := urlAdapter.NewURLBuilder(siteHost)

	// Create Turnstile validator
	turnstileValidator := httpAdapter.NewTurnstileValidator(conf.TurnstileSecret)

	// Create cross-cutting adapters
	messageLogger := logging.NewLogger()
	messageConfig := messageConfigAdapter.NewAdapter()
	messageValidation := messageValidationAdapter.NewAdapter()

	// Create message service (domain)
	messageService := messageDomain.NewMessageService(
		encryptionClient,
		storageClient,
		notificationPublisher,
		passwordHasher,
		urlBuilder,
		turnstileValidator,
		messageLogger,
		messageConfig,
		messageValidation,
	)

	var fileHandler *apiAdapter.FileAPIHandler
	if conf.ObjectStorageEndpoint != "" && conf.ObjectStorageAccessKey != "" && conf.ObjectStorageSecretKey != "" && conf.ObjectStorageBucket != "" {
		objectStorage, storageErr := s3Adapter.NewS3Adapter(
			conf.ObjectStorageEndpoint,
			conf.ObjectStorageAccessKey,
			conf.ObjectStorageSecretKey,
			conf.ObjectStorageBucket,
			conf.ObjectStorageUseSSL,
		)
		if storageErr != nil {
			logging.Fatal().Err(storageErr).Msg("Failed to create object storage adapter")
		}

		maxFileSize := conf.FileUploadMaxSize
		if maxFileSize <= 0 {
			maxFileSize = defaultFileUploadMaxSize
		}

		// Use the database-service gRPC client for durable upload-session state.
		// The in-memory adapter is the fallback only when the storage client is nil
		// (which cannot happen here since we Fatal above, but this keeps the
		// buildUploadStateAdapter contract generic).
		uploadState, uploadStateCloser := buildUploadStateAdapter(storageClient)
		if uploadStateCloser != nil {
			defer uploadStateCloser()
		}

		fileService := messageDomain.NewFileService(
			cryptoAdapter.NewFileEncryptionAdapter(encryptionClient),
			objectStorage,
			uploadState,
			logging.NewLogger(),
			maxFileSize,
		)

		// Sweep expired incomplete upload sessions hourly so abandoned sessions
		// don't retain their AES key in memory or leave dangling multipart uploads.
		go fileService.RunSessionCleanup(context.Background(), time.Hour)

		fileHandler = apiAdapter.NewFileAPIHandler(fileService)
	} else {
		logging.Warn().Msg("File upload routes disabled: object storage configuration is incomplete")
	}

	// Create web server (primary adapter). The encryption + storage clients
	// are also handed in directly so the /readyz probe can call HealthCheck
	// on them without re-routing through the message service.
	webServer := webAdapter.NewWebServer(messageService, encryptionClient, storageClient).WithFileHandler(fileHandler)

	// Start the server
	logging.Info().Msg("Starting message service with hexagonal architecture")
	if err := webServer.Start(); err != nil {
		logging.Fatal().Err(err).Msg("Failed to start hexagonal web server")
	}
}

// buildUploadStateAdapter returns a gRPC-backed UploadStatePort when a
// StorageClient is provided, or falls back to the in-memory adapter. The
// second return value is always nil (no additional resource to close —
// the caller already defers storageClient.Close()).
func buildUploadStateAdapter(storageClient *grpcClients.StorageClient) (secondary.UploadStatePort, func()) {
	if storageClient == nil {
		logging.Warn().Msg("File upload sessions will use in-memory state: no storage client provided")
		return memoryAdapter.NewMemoryUploadStateAdapter(), nil
	}
	logging.Info().Msg("Using database-service gRPC-backed durable upload session state")
	return storageClient.NewUploadStateAdapter(), nil
}

func (conf Config) getServiceNames() (string, string) {
	encryptionServiceName, err := validation.GetViperVariable(fmt.Sprintf("Encryption%sService", conf.RunningEnvironment))
	dbServiceName, err := validation.GetViperVariable(fmt.Sprintf("Database%sService", conf.RunningEnvironment))
	logging.Debug().Msg(dbServiceName)

	encryptionServiceName += ":50051"
	logging.Debug().Msg(encryptionServiceName)

	if err != nil {
		logging.Fatal().Err(err).Msg("something went wrong with getting the encryption-service address")
	}
	return encryptionServiceName, dbServiceName
}
