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

	pgHost := os.Getenv("PG_HOST")
	pgPort := os.Getenv("PG_PORT")
	pgUser := os.Getenv("PG_USER")
	pgPass := os.Getenv("PG_PASS")
	pgDbName := os.Getenv("PG_DBNAME")
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")

	db, err := loadDb(pgHost, pgPort, pgUser, pgPass, pgDbName)
	if err != nil {
		return nil, err
	}

	redisClient, err := loadRedis(redisHost, redisPort)
	if err != nil {
		return nil, err
	}

	return &Config{
		DB:    db,
		Redis: redisClient,
	}, nil
}
