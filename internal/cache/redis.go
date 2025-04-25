package cache

import (
	"context"
	"e-commerce/internal/config"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	Client *redis.Client
	config *config.RedisConfig
}

func New(ctx context.Context, cfg *config.RedisConfig) (*Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &Cache{
		Client: client,
		config: cfg,
	}, nil
}

func (c *Cache) Close() error {
	return c.Client.Close()
}
