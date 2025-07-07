package config

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func loadDb(pgHost, pgPort, pgUser, pgPass, pgDbName string) (*sql.DB, error) {
	pgConn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", pgUser, pgPass, pgHost, pgPort, pgDbName)

	db, err := sql.Open("postgres", pgConn)
	if err != nil {
		return nil, fmt.Errorf("error in loading db: %s", err.Error())
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error in pinging db: %s", err.Error())
	}

	return db, nil
}
