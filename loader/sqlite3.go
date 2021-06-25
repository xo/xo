package loader

import (
	"context"
	"regexp"

	"github.com/xo/xo/models"
	xo "github.com/xo/xo/types"
)

func init() {
	Register(&Loader{
		Driver:           "sqlite3",
		Mask:             "$%d",
		GoType:           Sqlite3GoType,
		Schema:           models.Sqlite3Schema,
		Tables:           models.Sqlite3Tables,
		TableColumns:     models.Sqlite3TableColumns,
		TableSequences:   models.Sqlite3TableSequences,
		TableForeignKeys: models.Sqlite3TableForeignKeys,
		TableIndexes:     models.Sqlite3TableIndexes,
		IndexColumns:     models.Sqlite3IndexColumns,
		ViewCreate:       models.Sqlite3ViewCreate,
		ViewDrop:         models.Sqlite3ViewDrop,
	})
}

// Sqlite3GoType parse a sqlite3 type into a Go type based on the column
// definition.
func Sqlite3GoType(ctx context.Context, d xo.Datatype) (string, string, error) {
	typ, nullable := d.Type, d.Nullable
	unsigned := false
	if unsignedRE.MatchString(typ) {
		unsigned = true
		unsignedRE.ReplaceAllString(typ, "")
	}
	var goType, zero string
	switch typ {
	case "bool", "boolean":
		goType, zero = "bool", "false"
		if nullable {
			goType, zero = "sql.NullBool", "sql.NullBool{}"
		}
	case "int", "integer", "tinyint", "smallint", "mediumint":
		goType, zero = Int32(ctx), "0"
		if nullable {
			goType, zero = "sql.NullInt64", "sql.NullInt64{}"
		}
	case "bigint":
		goType, zero = "int64", "0"
		if nullable {
			goType, zero = "sql.NullInt64", "sql.NullInt64{}"
		}
	case "numeric", "real", "double", "float", "decimal":
		goType, zero = "float64", "0.0"
		if nullable {
			goType, zero = "sql.NullFloat64", "sql.NullFloat64{}"
		}
	case "blob":
		goType, zero = "[]byte", "nil"
	case "timestamp", "datetime", "date", "timestamp with timezone", "time with timezone", "time without timezone", "timestamp without timezone":
		goType, zero = "Time", "Time{}"
		if nullable {
			goType, zero = "*Time", "nil"
		}
	default:
		// case "varchar", "character", "varying character", "nchar", "native character", "nvarchar", "text", "clob", "time":
		goType, zero = "string", `""`
		if nullable {
			goType, zero = "sql.NullString", "sql.NullString{}"
		}
	}
	// if unsigned ...
	if intRE.MatchString(goType) && unsigned && goType == Int32(ctx) {
		goType, zero = Uint32(ctx), "0"
	}
	return goType, zero, nil
}

// unsignedRE is the unsigned regexp.
var unsignedRE = regexp.MustCompile(`\s*unsigned\*`)
