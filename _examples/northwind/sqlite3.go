package main

import (
	"context"
	"database/sql"
	"fmt"

	models "github.com/xo/xo/_examples/northwind/sqlite3"
)

func runSqlite3(ctx context.Context, db *sql.DB) error {
	p, err := models.ProductByProductID(ctx, db, 16)
	if err != nil {
		return err
	}
	fmt.Printf("product %d: %q\n", p.ProductID, p.ProductName)
	return nil
}
