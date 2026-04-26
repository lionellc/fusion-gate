package redis

import (
	"fmt"

	"github.com/lionellc/fusion-gate/internal/config"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	*redis.Client
}

func NewRedisClient(cfg *config.Config) (*RedisClient, func(), error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
	})
	rc := &RedisClient{Client: client}
	cleanup := func() {
		_ = rc.Close()
	}
	return rc, cleanup, nil
}

func (c *RedisClient) Close() error {
	return c.Client.Close()
}
