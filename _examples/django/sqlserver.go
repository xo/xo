package main

import (
	"context"
	"database/sql"

	models "github.com/xo/xo/_examples/django/sqlserver"
)

func runSqlserver(ctx context.Context, db *sql.DB) error {
	return models.Run()
}
