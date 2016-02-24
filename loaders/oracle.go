// +build oracle

package loaders

import (
	"database/sql"
	"errors"

	_ "github.com/mattn/go-oci8"

	"github.com/knq/xo/internal"
)

func init() {
	/*internal.SchemaLoaders["oracle"] = internal.TypeLoader{
		Schemes:        []string{"oracle", "ora", "oci", "oci8"},
		QueryFunc:      OraParseQuery,
		LoadSchemaFunc: OraLoadSchemaTypes,
	}*/
}

// OraLoadSchemaTypes loads the oracle type definitions from a database.
func OraLoadSchemaTypes(args *internal.ArgType, db *sql.DB) error {
	return errors.New("not implemented")
}

// OraParseQuery parses a oracle query and generates a type for it.
func OraParseQuery(args *internal.ArgType, db *sql.DB) error {
	return errors.New("not implemented")
}
