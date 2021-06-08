// Command northwind demonstrates using generated models for the northwind
// sample database.
//
// Schema/data comes from the Yugabyte Database Git repository. See README.md
// for more information.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	// drivers
	_ "github.com/lib/pq"

	"github.com/xo/dburl/passfile"
	"github.com/xo/xo/_examples/northwind/models"
)

func main() {
	verbose := flag.Bool("v", false, "toggle verbose")
	dsn := flag.String("dsn", "", "database dsn")
	id := flag.Int("id", 16, "product id to retrieve")
	flag.Parse()
	if err := run(context.Background(), *verbose, *dsn, *id); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, verbose bool, dsn string, id int) error {
	if verbose {
		logger := func(s string, v ...interface{}) {
			fmt.Printf("-------------------------------------\nQUERY: %s\n  VAL: %v\n", s, v)
		}
		models.SetLogger(logger)
	}
	db, err := passfile.Open(dsn, "xopass")
	if err != nil {
		return err
	}
	p, err := models.ProductByProductID(ctx, db, int16(id))
	if err != nil {
		return err
	}
	fmt.Printf("product %d: %#v\n", id, p)
	return nil
}
