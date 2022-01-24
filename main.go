// Command xo generates code from database schemas and custom queries. Works
// with PostgreSQL, MySQL, Microsoft SQL Server, Oracle Database, and SQLite3.
package main

//go:generate ./gen.sh models
//go:generate go generate ./internal

import (
	"context"
	"fmt"
	"os"

	// drivers
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/sijms/go-ora/v2"

	"github.com/xo/xo/cmd"
)

// version is the app version.
var version = "0.0.0-dev"

func main() {
	if err := cmd.Run(context.Background(), "xo", version, os.Args[1:]...); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
