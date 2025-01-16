package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
)

type RedisManager struct {
	Client *redis.Client
}

func (r *RedisManager) AddTask(ctx context.Context, meetingID, taskID string) error {
	key := fmt.Sprintf("meeting:%s:tasks", meetingID)
	return r.Client.HSet(ctx, key, taskID, "pending").Err()
}

func (r *RedisManager) UpdateTaskStatus(ctx context.Context, meetingID, taskID, status string) error {
	key := fmt.Sprintf("meeting:%s:tasks", meetingID)
	return r.Client.HSet(ctx, key, taskID, status).Err()
}

func (r *RedisManager) AllTasksCompleted(ctx context.Context, meetingID string) (bool, error) {
	key := fmt.Sprintf("meeting:%s:tasks", meetingID)
	tasks, err := r.Client.HGetAll(ctx, key).Result()
	if err != nil {
		return false, err
	}
	for _, status := range tasks {
		if status != "completed" {
			return false, nil
		}
	}
	return true, nil
}
