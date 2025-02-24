package redis

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Singleton instance of redis cache client
var (
	instance *redisCache
	once     sync.Once
	ctx      = context.Background()
)

// redisCache implements Cache interface
type redisCache struct {
	client *redis.Client
}

func NewRedis(redisUrl string) (*redisCache, error) {
	var err error
	once.Do(func() {
		instance = &redisCache{}
		err = instance.Connect(redisUrl)
	})
	if err != nil {
		return nil, err
	}
	return instance, nil
}

// Connect initializes the Redis client
func (r *redisCache) Connect(conn string) error {
	opts, err := redis.ParseURL(conn)
	if err != nil {
		return fmt.Errorf("failed to parse Redis URL: %v", err)
	}

	r.client = redis.NewClient(opts)
	if err := r.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %v", err)
	}
	return nil
}

// Get retrieves a value from Redis
func (r *redisCache) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", nil
	}
	return val, err
}

// Set stores a value in Redis with a TTL
func (r *redisCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

// SetTTL sets the TTL for a key
func (r *redisCache) SetTTL(ctx context.Context, key string, ttl time.Duration) error {
	return r.client.Expire(ctx, key, ttl).Err()
}

// Delete removes a key from Redis
func (r *redisCache) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// Exist checks if a key exists in Redis
func (r *redisCache) Exist(ctx context.Context, key string) int64 {
	val, _ := r.client.Exists(ctx, key).Result()
	return val
}

// SetHash sets a hash field in Redis
func (r *redisCache) SetHash(ctx context.Context, key string, values map[string]interface{}) error {
	return r.client.HSet(ctx, key, values).Err()
}

func (r *redisCache) GetHash(ctx context.Context, key string, field string) (string, error) {
	val, err := r.client.HGet(ctx, key, field).Result()
	if errors.Is(err, redis.Nil) {
		return "", fmt.Errorf("cache miss: key=%s field=%s", key, field) // Return an error explicitly
	}
	return val, err
}

// InvalidateHash deletes an entire hash key
func (r *redisCache) InvalidateHash(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// Close the Redis connection
func (r *redisCache) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}
