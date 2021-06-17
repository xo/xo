package loader

import (
	"context"
	"regexp"
	"strings"

	"github.com/xo/xo/models"
)

func init() {
	Register(&Loader{
		Driver:           "oracle",
		Mask:             ":%d",
		GoType:           OracleGoType,
		Schema:           models.OracleSchema,
		Tables:           models.OracleTables,
		TableColumns:     models.OracleTableColumns,
		TableSequences:   models.OracleTableSequences,
		TableForeignKeys: models.OracleTableForeignKeys,
		TableIndexes:     models.OracleTableIndexes,
		IndexColumns:     models.OracleIndexColumns,
		ViewPrefix:       "XO$",
		QueryColumns:     OracleQueryColumns,
		/*
			ViewCreate:       models.OracleViewCreate,
			ViewDrop:         models.OracleViewDrop,
		*/
	})
}

// orLenRE is a regexp that matches lengths.
var orLenRE = regexp.MustCompile(`\([0-9]+\)`)

// OracleGoType parse a oracle type into a Go type based on the column
// definition.
func OracleGoType(ctx context.Context, typ string, nullable bool) (string, string, int, error) {
	// extract precision
	typ, prec, scale, err := parsePrec(typ)
	if err != nil {
		return "", "", 0, err
	}
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
	case "date", "timestamp", "timestamp with time zone":
		goType, zero = "time.Time", "time.Time{}"
		if nullable {
			goType, zero = "sql.NullTime", "sql.NullTime{}"
		}
	case "blob", "long raw", "raw":
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
	return goType, zero, prec, nil
}

// OracleQueryColumns parses the query and generates a type for it.
func OracleQueryColumns(ctx context.Context, db models.DB, schema string, inspect []string) ([]*models.Column, error) {
	// create temporary view xoid
	xoid := "XO$" + randomID()
	viewq := `CREATE GLOBAL TEMPORARY TABLE ` + xoid + ` ` +
		`ON COMMIT PRESERVE ROWS ` +
		`AS ` + strings.Join(inspect, "\n")
	models.Logf(viewq)
	if _, err := db.ExecContext(ctx, viewq); err != nil {
		return nil, err
	}
	// load columns
	cols, err := models.OracleTableColumns(ctx, db, schema, xoid)
	// drop inspect view
	dropq := `DROP TABLE ` + xoid
	models.Logf(dropq)
	_, _ = db.ExecContext(ctx, dropq)
	// load column information
	return cols, err
}
