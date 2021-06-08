package main

import (
	"context"
	"database/sql"

	models "github.com/xo/xo/_examples/django/mysql"
)

func runMysql(ctx context.Context, db *sql.DB) error {
	return models.Run()
}
