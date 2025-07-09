package config

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func loadDb(pgUrl string) (*sql.DB, error) {
	db, err := sql.Open("postgres", pgUrl)
	if err != nil {
		return nil, fmt.Errorf("error in loading db: %s", err.Error())
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error in pinging db: %s", err.Error())
	}

	return db, nil
}
