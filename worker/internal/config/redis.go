package config

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func loadRedis(url, pass string) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     url,
		Password: pass,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("erorr in loading redis client: %s", err.Error())
	}

	return client, nil
}
