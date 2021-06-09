package main

import (
	"context"
	"database/sql"

	models "github.com/xo/xo/_examples/northwind/sqlite3"
)

func runSqlite3(ctx context.Context, db *sql.DB) error {
	return models.Run()
}
