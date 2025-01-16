package rabbitmq

import (
	"log"

	"github.com/streadway/amqp"
)

func DeclareQueues(channel *amqp.Channel, queues []string) error {
	for _, queue := range queues {
		_, err := channel.QueueDeclare(
			queue, // queue name
			true,  // durable
			false, // delete when unused
			false, // exclusive
			false, // no-wait
			nil,   // arguments
		)
		if err != nil {
			log.Printf("Failed to declare RabbitMQ queue '%s': %v", queue, err)
			return err
		}
		log.Printf("Queue '%s' declared successfully.", queue)
	}
	return nil
}
