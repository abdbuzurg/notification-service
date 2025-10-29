package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type AMQPPublisher struct {
	conn *amqp091.Connection
}

func NewAMQPPublisher(conn *amqp091.Connection) *AMQPPublisher {
	return &AMQPPublisher{conn: conn}
}

func (p *AMQPPublisher) Publish(body []byte, queueName string) error {
	return p.publishInternal(body, queueName, "")
}

func (p *AMQPPublisher) PublishWithTTL(body []byte, queueName string, ttl string) error {
	return p.publishInternal(body, queueName, ttl)
}

func (p *AMQPPublisher) publishInternal(body []byte, queueName string, ttl string) error {
	ch, err := p.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %w", err)
	}
	defer ch.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	props := amqp091.Publishing{
		ContentType: "application/json",
		Body:        body,
	}
	if ttl != "" {
		props.Expiration = ttl
	}

	err = ch.PublishWithContext(ctx, "", queueName, false, false, props)
	if err != nil {
		return fmt.Errorf("failed to publish to queue '%s': %w", queueName, err)
	}
	return nil
}
