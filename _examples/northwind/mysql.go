package main

import (
	"context"
	"database/sql"

	models "github.com/xo/xo/_examples/northwind/mysql"
)

func runMysql(ctx context.Context, db *sql.DB) error {
	return models.Run()
}
