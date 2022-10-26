// forms.go
package rabbitmq

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	// When reconnecting to the server after connection failure
	reconnectDelay = 5 * time.Second
)

var (
	ErrDisconnected = errors.New("disconnected from rabbitmq, trying to reconnect")
)

type RabClient struct {
	pushQueue     string
	connection    *amqp.Connection
	channel       *amqp.Channel
	done          <-chan interface{}
	notifyClose   chan *amqp.Error
	notifyConfirm chan amqp.Confirmation
	isConnected   bool
	alive         bool
	threads       int
	wg            *sync.WaitGroup
}

func NewRab(listenQueue, pushQueue, addr string, done <-chan interface{}) *RabClient {
	threads := runtime.GOMAXPROCS(0)
	if numCPU := runtime.NumCPU(); numCPU > threads {
		threads = numCPU
	}

	client := RabClient{
		pushQueue:     pushQueue,
		connection:    &amqp.Connection{},
		channel:       &amqp.Channel{},
		done:          done,
		notifyClose:   make(chan *amqp.Error),
		notifyConfirm: make(chan amqp.Confirmation),
		isConnected:   false,
		alive:         true,
		threads:       threads,
		wg:            &sync.WaitGroup{},
	}
	client.wg.Add(threads)

	go client.handleReconnect(addr)
	return &client
}

// handleReconnect will wait for a connection error on
// notifyClose, and then continuously attempt to reconnect.
func (c *RabClient) handleReconnect(listenQueue, addr string) {
	for c.alive {
		c.isConnected = false
		t := time.Now()
		fmt.Printf("Attempting to connect to rabbitMQ: %s\n", addr)
		var retryCount int
		for !c.connect(listenQueue, addr) {
			if !q.alive {
				return
			}
			select {
			case <-q.done:
				return
			case <-time.After(reconnectDelay + time.Duration(retryCount)*time.Second):
				c.logger.Printf("disconnected from rabbitMQ and failed to connect")
				retryCount++
			}
		}
		q.logger.Printf("Connected to rabbitMQ in: %vms", time.Since(t).Milliseconds())
		select {
		case <-c.done:
			return
		case <-c.notifyClose:
		}
	}
}

// connect will make a single attempt to connect to
// RabbitMq. It returns the success of the attempt.
func (c *RabClient) connect(listenQueue, addr string) bool {
	conn, err := amqp.Dial(addr)
	if err != nil {
		c.logger.Printf("failed to dial rabbitMQ server: %v", err)
		return false
	}
	ch, err := conn.Channel()
	if err != nil {
		c.logger.Printf("failed connecting to channel: %v", err)
		return false
	}
	ch.Confirm(false)
	_, err = ch.QueueDeclare(
		listenQueue,
		true,  // Durable
		false, // Delete when unused
		false, // Exclusive
		false, // No-wait
		nil,   // Arguments
	)
	if err != nil {
		c.logger.Printf("failed to declare listen queue: %v", err)
		return false
	}

	_, err = ch.QueueDeclare(
		c.pushQueue,
		true,  // Durable
		false, // Delete when unused
		false, // Exclusive
		false, // No-wait
		nil,   // Arguments
	)
	if err != nil {
		c.logger.Printf("failed to declare push queue: %v", err)
		return false
	}
	c.changeConnection(conn, ch)
	c.isConnected = true
	return true
}
func (c *RabClient) changeConnection(connection *amqp.Connection, channel *amqp.Channel) {
	c.connection = connection
	c.channel = channel
	c.notifyClose = make(chan *amqp.Error)
	c.notifyConfirm = make(chan amqp.Confirmation)
	c.channel.NotifyClose(c.notifyClose)
	c.channel.NotifyPublish(c.notifyConfirm)
}

// Push will push data onto the queue, and wait for a confirmation.
// If no confirms are received until within the resendTimeout,
// it continuously resends messages until a confirmation is received.
// This will block until the server sends a confirm.
func (c *RabClient) Push(data []byte) error {
	if !c.isConnected {
		return errors.New("failed to push push: not connected")
	}
	for {
		err := c.UnsafePush(data)
		if err != nil {
			if err == ErrDisconnected {
				continue
			}
			return err
		}
		select {
		case confirm := <-c.notifyConfirm:
			if confirm.Ack {
				return nil
			}
		case <-time.After(resendDelay):
		}
	}
}

// UnsafePush will push to the queue without checking for
// confirmation. It returns an error if it fails to connect.
// No guarantees are provided for whether the server will
// receive the message.
func (c *RabClient) UnsafePush(data []byte) error {
	if !c.isConnected {
		return ErrDisconnected
	}
	return c.channel.Publish(
		"",     // Exchange
		c.name, // Routing key
		false,  // Mandatory
		false,  // Immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        data,
		},
	)
}
