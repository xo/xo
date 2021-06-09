package main

import (
	"context"
	"database/sql"

	models "github.com/xo/xo/_examples/northwind/sqlserver"
)

func runSqlserver(ctx context.Context, db *sql.DB) error {
	return models.Run()
}
