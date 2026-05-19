package email

import (
	"context"

	notificationConsumer "github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/adapters/primary/consumer"
	rabbitMQConsumer "github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/adapters/secondary/rabbitmq"
	sharedConfig "github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/adapters/secondary/shared"
	smtpSender "github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/adapters/secondary/smtp"
	notificationDomain "github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/ports/contracts"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/config"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/logging"
	"github.com/Anthony-Bible/password-exchange/app/pkg/validation"
	"time"
)

type Config struct {
	config.PassConfig `mapstructure:",squash"`
}

type loggerAdapter struct{}

func (l *loggerAdapter) Debug() contracts.LogEvent { return &logEvent{log.Debug()} }
func (l *loggerAdapter) Info() contracts.LogEvent  { return &logEvent{log.Info()} }
func (l *loggerAdapter) Warn() contracts.LogEvent  { return &logEvent{log.Warn()} }
func (l *loggerAdapter) Error() contracts.LogEvent { return &logEvent{log.Error()} }

type logEvent struct {
	event *log.Event
}

func (e *logEvent) Err(err error) contracts.LogEvent { e.event = e.event.Err(err); return e }
func (e *logEvent) Str(key, value string) contracts.LogEvent {
	e.event = e.event.Str(key, value)
	return e
}
func (e *logEvent) Int(key string, value int) contracts.LogEvent {
	e.event = e.event.Int(key, value)
	return e
}
func (e *logEvent) Bool(key string, value bool) contracts.LogEvent {
	e.event = e.event.Bool(key, value)
	return e
}
func (e *logEvent) Dur(key string, value time.Duration) contracts.LogEvent {
	e.event = e.event.Dur(key, value)
	return e
}
func (e *logEvent) Float64(key string, value float64) contracts.LogEvent {
	e.event = e.event.Float64(key, value)
	return e
}
func (e *logEvent) Msg(msg string) { e.event.Msg(msg) }

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
	loggerPort := &loggerAdapter{}
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
