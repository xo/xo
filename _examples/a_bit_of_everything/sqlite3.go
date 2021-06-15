package main

import (
	"context"
	"database/sql"

	models "github.com/mmmcorp/xo/_examples/a_bit_of_everything/sqlite3"
)

func runSqlite3(ctx context.Context, db *sql.DB) error {
	return models.Run()
}
