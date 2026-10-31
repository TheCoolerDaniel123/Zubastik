package redis

import (
	"auth-service/internal/storage/cache"
	"auth-service/pkg/clients/redis"
	"time"

	"context"

	r "github.com/go-redis/redis/v9"
)

type CodeRepo struct {
	cli *redis.RedisClient
}

func NewCodeRepo(cli *redis.RedisClient) *CodeRepo {
	return &CodeRepo{cli: cli}
}

// TODO: Check if TTL is provided
func (c *CodeRepo) SetValue(ctx context.Context, key, value string, ttl time.Duration) error {
	return c.cli.Set(ctx, key, value, ttl).Err()
}

func (c *CodeRepo) GetValue(ctx context.Context, phone string) (*string, error) {
	code, err := c.cli.Get(ctx, phone).Result()
	if err != nil {
		if err == r.Nil {
			return nil, cache.ErrCodeExpired
		}
		return nil, err
	}
	return &code, nil
}

func (c *CodeRepo) DeleteValue(ctx context.Context, key string) error {
	return c.cli.Del(ctx, key).Err()
}
