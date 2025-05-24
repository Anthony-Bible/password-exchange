package email

import (
	"context"

	"github.com/Anthony-Bible/password-exchange/app/internal/shared/config"
	notificationDomain "github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/domain"
	notificationConsumer "github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/adapters/primary/consumer"
	smtpSender "github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/adapters/secondary/smtp"
	rabbitMQConsumer "github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/adapters/secondary/rabbitmq"
	"github.com/rs/zerolog/log"
)

type Config struct {
	config.PassConfig `mapstructure:",squash"`
}


func (conf Config) StartProcessing() {
	// Use hexagonal architecture
	conf.startHexagonalProcessing()
}

func (conf Config) startHexagonalProcessing() {
	ctx := context.Background()

	// Create email connection configuration
	emailConn := notificationDomain.EmailConnection{
		Host:     conf.EmailHost,
		Port:     conf.EmailPort,
		User:     conf.EmailUser,
		Password: conf.EmailPass,
		From:     conf.EmailFrom,
	}

	// Create queue connection configuration
	queueConn := notificationDomain.QueueConnection{
		Host:      conf.RabHost,
		Port:      conf.RabPort,
		User:      conf.RabUser,
		Password:  conf.RabPass,
		QueueName: conf.RabQName,
	}

	// Create secondary adapters
	emailSender := smtpSender.NewSMTPSender(emailConn)
	queueConsumer := rabbitMQConsumer.NewRabbitMQConsumer()

	// Create notification service (domain)
	notificationService := notificationDomain.NewNotificationService(emailSender, queueConsumer, nil)

	// Create primary adapter (consumer)
	consumer := notificationConsumer.NewNotificationConsumer(notificationService, queueConn, 100)

	// Start processing
	log.Info().Msg("Starting notification service with hexagonal architecture")
	if err := consumer.Start(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to start hexagonal notification consumer")
	}
}

