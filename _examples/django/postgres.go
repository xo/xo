package main

import (
	"context"
	"database/sql"

	models "github.com/mmmcorp/xo/_examples/django/postgres"
)

func runPostgres(ctx context.Context, db *sql.DB) error {
	return models.Run()
}
