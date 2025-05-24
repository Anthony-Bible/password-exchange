package email

import (
	"bytes"
	"context"
	"fmt"
	"net/smtp"
	"os"
	"text/template"

	"github.com/Anthony-Bible/password-exchange/app/internal/shared/config"
	notificationDomain "github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/domain"
	notificationConsumer "github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/adapters/primary/consumer"
	smtpSender "github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/adapters/secondary/smtp"
	rabbitMQConsumer "github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/adapters/secondary/rabbitmq"
	pb "github.com/Anthony-Bible/password-exchange/app/pkg/pb/message"
	"github.com/golang/protobuf/proto"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

type Config struct {
	Other             map[string]interface{} `mapstructure:",remain"`
	Channel           *amqp.Channel          `mapstructure:",omitempty"`
	config.PassConfig `mapstructure:",squash"`
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Error().Msgf("%s: %s", msg, err)
	}
}
func (conf *Config) GetConn(rabbitUrl string) error {
	conn, err := amqp.Dial(rabbitUrl)
	if err != nil {
		log.Err(err).Msg("Problem with connecting")
	}
	ch, err := conn.Channel()
	conf.Channel = ch
	fmt.Printf("Creating connection: %+v", conf)
	return err
}
func (conn Config) startConsumer(queueName string, handler func(conf Config, d amqp.Delivery) bool, concurrency int) {
	fmt.Printf("%+v", conn)
	q, err := conn.Channel.QueueDeclare(
		conn.RabQName, // name
		true,          // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	defer conn.Channel.Close()
	failOnError(err, "Failed to declare a queue")
	err = conn.Channel.Qos(
		1,     //prefetch count
		0,     //prefetch size
		false, //global
	)
	failOnError(err, "failed to set qos")
	msgs, err := conn.Channel.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")
	var forever chan struct{}
	for i := 0; i < concurrency; i++ {
		go func() {
			for msg := range msgs {
				// if tha handler returns true then ACK, else NACK
				// the message back into the rabbit queue for
				// another round of processing
				if handler(conn, msg) {
					msg.Ack(false)
				} else {
					msg.Nack(false, true)
				}
			}
			log.Fatal().Msg("Rabbit consumer closed - critical Error")
			os.Exit(1)
		}()
	}
	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
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

// Legacy methods kept for backward compatibility
func (conf Config) startLegacyProcessing() {
	rabUrl := fmt.Sprintf("amqp://%s:%s@%s", conf.RabUser, conf.RabPass, conf.RabHost)
	err := conf.GetConn(rabUrl)
	if err != nil {
		log.Fatal().Err(err)
	}
	conf.startConsumer(conf.RabQName, handler, 100)
}
func handler(conf Config, d amqp.Delivery) bool {
	if d.Body == nil {
		log.Error().Msg("Error, no message body")
		return false
	}
	bodyUnmarshal := pb.Message{}
	err := proto.Unmarshal(d.Body, &bodyUnmarshal)
	if err != nil {
		log.Error().Msg("Error with unmarshaling body")
		return false
	}
	conf.Deliver(bodyUnmarshal)
	//sendEmail Here
	return true

}

func (conf Config) Deliver(msg pb.Message) error {
	//set neccessary info for environment variables

	// Sender data.
	// Receiver email address.
	to := msg.OtherEmail
	// smtp server configuration.
	fullhost := fmt.Sprintf("%s:%d", conf.EmailHost, conf.EmailPort)
	// Authentication.
	auth := smtp.PlainAuth("", conf.EmailUser, conf.EmailPass, conf.EmailHost)

	t, err := template.ParseFiles("/templates/email_template.html")
	if err != nil {
		log.Error().Err(err).Msg("template not found")

		return err
	}

	mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body := []byte("From: Password Exchange <server@password.exchange>\r\n" + "To: " + to + "\r\n" +
		fmt.Sprintf("Subject: Encrypted Messsage from Password exchange from %s \r\n", msg.Firstname) +
		mimeHeaders)
	buf := bytes.NewBuffer(body)
	err = t.Execute(buf, struct {
		Body    string
		Message string
	}{
		Body:    fmt.Sprintf("Hi %s, \n %s used our service at <a href=\"https://password.exchange\"> Password Exchange </a> to send a secure message to you. We've included a link to view the message below, to find out more information go to https://password.exchange/about", msg.Otherfirstname, msg.Firstname),
		Message: msg.Content,
	})

	if err != nil {
		log.Error().Err(err).Msg("Something went wrong with rendering email template")
		return err
	}
	// Sending email.
	if err = smtp.SendMail(fullhost, auth, conf.EmailFrom, []string{to}, buf.Bytes()); err != nil {
		log.Error().Err(err).Msgf("emailhost: %s from: %s to: %s authHost: %s", conf.EmailHost, conf.EmailFrom, to, conf.EmailHost)
	}

	return err
}