// Command northwind demonstrates using generated models for the northwind
// sample database.
//
// Schema/data comes from the Yugabyte Database Git repository. See README.md
// for more information.
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"os/user"

	// drivers
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/godror/godror"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	//_ "github.com/sijms/go-ora"

	// models
	"github.com/xo/xo/_examples/northwind/mysql"
	"github.com/xo/xo/_examples/northwind/oracle"
	"github.com/xo/xo/_examples/northwind/postgres"
	"github.com/xo/xo/_examples/northwind/sqlite3"
	"github.com/xo/xo/_examples/northwind/sqlserver"

	"github.com/xo/dburl"
	"github.com/xo/dburl/passfile"
)

func init() {
	old := dburl.Unregister("oracle")
	dburl.RegisterAlias("godror", "oracle")
	for _, alias := range old.Aliases {
		dburl.RegisterAlias("godror", alias)
	}
}

func main() {
	verbose := flag.Bool("v", false, "verbose")
	dsn := flag.String("dsn", "", "dsn")
	flag.Parse()
	if err := run(context.Background(), *verbose, *dsn); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, verbose bool, dsn string) error {
	if verbose {
		logger := func(s string, v ...interface{}) {
			fmt.Printf("-------------------------------------\nQUERY: %s\n  VAL: %v\n\n", s, v)
		}
		mysql.SetLogger(logger)
		oracle.SetLogger(logger)
		postgres.SetLogger(logger)
		sqlite3.SetLogger(logger)
		sqlserver.SetLogger(logger)
	}
	v, err := user.Current()
	if err != nil {
		return err
	}
	// parse url
	u, err := parse(dsn)
	if err != nil {
		return err
	}
	// open database
	db, err := passfile.OpenURL(u, v.HomeDir, "xopass")
	if err != nil {
		return err
	}
	var f func(context.Context, *sql.DB) error
	switch u.Driver {
	case "mysql":
		f = runMysql
	case "oracle", "godror":
		f = runOracle
	case "postgres":
		f = runPostgres
	case "sqlite3":
		f = runSqlite3
	case "sqlserver":
		f = runSqlserver
	}
	return f(ctx, db)
}

func parse(dsn string) (*dburl.URL, error) {
	v, err := dburl.Parse(dsn)
	if err != nil {
		return nil, err
	}
	switch v.Driver {
	case "mysql":
		q := v.Query()
		q.Set("parseTime", "true")
		v.RawQuery = q.Encode()
		return dburl.Parse(v.String())
	case "sqlite3":
		q := v.Query()
		q.Set("_loc", "auto")
		v.RawQuery = q.Encode()
		return dburl.Parse(v.String())
	}
	return v, nil
}
