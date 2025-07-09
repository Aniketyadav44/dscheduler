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
	pgUrl := os.Getenv("DB_URL")

	db, err := loadDb(pgUrl)
	if err != nil {
		return nil, fmt.Errorf("error in loading db:%s ", err.Error())
	}

	return &Config{
		Port: dashboardPort,
		DB:   db,
	}, nil

}
