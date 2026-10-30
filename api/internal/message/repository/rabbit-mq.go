package repository

import (
	"api/domain"
	"context"
	"encoding/json"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

const queueName = "southpark_messages"

type RabbitMQMessageRepository struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	mu         sync.Mutex
}

func NewRabbitMQMessageRepository(url string) (*RabbitMQMessageRepository, error) {
	connection, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	channel, err := connection.Channel()
	if err != nil {
		connection.Close()
		return nil, err
	}

	_, err = channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		channel.Close()
		connection.Close()
		return nil, err
	}

	return &RabbitMQMessageRepository{
		connection: connection,
		channel:    channel,
	}, nil
}

func (r *RabbitMQMessageRepository) Store(ctx context.Context, message domain.Message) (string, error) {
	body, err := json.Marshal(message)
	if err != nil {
		return "", err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	err = r.channel.PublishWithContext(
		ctx,
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
	if err != nil {
		return "", err
	}

	return "message published", nil
}

func (r *RabbitMQMessageRepository) Close() {
	if r.channel != nil {
		r.channel.Close()
	}

	if r.connection != nil {
		r.connection.Close()
	}
}
