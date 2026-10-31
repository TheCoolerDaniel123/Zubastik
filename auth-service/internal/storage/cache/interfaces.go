package cache

import (
	"context"
	"time"
)

type CodeDatabase interface {
	SetValue(ctx context.Context, key, value string, ttl time.Duration) error
	GetValue(ctx context.Context, phone string) (*string, error)
	DeleteValue(ctx context.Context, key string) error
}
