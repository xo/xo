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
		TableColumns:     oracleTableCols,
		TableSequences:   models.OracleTableSequences,
		TableForeignKeys: models.OracleTableForeignKeys,
		TableIndexes:     models.OracleTableIndexes,
		IndexColumns:     models.OracleIndexColumns,
		ViewCreate:       models.OracleViewCreate,
		ViewTruncate:     models.OracleViewTruncate,
		ViewDrop:         models.OracleViewDrop,
	})
}

func oracleTableCols(ctx context.Context, db models.DB, schema, table string) ([]*models.Column, error) {
	return models.OracleTableColumns(ctx, db, schema, table, schema, table)
}

// OracleGoType parse a oracle type into a Go type based on the column
// definition.
func OracleGoType(ctx context.Context, d xo.Datatype) (string, string, error) {
	var goType, zero string
	// strip remaining length (on things like timestamp)
	switch orLenRE.ReplaceAllString(d.Type, "") {
	case "char", "nchar", "varchar", "varchar2", "nvarchar2", "clob", "nclob", "rowid":
		goType, zero = "string", `""`
		if d.Nullable {
			goType, zero = "sql.NullString", "sql.NullString{}"
		}
	case "number":
		switch {
		case d.Prec == 0 && d.Scale == 0 && !d.Nullable:
			goType, zero = "int", "0"
		case d.Scale != 0 && !d.Nullable:
			goType, zero = "float64", "0.0"
		case d.Scale != 0 && d.Nullable:
			goType, zero = "sql.NullFloat64", "sql.NullFloat64{}"
		case !d.Nullable:
			goType, zero = "int64", "0"
		default:
			goType, zero = "sql.NullInt64", "sql.NullInt64{}"
		}
	case "float":
		goType, zero = "float64", "0.0"
		if d.Nullable {
			goType, zero = "sql.NullFloat64", "sql.NullFloat64{}"
		}
	case "date", "timestamp", "timestamp with time zone", "timestamp with local time zone":
		goType, zero = "time.Time", "time.Time{}"
		if d.Nullable {
			goType, zero = "sql.NullTime", "sql.NullTime{}"
		}
	case "blob", "long raw", "raw", "xmltype":
		goType, zero = "[]byte", "nil"
	default:
		goType, zero = SchemaGoType(ctx, d.Type, d.Nullable)
	}
	// handle bools
	switch {
	case goType == "int64" && d.Prec == 1 && !d.Nullable:
		goType, zero = "bool", "false"
	case goType == "sql.NullInt64" && d.Prec == 1 && d.Nullable:
		goType, zero = "sql.NullBool", "sql.NullBool{}"
	}
	return goType, zero, nil
}

// orLenRE is a regexp that matches lengths.
var orLenRE = regexp.MustCompile(`\([0-9]+\)`)
