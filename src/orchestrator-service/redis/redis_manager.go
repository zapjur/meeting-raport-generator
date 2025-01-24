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
		if status != "completed" {
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
		if taskTypeInTaskID(taskID, taskType) && status != "completed" {
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

func (r *RedisManager) SetMeetingEmail(ctx context.Context, meetingID, email string) error {
	key := fmt.Sprintf("meeting:%s:email", meetingID)
	err := r.Client.Set(ctx, key, email, 0).Err()
	if err != nil {
		return err
	}

	log.Printf("Saved email for meeting_id=%s: email=%s", meetingID, email)
	return nil
}

func (r *RedisManager) GetMeetingEmail(ctx context.Context, meetingID string) (string, error) {
	key := fmt.Sprintf("meeting:%s:email", meetingID)
	return r.Client.Get(ctx, key).Result()
}

func (r *RedisManager) DeleteAllMeetingEntries(ctx context.Context, meetingID string) error {
	pattern := fmt.Sprintf("meeting:%s:*", meetingID)
	var cursor uint64
	for {
		keys, nextCursor, err := r.Client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			log.Printf("Error during SCAN for meeting_id=%s: %v", meetingID, err)
			return err
		}

		if len(keys) > 0 {
			err = r.Client.Del(ctx, keys...).Err()
			if err != nil {
				log.Printf("Error deleting keys for meeting_id=%s: %v", meetingID, err)
				return err
			}
		}

		if nextCursor == 0 {
			break
		}
		cursor = nextCursor
	}

	log.Printf("Deleted all Redis entries for meeting_id=%s", meetingID)
	return nil
}
