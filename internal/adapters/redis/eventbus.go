package redis

import (
	"github.com/go-redis/redis/v8"
)

type RedisEventBus struct {
	client *redis.Client
}

func NewRedisEventBus(rdb *redis.Client) *RedisEventBus {
	return &RedisEventBus{client: rdb}
}

func (eb *RedisEventBus) Publish(topic string, payload interface{}) error {
	// implementação real com XAdd mais tarde
	return nil
}