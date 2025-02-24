package caches

import (
	"context"
	"time"
)

// Cache interface represents the required methods for the factory
type Cache interface {
	Connect(conn string) error
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	SetTTL(ctx context.Context, key string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exist(ctx context.Context, keys string) int64
	SetHash(ctx context.Context, key string, values map[string]interface{}) error
	GetHash(ctx context.Context, key string, field string) (string, error)
	InvalidateHash(ctx context.Context, key string) error
}
