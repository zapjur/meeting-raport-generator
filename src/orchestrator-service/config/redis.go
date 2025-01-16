package config

import (
	"fmt"
	"github.com/go-redis/redis/v8"
)

func ConnectToRedis() (*redis.Client, error) {
	addr := "redis:6379"

	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	_, err := client.Ping(client.Context()).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}
	return client, nil
}
