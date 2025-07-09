package config

import (
	"database/sql"
	"os"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	DB    *sql.DB
	Redis *redis.Client
}

func LoadConfig() (*Config, error) {
	// if err := godotenv.Load("../.env"); err != nil {
	// 	return nil, fmt.Errorf("error in loading env: %s", err.Error())
	// }

	pgUrl := os.Getenv("DB_URL")
	redisUrl := os.Getenv("REDIS_URL")
	redisPass := os.Getenv("REDIS_PASS")

	db, err := loadDb(pgUrl)
	if err != nil {
		return nil, err
	}

	redisClient, err := loadRedis(redisUrl, redisPass)
	if err != nil {
		return nil, err
	}

	return &Config{
		DB:    db,
		Redis: redisClient,
	}, nil
}
