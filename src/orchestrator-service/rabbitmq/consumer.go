package rabbitmq

import (
	"log"

	"github.com/streadway/amqp"
)

type RabbitMQConsumer struct {
	Channel *amqp.Channel
}

func (r *RabbitMQConsumer) Consume(queue string, handler func(amqp.Delivery)) error {
	msgs, err := r.Channel.Consume(
		queue,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	go func() {
		for msg := range msgs {
			handler(msg)
		}
	}()

	log.Printf("Consuming messages from queue '%s'", queue)
	return nil
}
