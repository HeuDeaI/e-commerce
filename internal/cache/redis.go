package cache

import (
	"context"
	"e-commerce/internal/config"
	"encoding/json"
	"time"

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

func (c *Cache) GetCache(ctx context.Context, key string, dest interface{}) error {
	cachedData, err := c.Client.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(cachedData), dest)
}

func (c *Cache) SetCache(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	serializedData, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.Client.Set(ctx, key, serializedData, ttl).Err()
}

func (c *Cache) DeleteCache(ctx context.Context, key string) error {
	return c.Client.Del(ctx, key).Err()
}

func (c *Cache) Close() error {
	return c.Client.Close()
}
