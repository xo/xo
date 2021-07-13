package loader

import (
	"context"
	"regexp"

	"github.com/xo/xo/models"
	xo "github.com/xo/xo/types"
)

func init() {
	Register(&Loader{
		Driver:           "oracle",
		Mask:             ":%d",
		GoType:           OracleGoType,
		Schema:           models.OracleSchema,
		Procs:            models.OracleProcs,
		ProcParams:       models.OracleProcParams,
		Tables:           models.OracleTables,
		TableColumns:     models.OracleTableColumns,
		TableSequences:   models.OracleTableSequences,
		TableForeignKeys: models.OracleTableForeignKeys,
		TableIndexes:     models.OracleTableIndexes,
		IndexColumns:     models.OracleIndexColumns,
		ViewCreate:       models.OracleViewCreate,
		ViewTruncate:     models.OracleViewTruncate,
		ViewDrop:         models.OracleViewDrop,
	})
}

// OracleGoType parse a oracle type into a Go type based on the column
// definition.
func OracleGoType(ctx context.Context, d xo.Datatype) (string, string, error) {
	typ, nullable, prec, scale := d.Type, d.Nullable, d.Prec, d.Scale
	var goType, zero string
	// strip remaining length (on things like timestamp)
	switch orLenRE.ReplaceAllString(typ, "") {
	case "char", "nchar", "varchar", "varchar2", "nvarchar2", "clob", "nclob", "rowid":
		goType, zero = "string", `""`
		if nullable {
			goType, zero = "sql.NullString", "sql.NullString{}"
		}
	case "number":
		switch {
		case prec == 0 && scale == 0 && !nullable:
			goType, zero = "int", "0"
		case scale != 0 && !nullable:
			goType, zero = "float64", "0.0"
		case scale != 0 && nullable:
			goType, zero = "sql.NullFloat64", "sql.NullFloat64{}"
		case !nullable:
			goType, zero = "int64", "0"
		default:
			goType, zero = "sql.NullInt64", "sql.NullInt64{}"
		}
	case "float":
		goType, zero = "float64", "0.0"
		if nullable {
			goType, zero = "sql.NullFloat64", "sql.NullFloat64{}"
		}
	case "date", "timestamp", "timestamp with time zone", "timestamp with local time zone":
		goType, zero = "time.Time", "time.Time{}"
		if nullable {
			goType, zero = "sql.NullTime", "sql.NullTime{}"
		}
	case "blob", "long raw", "raw", "xmltype":
		goType, zero = "[]byte", "nil"
	default:
		goType, zero = schemaGoType(ctx, typ)
	}
	// handle bools
	switch {
	case goType == "int" && prec == 1 && !nullable:
		goType, zero = "bool", "false"
	case goType == "int" && prec == 1 && nullable:
		goType, zero = "sql.NullBool", "sql.NullBool{}"
	}
	return goType, zero, nil
}

// orLenRE is a regexp that matches lengths.
var orLenRE = regexp.MustCompile(`\([0-9]+\)`)
