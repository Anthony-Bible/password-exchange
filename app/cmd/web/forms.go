package web

import (
	"fmt"

	"github.com/Anthony-Bible/password-exchange/app/internal/shared/config"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/validation"
	messageDomain "github.com/Anthony-Bible/password-exchange/app/internal/domains/message/domain"
	grpcClients "github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/secondary/grpc_clients"
	rabbitMQAdapter "github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/secondary/rabbitmq"
	bcryptAdapter "github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/secondary/bcrypt"
	urlAdapter "github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/secondary/url"
	webAdapter "github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/primary/web"
	"github.com/rs/zerolog/log"
)

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
		log.Fatal().Err(err).Msg("Failed to create encryption client")
	}
	defer encryptionClient.Close()
	
	storageClient, err := grpcClients.NewStorageClient(dbServiceName)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create storage client")
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
		log.Fatal().Err(err).Msg("Failed to create notification publisher")
	}
	defer notificationPublisher.Close()
	
	// Create other secondary adapters
	passwordHasher := bcryptAdapter.NewPasswordHasher(11)
	
	environment := conf.RunningEnvironment
	siteHost, err := validation.GetViperVariable(environment + "Host")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get site host")
	}
	urlBuilder := urlAdapter.NewURLBuilder(siteHost)
	
	// Create message service (domain)
	messageService := messageDomain.NewMessageService(
		encryptionClient,
		storageClient,
		notificationPublisher,
		passwordHasher,
		urlBuilder,
	)
	
	// Create web server (primary adapter)
	webServer := webAdapter.NewWebServer(messageService)
	
	// Start the server
	log.Info().Msg("Starting message service with hexagonal architecture")
	if err := webServer.Start(); err != nil {
		log.Fatal().Err(err).Msg("Failed to start hexagonal web server")
	}
}


func (conf Config) getServiceNames() (string, string) {
	encryptionServiceName, err := validation.GetViperVariable(fmt.Sprintf("Encryption%sService", conf.RunningEnvironment))
	dbServiceName, err := validation.GetViperVariable(fmt.Sprintf("Database%sService", conf.RunningEnvironment))
	log.Debug().Msg(dbServiceName)

	encryptionServiceName += ":50051"
	log.Debug().Msg(encryptionServiceName)

	if err != nil {
		log.Fatal().Err(err).Msg("something went wrong with getting the encryption-service address")
	}
	return encryptionServiceName, dbServiceName
}
