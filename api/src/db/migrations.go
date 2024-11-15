package db

import (
	"github.com/jackc/pgx"
	db_migrations "infraops.dev/statuspage-core/db/migrations"
)

func Migrations(conn *pgx.Conn) {
	db_migrations.Logs(conn)
}
