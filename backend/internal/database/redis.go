package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// DefaultRedisDB is the default Redis database number
	DefaultRedisDB = 0
	
	// DefaultRedisPoolSize is the default connection pool size
	DefaultRedisPoolSize = 10
	
	// DefaultRedisDialTimeout is the default dial timeout
	DefaultRedisDialTimeout = 5 * time.Second
	
	// DefaultRedisReadTimeout is the default read timeout
	DefaultRedisReadTimeout = 3 * time.Second
	
	// DefaultRedisWriteTimeout is the default write timeout
	DefaultRedisWriteTimeout = 3 * time.Second
)

// NewRedisClient creates a new Redis client with connection pooling
func NewRedisClient(url string) (*redis.Client, error) {
	// Parse Redis URL
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	// Configure client options
	opt.PoolSize = DefaultRedisPoolSize
	opt.DialTimeout = DefaultRedisDialTimeout
	opt.ReadTimeout = DefaultRedisReadTimeout
	opt.WriteTimeout = DefaultRedisWriteTimeout
	opt.MaxRetries = 3
	opt.MinRetryBackoff = 100 * time.Millisecond
	opt.MaxRetryBackoff = 3 * time.Second

	// Create client
	client := redis.NewClient(opt)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("Successfully connected to Redis")

	return client, nil
}

// RedisHealthCheck performs a health check on the Redis connection
func RedisHealthCheck(ctx context.Context, client *redis.Client) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis health check failed: %w", err)
	}

	return nil
}

// CloseRedisConnection closes the Redis connection gracefully
func CloseRedisConnection(client *redis.Client) error {
	if err := client.Close(); err != nil {
		return fmt.Errorf("failed to close Redis connection: %w", err)
	}

	log.Println("Redis connection closed successfully")
	return nil
}

// RedisSet sets a key-value pair with optional expiration
func RedisSet(ctx context.Context, client *redis.Client, key string, value interface{}, expiration time.Duration) error {
	if err := client.Set(ctx, key, value, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set Redis key %s: %w", key, err)
	}
	return nil
}

// RedisGet gets a value by key
func RedisGet(ctx context.Context, client *redis.Client, key string) (string, error) {
	val, err := client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("key %s does not exist", key)
	} else if err != nil {
		return "", fmt.Errorf("failed to get Redis key %s: %w", key, err)
	}
	return val, nil
}

// RedisDelete deletes a key
func RedisDelete(ctx context.Context, client *redis.Client, keys ...string) error {
	if err := client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("failed to delete Redis keys: %w", err)
	}
	return nil
}

// RedisExists checks if a key exists
func RedisExists(ctx context.Context, client *redis.Client, key string) (bool, error) {
	val, err := client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check Redis key existence: %w", err)
	}
	return val > 0, nil
}

// RedisTTL gets the remaining TTL of a key
func RedisTTL(ctx context.Context, client *redis.Client, key string) (time.Duration, error) {
	ttl, err := client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL for key %s: %w", key, err)
	}
	return ttl, nil
}

// RedisPublish publishes a message to a channel
func RedisPublish(ctx context.Context, client *redis.Client, channel string, message interface{}) error {
	if err := client.Publish(ctx, channel, message).Err(); err != nil {
		return fmt.Errorf("failed to publish to channel %s: %w", channel, err)
	}
	return nil
}

// RedisSubscribe subscribes to channels
func RedisSubscribe(ctx context.Context, client *redis.Client, channels ...string) *redis.PubSub {
	return client.Subscribe(ctx, channels...)
}

// RedisIncr increments a key
func RedisIncr(ctx context.Context, client *redis.Client, key string) (int64, error) {
	val, err := client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment key %s: %w", key, err)
	}
	return val, nil
}

// RedisExpire sets expiration on a key
func RedisExpire(ctx context.Context, client *redis.Client, key string, expiration time.Duration) error {
	if err := client.Expire(ctx, key, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set expiration on key %s: %w", key, err)
	}
	return nil
}

// RedisSAdd adds members to a set
func RedisSAdd(ctx context.Context, client *redis.Client, key string, members ...interface{}) error {
	if err := client.SAdd(ctx, key, members...).Err(); err != nil {
		return fmt.Errorf("failed to add members to set %s: %w", key, err)
	}
	return nil
}

// RedisSRem removes members from a set
func RedisSRem(ctx context.Context, client *redis.Client, key string, members ...interface{}) error {
	if err := client.SRem(ctx, key, members...).Err(); err != nil {
		return fmt.Errorf("failed to remove members from set %s: %w", key, err)
	}
	return nil
}

// RedisSMembers gets all members of a set
func RedisSMembers(ctx context.Context, client *redis.Client, key string) ([]string, error) {
	members, err := client.SMembers(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get members of set %s: %w", key, err)
	}
	return members, nil
}

