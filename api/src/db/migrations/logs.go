package db_migrations

import (
	"fmt"
	"os"

	"github.com/jackc/pgx"
)

type LogEntry struct {
	Timestamp string // Assuming ISO 8601 format: "2006-01-02T15:04:05Z07:00"
	Level     string
	Message   string
}

func Logs(conn *pgx.Conn) {
	createTableSQL := `CREATE TABLE IF NOT EXISTS logs (
		id SERIAL PRIMARY KEY,
		timestamp TIMESTAMPTZ NOT NULL,
		level VARCHAR(50),
		message TEXT
	); TRUNCATE public.logs;`

	_, err := conn.Exec(createTableSQL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logs table: %v\n", err)
		os.Exit(1)
	}
}

func InsertLogEntry(conn *pgx.Conn, entry LogEntry) {
	insertSQL := `INSERT INTO logs (timestamp, level, message) VALUES ($1, $2, $3)`

	_, err := conn.Exec(insertSQL, entry.Timestamp, entry.Level, entry.Message)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to insert log entry: %v\n", err)
		os.Exit(1)
	}
}
