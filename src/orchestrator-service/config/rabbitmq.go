package config

import (
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
)

func ConnectToRabbitMQ(uri string) (*amqp.Connection, *amqp.Channel, error) {
	const retries = 5
	const retryDelay = 5 * time.Second

	var conn *amqp.Connection
	var channel *amqp.Channel
	var err error

	for i := 0; i < retries; i++ {
		conn, err = amqp.Dial(uri)
		if err == nil {
			channel, err = conn.Channel()
			if err == nil {
				log.Println("Successfully connected to RabbitMQ.")
				return conn, channel, nil
			}
		}

		log.Printf("RabbitMQ connection failed (%d/%d): %v", i+1, retries, err)
		time.Sleep(retryDelay)
	}

	return nil, nil, fmt.Errorf("failed to connect to RabbitMQ after %d retries: %w", retries, err)
}
