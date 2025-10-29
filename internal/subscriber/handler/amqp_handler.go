package handler

import (
	"fmt"
	"log"
	"sync"

	"github.com/rabbitmq/amqp091-go"
	"github.com/spf13/viper"
)

type NotificationProcessor interface {
	ProcessNotification(body []byte)
}

type AMQPSubscriber struct {
	conn    *amqp091.Connection
	service NotificationProcessor
}

func NewAMQPSubscriber(conn *amqp091.Connection, service NotificationProcessor) *AMQPSubscriber {
	return &AMQPSubscriber{conn: conn, service: service}
}

func (s *AMQPSubscriber) Start(queueNames []string) error {
	ch, err := s.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %w", err)
	}

	if err := s.setupRetryQueue(ch); err != nil {
		return fmt.Errorf("failed to setup retry queue: %w", err)
	}

	var wg sync.WaitGroup
	log.Printf("Starting listeners for %d queues...", len(queueNames))
	for _, queueName := range queueNames {
		wg.Add(1)
		go func(qName string) {
			defer wg.Done()
			if err := s.listenOnQueue(ch, qName); err != nil {
				log.Printf("ERROR: Listener for queue '%s' stopped: %v", qName, err)
			}
		}(queueName)
	}
	wg.Wait()
	return nil
}

func (s *AMQPSubscriber) setupRetryQueue(ch *amqp091.Channel) error {
	retryExchange := viper.GetString("rabbitmq.retry_config.exchange")
	retryQueue := viper.GetString("rabbitmq.retry_config.queue")
	targetQueue := viper.GetString("rabbitmq.retry_config.routing_key")

	err := ch.ExchangeDeclare(retryExchange, "direct", true, false, false, false, nil)
	if err != nil {
		return err
	}

	q, err := ch.QueueDeclare(targetQueue, true, false, false, false, nil)
	if err != nil {
		return err
	}
	err = ch.QueueBind(q.Name, targetQueue, retryExchange, false, nil)
	if err != nil {
		return err
	}

	_, err = ch.QueueDeclare(retryQueue, true, false, false, false, amqp091.Table{
		"x-dead-letter-exchange":    retryExchange,
		"x-dead-letter-routing-key": targetQueue,
	})
	if err != nil {
		return err
	}

	log.Println("Successfully set up RabbitMQ retry queue and DLX.")
	return nil
}

func (s *AMQPSubscriber) listenOnQueue(ch *amqp091.Channel, queueName string) error {
	q, err := ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to declare queue '%s': %w", queueName, err)
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to register consumer for '%s': %w", q.Name, err)
	}

	log.Printf(" [*] Waiting for messages on queue: %s", q.Name)
	for d := range msgs {
		log.Printf("Received a message from queue '%s'", q.Name)
		s.service.ProcessNotification(d.Body)
	}
	return nil
}
