package email

import (
	"fmt"
	"os"

	"github.com/Anthony-Bible/password-exchange/app/config"
	pb "github.com/Anthony-Bible/password-exchange/app/messagepb"
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
	return err
}
func (conn Config) startConsumer(queueName string, handler func(d amqp.Delivery) bool, concurrency int) {
	fmt.Printf("queuename: %s\n", conn.RabQName)
	q, err := conn.Channel.QueueDeclare(
		conn.RabQName, // name
		true,          // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	failOnError(err, "Failed to declare a queue")
	err = conn.Channel.Qos(
		100,   //prefetch count
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

	//var forever chan struct{}
	for i := 0; i < concurrency; i++ {
		go func() {
			for msg := range msgs {
				// if tha handler returns true then ACK, else NACK
				// the message back into the rabbit queue for
				// another round of processing
				if handler(msg) {
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

}

func (conf Config) StartProcessing() {
	rabUrl := fmt.Sprintf("amqp://%s:%s@%s", conf.RabUser, conf.RabPass, conf.RabHost)
	err := conf.GetConn(rabUrl)
	if err != nil {
		log.Fatal().Err(err)
	}
	fmt.Printf("Full config %+v", conf)
	conf.startConsumer(conf.RabQName, handler, 100)
	forever := make(chan bool)
	<-forever

}
func handler(d amqp.Delivery) bool {
	if d.Body == nil {
		log.Error().Msg("Error, no message body")
		return false
	}
	bodyUnmarshal := pb.Message{}
	err := proto.Unmarshal(d.Body, bodyUnmarshal)
	if err != nil {
		log.Error().Msg("Error with unmarshaling body")
	}
	//sendEmail Here
	return true

}
