package email

import (
	"context"

	notificationConsumer "github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/adapters/primary/consumer"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/adapters/secondary/logger"
	rabbitMQConsumer "github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/adapters/secondary/rabbitmq"
	sharedConfig "github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/adapters/secondary/shared"
	smtpSender "github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/adapters/secondary/smtp"
	notificationDomain "github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/config"
	"github.com/Anthony-Bible/password-exchange/app/pkg/validation"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/logging"
)

type Config struct {
	config.PassConfig `mapstructure:",squash"`
}

// Simple validation adapter using existing validation package
type validationAdapter struct{}

func (v *validationAdapter) ValidateEmail(email string) error {
	return validation.ValidateEmail(email)
}

func (v *validationAdapter) SanitizeEmailForLogging(email string) string {
	return validation.SanitizeEmailForLogging(email)
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

	// Create port adapters using existing functionality
	configPort := sharedConfig.NewSharedConfigAdapter(conf.PassConfig)
	loggerPort := logger.NewAdapter()
	validationPort := &validationAdapter{}

	// Create secondary adapters
	emailSender := smtpSender.NewSMTPSender(emailConn, configPort, loggerPort, validationPort)
	queueConsumer := rabbitMQConsumer.NewRabbitMQConsumer()

	// Create notification service (domain) - using WithReminder constructor with nil reminder service since email command doesn't need reminders
	notificationService := notificationDomain.NewNotificationServiceWithReminder(emailSender, queueConsumer, nil, nil, loggerPort, validationPort, configPort)

	// Create primary adapter (consumer)
	consumer := notificationConsumer.NewNotificationConsumer(notificationService, queueConn, 100)

	// Start processing
	log.Info().Msg("Starting notification service with hexagonal architecture")
	if err := consumer.Start(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to start hexagonal notification consumer")
	}
}
