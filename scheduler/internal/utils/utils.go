package utils

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

func ReleaseRedisLock(redis *redis.Client, key string, value any) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// script for making sure only the locking owner releases it
	script := `
        if redis.call("get", KEYS[1]) == ARGV[1] then
            return redis.call("del", KEYS[1])
        else
            return 0
        end
    `
	_, err := redis.Eval(ctx, script, []string{key}, value).Result()
	return err
}
