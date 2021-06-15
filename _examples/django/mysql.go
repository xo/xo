package main

import (
	"context"
	"database/sql"

	models "github.com/mmmcorp/xo/_examples/django/mysql"
)

func runMysql(ctx context.Context, db *sql.DB) error {
	return models.Run()
}
