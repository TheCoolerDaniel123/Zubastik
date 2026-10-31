package redis

import (
	"auth-service/internal/config"

	"github.com/go-redis/redis/v9"
)

type RedisClient struct {
	*redis.Client
}

func NewRedisClient(cacheCfg *config.Cache) *RedisClient {
	return &RedisClient{
		redis.NewClient(&redis.Options{
			Addr:     cacheCfg.Host + ":" + cacheCfg.Port,
			Password: cacheCfg.Password,
		}),
	}
}
