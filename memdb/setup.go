package memdb

import (
	"github.com/redis/go-redis/v9"
	"gitlab.com/MikeTTh/env"
)

var redisClient *redis.Client

func InitRedisConnection() error {
	redisClientOptions, err := redis.ParseURL(env.String("REDIS_URL", "redis://localhost:6379/0"))
	if err != nil {
		return err
	}

	redisClient = redis.NewClient(redisClientOptions)
	return nil
}
