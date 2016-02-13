package loaders

import (
	"database/sql"
	"errors"

	_ "github.com/mattn/go-sqlite3"

	"github.com/knq/xo/internal"
)

func init() {
	internal.SchemaLoaders["sqlite"] = internal.TypeLoader{
		Schemes:        []string{"sqlite", "sqlite3"},
		QueryFunc:      SqParseQuery,
		LoadSchemaFunc: SqLoadSchemaTypes,
	}
}

// SqLoadSchemaTypes loads the sqlite type definitions from a database.
func SqLoadSchemaTypes(args *internal.ArgType, db *sql.DB) error {
	return errors.New("not implemented")
}

// SqParseQuery parses a sqlite query and generates a type for it.
func SqParseQuery(args *internal.ArgType, db *sql.DB) error {
	return errors.New("not implemented")
}
