package loader

import (
	"github.com/xo/xo/models"
	xo "github.com/xo/xo/types"
)

func init() {
	Register("sqlite3", Loader{
		Mask:             "$%d",
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
func Sqlite3GoType(d xo.Type, schema, itype, utype string) (string, string, error) {
	var goType, zero string
	switch d.Type {
	case "bool", "boolean":
		goType, zero = "bool", "false"
		if d.Nullable {
			goType, zero = "sql.NullBool", "sql.NullBool{}"
		}
	case "int", "integer", "tinyint", "smallint", "mediumint":
		goType, zero = itype, "0"
		if d.Nullable {
			goType, zero = "sql.NullInt64", "sql.NullInt64{}"
		}
	case "bigint":
		goType, zero = "int64", "0"
		if d.Nullable {
			goType, zero = "sql.NullInt64", "sql.NullInt64{}"
		}
	case "numeric", "real", "double", "float", "decimal":
		goType, zero = "float64", "0.0"
		if d.Nullable {
			goType, zero = "sql.NullFloat64", "sql.NullFloat64{}"
		}
	case "blob":
		goType, zero = "[]byte", "nil"
	case "timestamp", "datetime", "date", "timestamp with timezone", "time with timezone", "time without timezone", "timestamp without timezone":
		goType, zero = "Time", "Time{}"
		if d.Nullable {
			goType, zero = "*Time", "nil"
		}
	default:
		// case "varchar", "character", "varying character", "nchar", "native character", "nvarchar", "text", "clob", "time":
		goType, zero = "string", `""`
		if d.Nullable {
			goType, zero = "sql.NullString", "sql.NullString{}"
		}
	}
	// if unsigned ...
	if intRE.MatchString(goType) && d.Unsigned {
		if goType == itype {
			goType, zero = utype, "0"
		} else {
			goType = "u" + goType
		}
	}
	return goType, zero, nil
}
