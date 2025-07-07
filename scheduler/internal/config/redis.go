package config

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func loadRedis(host, port string) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr: host + ":" + port,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("erorr in loading redis client: %s", err.Error())
	}

	return client, nil
}
