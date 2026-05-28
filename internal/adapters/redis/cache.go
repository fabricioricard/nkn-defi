package redis

import (
	"context"
	"github.com/go-redis/redis/v8"
)

type Cache struct {
	client *redis.Client
}

func NewCache(rdb *redis.Client) *Cache {
	return &Cache{client: rdb}
}

// Métodos para implementar a interface ports/cache.Cache (vamos definir a interface)
func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	return "", nil
}
func (c *Cache) Set(ctx context.Context, key string, value interface{}) error {
	return nil
}