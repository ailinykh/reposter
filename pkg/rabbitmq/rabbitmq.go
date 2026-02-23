package rabbitmq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func New(url string) (*Rabbit, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a RabbitMQ channel: %w", err)
	}

	return &Rabbit{
		ch: ch,
	}, nil
}

type Rabbit struct {
	ch *amqp.Channel
}

func (r *Rabbit) Consume(name string) (<-chan amqp.Delivery, error) {
	q, err := r.ch.QueueDeclare(
		name,  // queue name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)

	if err != nil {
		return nil, fmt.Errorf("failed to declare queue %w", err)
	}

	return r.ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
}

func (r *Rabbit) Publish(name string, body []byte) error {
	// FIXME: unnecessary queue declaration can introduce overhead
	q, err := r.ch.QueueDeclare(name, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("Failed to declare a callback queue: %w", err)
	}

	return r.ch.Publish(
		"",     // exchange
		q.Name, // routing key (queue name)
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		})
}
