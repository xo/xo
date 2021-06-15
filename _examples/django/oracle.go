package main

import (
	"context"
	"database/sql"

	models "github.com/mmmcorp/xo/_examples/django/oracle"
)

func runOracle(ctx context.Context, db *sql.DB) error {
	return models.Run()
}
