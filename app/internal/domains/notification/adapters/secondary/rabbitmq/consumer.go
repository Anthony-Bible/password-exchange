package rabbitmq

import (
	"context"
	"fmt"
	"os"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/domain"
	pb "github.com/Anthony-Bible/password-exchange/app/pkg/pb/message"
	"github.com/Anthony-Bible/password-exchange/app/pkg/validation"
	"github.com/golang/protobuf/proto"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

// RabbitMQConsumer implements the QueueConsumerPort using RabbitMQ
type RabbitMQConsumer struct {
	connection *amqp.Connection
	channel    *amqp.Channel
}

// NewRabbitMQConsumer creates a new RabbitMQ consumer
func NewRabbitMQConsumer() *RabbitMQConsumer {
	return &RabbitMQConsumer{}
}

// Connect establishes a connection to RabbitMQ
func (r *RabbitMQConsumer) Connect(ctx context.Context, queueConn domain.QueueConnection) error {
	rabbitURL := fmt.Sprintf("amqp://%s:%s@%s:%d/", queueConn.User, queueConn.Password, queueConn.Host, queueConn.Port)
	
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Error().Err(err).Str("url", rabbitURL).Msg("Failed to connect to RabbitMQ")
		return fmt.Errorf("%w: %v", domain.ErrQueueConnectionFailed, err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Error().Err(err).Msg("Failed to open RabbitMQ channel")
		conn.Close()
		return fmt.Errorf("%w: %v", domain.ErrQueueConnectionFailed, err)
	}

	r.connection = conn
	r.channel = ch

	log.Info().Str("host", queueConn.Host).Int("port", queueConn.Port).Msg("Connected to RabbitMQ")
	return nil
}

// StartConsuming starts consuming messages from the queue
func (r *RabbitMQConsumer) StartConsuming(ctx context.Context, queueConn domain.QueueConnection, handler domain.MessageHandler, concurrency int) error {
	if r.channel == nil {
		if err := r.Connect(ctx, queueConn); err != nil {
			return err
		}
	}

	// Declare the queue
	q, err := r.channel.QueueDeclare(
		queueConn.QueueName, // name
		true,                // durable
		false,               // delete when unused
		false,               // exclusive
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		log.Error().Err(err).Str("queue", queueConn.QueueName).Msg("Failed to declare queue")
		return fmt.Errorf("%w: %v", domain.ErrQueueConsumeFailed, err)
	}

	// Set QoS
	err = r.channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to set QoS")
		return fmt.Errorf("%w: %v", domain.ErrQueueConsumeFailed, err)
	}

	// Start consuming
	msgs, err := r.channel.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to register consumer")
		return fmt.Errorf("%w: %v", domain.ErrQueueConsumeFailed, err)
	}

	log.Info().Str("queue", queueConn.QueueName).Int("concurrency", concurrency).Msg("Starting RabbitMQ consumer")

	// Start worker goroutines
	for i := 0; i < concurrency; i++ {
		go r.messageWorker(ctx, msgs, handler, i)
	}

	// Wait for context cancellation
	<-ctx.Done()
	log.Info().Msg("RabbitMQ consumer shutting down")
	return nil
}

// messageWorker processes messages from the queue
func (r *RabbitMQConsumer) messageWorker(ctx context.Context, msgs <-chan amqp.Delivery, handler domain.MessageHandler, workerID int) {
	log.Debug().Int("workerId", workerID).Msg("Starting RabbitMQ message worker")

	for {
		select {
		case <-ctx.Done():
			log.Debug().Int("workerId", workerID).Msg("Message worker shutting down")
			return
		case msg, ok := <-msgs:
			if !ok {
				log.Error().Int("workerId", workerID).Msg("RabbitMQ message channel closed")
				os.Exit(1) // Critical error - exit to restart
				return
			}

			success := r.handleMessage(ctx, msg, handler, workerID)
			if success {
				msg.Ack(false)
			} else {
				msg.Nack(false, true) // Requeue the message
			}
		}
	}
}

// handleMessage processes a single message
func (r *RabbitMQConsumer) handleMessage(ctx context.Context, delivery amqp.Delivery, handler domain.MessageHandler, workerID int) bool {
	if delivery.Body == nil {
		log.Error().Int("workerId", workerID).Msg("Received message with empty body")
		return false
	}

	// Unmarshal protobuf message
	var pbMsg pb.Message
	err := proto.Unmarshal(delivery.Body, &pbMsg)
	if err != nil {
		log.Error().Err(err).Int("workerId", workerID).Msg("Failed to unmarshal message")
		return false
	}

	// Convert to domain message
	queueMsg := domain.QueueMessage{
		Email:          pbMsg.Email,
		FirstName:      pbMsg.Firstname,
		OtherFirstName: pbMsg.Otherfirstname,
		OtherLastName:  pbMsg.OtherLastName,
		OtherEmail:     pbMsg.OtherEmail,
		UniqueID:       pbMsg.Uniqueid,
		Content:        pbMsg.Content,
		URL:            pbMsg.Url,
		Hidden:         pbMsg.Hidden,
		Captcha:        pbMsg.Captcha,
	}

	// Handle the message
	err = handler.HandleMessage(ctx, queueMsg)
	if err != nil {
		log.Error().Err(err).Int("workerId", workerID).Str("to", validation.SanitizeEmailForLogging(queueMsg.OtherEmail)).Msg("Failed to handle message")
		return false
	}

	log.Debug().Int("workerId", workerID).Str("to", validation.SanitizeEmailForLogging(queueMsg.OtherEmail)).Msg("Successfully handled message")
	return true
}

// Close closes the RabbitMQ connection
func (r *RabbitMQConsumer) Close() error {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.connection != nil {
		r.connection.Close()
	}
	log.Info().Msg("RabbitMQ connection closed")
	return nil
}