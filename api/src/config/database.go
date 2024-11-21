package config

import (
	"fmt"
	"os"

	"github.com/jackc/pgx"
	db "iammati/statuspage/db"
)

func Database() *pgx.Conn {
	connConfig := pgx.ConnConfig{
		Host:     "statuspage-db",
		Port:     5432,
		Database: "statuspage",
		User:     "statuspage",
		Password: "statuspage",
	}
	conn, err := pgx.Connect(connConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	db.Migrations(conn)

	return conn
}
