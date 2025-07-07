package config

import (
	"database/sql"
	"fmt"
	"os"
)

type Config struct {
	Port string
	DB   *sql.DB
}

func LoadConfig() (*Config, error) {
	// if err := godotenv.Load("../.env"); err != nil {
	// 	return nil, fmt.Errorf("error loading env: %s", err.Error())
	// }

	dashboardPort := os.Getenv("DASHBOARD_PORT")
	pgHost := os.Getenv("PG_HOST")
	pgPort := os.Getenv("PG_PORT")
	pgUser := os.Getenv("PG_USER")
	pgPass := os.Getenv("PG_PASS")
	pgDbName := os.Getenv("PG_DBNAME")

	db, err := loadDb(pgHost, pgPort, pgUser, pgPass, pgDbName)
	if err != nil {
		return nil, fmt.Errorf("error in loading db:%s ", err.Error())
	}

	return &Config{
		Port: dashboardPort,
		DB:   db,
	}, nil

}
