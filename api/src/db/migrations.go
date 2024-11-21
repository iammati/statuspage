package db

import (
	"github.com/jackc/pgx"
	db_migrations "iammati/statuspage/db/migrations"
)

func Migrations(conn *pgx.Conn) {
	db_migrations.Logs(conn)
}
