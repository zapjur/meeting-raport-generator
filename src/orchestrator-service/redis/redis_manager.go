package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"strings"
)

type RedisManager struct {
	Client *redis.Client
}

func (r *RedisManager) AddTask(ctx context.Context, meetingID, taskID string) error {
	key := fmt.Sprintf("meeting:%s:tasks", meetingID)
	return r.Client.HSet(ctx, key, taskID, "pending").Err()
}

func (r *RedisManager) UpdateTaskStatus(ctx context.Context, meetingId, taskId, status string) error {
	key := "meeting:" + meetingId + ":tasks"
	err := r.Client.HSet(ctx, key, taskId, status).Err()
	if err != nil {
		return err
	}

	log.Printf("Updated Redis for meeting_id=%s, task_id=%s, status=%s", meetingId, taskId, status)
	return nil
}

func (r *RedisManager) AllTasksCompleted(ctx context.Context, meetingID string) (bool, error) {
	key := fmt.Sprintf("meeting:%s:tasks", meetingID)
	tasks, err := r.Client.HGetAll(ctx, key).Result()
	if err != nil {
		return false, err
	}
	for _, status := range tasks {
		if status == "pending" {
			return false, nil
		}
	}
	return true, nil
}

func (r *RedisManager) AllTasksOfTypeCompleted(ctx context.Context, meetingID, taskType string) (bool, error) {
	key := fmt.Sprintf("meeting:%s:tasks", meetingID)
	tasks, err := r.Client.HGetAll(ctx, key).Result()
	if err != nil {
		return false, err
	}

	for taskID, status := range tasks {
		if taskTypeInTaskID(taskID, taskType) && status == "pending" {
			return false, nil
		}
	}
	return true, nil
}

func taskTypeInTaskID(taskID, taskType string) bool {
	parts := strings.Split(taskID, "-")
	if len(parts) < 2 {
		return false
	}
	return parts[1] == taskType
}

func (r *RedisManager) SetMeetingStatus(ctx context.Context, meetingID, status string) error {
	key := fmt.Sprintf("meeting:%s:status", meetingID)
	return r.Client.Set(ctx, key, status, 0).Err()
}

func (r *RedisManager) GetMeetingStatus(ctx context.Context, meetingID string) (string, error) {
	key := fmt.Sprintf("meeting:%s:status", meetingID)
	return r.Client.Get(ctx, key).Result()
}
