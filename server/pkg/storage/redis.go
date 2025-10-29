package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisStorage implements Storage using Redis
type RedisStorage struct {
	client *redis.Client
}

// NewRedis creates a new Redis storage instance
func NewRedis(cfg RedisConfig) (*RedisStorage, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.Database,
	})

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisStorage{client: client}, nil
}

// Get retrieves a value from storage
func (r *RedisStorage) Get(key string) (interface{}, error) {
	ctx := context.Background()
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	var result interface{}
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return nil, err
	}

	return result, nil
}

// Set stores a value in storage
func (r *RedisStorage) Set(key string, value interface{}) error {
	ctx := context.Background()
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, data, 0).Err()
}

// Delete removes a value from storage
func (r *RedisStorage) Delete(key string) error {
	ctx := context.Background()
	return r.client.Del(ctx, key).Err()
}

// List returns all key-value pairs
func (r *RedisStorage) List() map[string]interface{} {
	ctx := context.Background()
	keys, err := r.client.Keys(ctx, "*").Result()
	if err != nil {
		return make(map[string]interface{})
	}

	result := make(map[string]interface{})
	for _, key := range keys {
		val, err := r.client.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var data interface{}
		if err := json.Unmarshal([]byte(val), &data); err != nil {
			continue
		}
		result[key] = data
	}

	return result
}

// SaveAgent saves agent information
func (r *RedisStorage) SaveAgent(agent interface{}) error {
	ctx := context.Background()
	
	// Extract hostname for key
	hostname := getHostname(agent)
	if hostname == "" {
		return fmt.Errorf("hostname not found in agent data")
	}

	data, err := json.Marshal(agent)
	if err != nil {
		return err
	}

	// Store with TTL
	return r.client.Set(ctx, fmt.Sprintf("agent:%s", hostname), data, 24*time.Hour).Err()
}

// SaveHeartbeat saves heartbeat data
func (r *RedisStorage) SaveHeartbeat(agentID string, heartbeat interface{}) error {
	ctx := context.Background()
	
	data, err := json.Marshal(heartbeat)
	if err != nil {
		return err
	}

	// Store with short TTL (1 hour)
	key := fmt.Sprintf("heartbeat:%s:%d", agentID, time.Now().Unix())
	return r.client.Set(ctx, key, data, time.Hour).Err()
}

// GetAgents retrieves all agents
func (r *RedisStorage) GetAgents(filter interface{}) ([]interface{}, error) {
	ctx := context.Background()
	
	keys, err := r.client.Keys(ctx, "agent:*").Result()
	if err != nil {
		return nil, err
	}

	var results []interface{}
	for _, key := range keys {
		val, err := r.client.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var agent interface{}
		if err := json.Unmarshal([]byte(val), &agent); err != nil {
			continue
		}
		results = append(results, agent)
	}

	return results, nil
}

// Close closes the Redis connection
func (r *RedisStorage) Close() error {
	return r.client.Close()
}
