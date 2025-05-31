package rabbitmq

import (
	"context"
	"fmt"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/domain"
	messagepb "github.com/Anthony-Bible/password-exchange/app/pkg/pb/message"
	"github.com/Anthony-Bible/password-exchange/app/pkg/validation"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

// NotificationPublisher implements the NotificationPublisherPort using RabbitMQ
// This is used by the reminder service to publish notification messages to the queue
// instead of sending emails directly
type NotificationPublisher struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	queueName  string
}

// NotificationConfig holds RabbitMQ connection configuration
type NotificationConfig struct {
	Host      string
	Port      int
	User      string
	Password  string
	QueueName string
}

// NewNotificationPublisher creates a new RabbitMQ notification publisher
func NewNotificationPublisher(config NotificationConfig) (*NotificationPublisher, error) {
	rabbitURL := fmt.Sprintf("amqp://%s:%s@%s:%d/", config.User, config.Password, config.Host, config.Port)

	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Error().Err(err).Str("url", rabbitURL).Msg("Failed to connect to RabbitMQ")
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Error().Err(err).Msg("Failed to open RabbitMQ channel")
		conn.Close()
		return nil, fmt.Errorf("failed to open RabbitMQ channel: %w", err)
	}

	return &NotificationPublisher{
		connection: conn,
		channel:    ch,
		queueName:  config.QueueName,
	}, nil
}

// PublishNotification publishes a notification message to RabbitMQ for processing
// Used by the reminder service to queue reminder emails instead of sending them directly
func (p *NotificationPublisher) PublishNotification(ctx context.Context, req domain.NotificationRequest) error {
	log.Debug().Str("recipientEmail", validation.SanitizeEmailForLogging(req.To)).Msg("Publishing notification message")

	// Declare the queue
	q, err := p.channel.QueueDeclare(
		p.queueName, // name
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		log.Error().Err(err).Str("queue", p.queueName).Msg("Failed to declare queue")
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Create protobuf message
	pbMsg := &messagepb.Message{
		Email:          req.From,
		FirstName:      req.FromName,
		OtherFirstName: req.RecipientName,
		OtherEmail:     req.To,
		Content:        req.MessageContent,
		Url:            req.MessageURL,
		Hidden:         req.Hidden,
	}

	// Marshal the message
	data, err := proto.Marshal(pbMsg)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal notification message")
		return fmt.Errorf("failed to marshal notification message: %w", err)
	}

	// Create context with timeout
	publishCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Publish the message
	err = p.channel.PublishWithContext(publishCtx,
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/protobuf",
			Body:         data,
		})

	if err != nil {
		log.Error().Err(err).Str("recipientEmail", validation.SanitizeEmailForLogging(req.To)).Msg("Failed to publish notification message")
		return fmt.Errorf("failed to publish notification message: %w", err)
	}

	log.Info().Str("recipientEmail", validation.SanitizeEmailForLogging(req.To)).Str("queue", p.queueName).Msg("Notification message published successfully")
	return nil
}

// Close closes the RabbitMQ connection
func (p *NotificationPublisher) Close() error {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.connection != nil {
		p.connection.Close()
	}
	log.Info().Msg("RabbitMQ notification publisher connection closed")
	return nil
}