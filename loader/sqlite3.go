package loader

import (
	"context"
	"regexp"
	"strings"

	"github.com/xo/xo/models"
	"github.com/xo/xo/templates/gotpl"
)

func init() {
	Register(&Loader{
		Driver: "sqlite3",
		Kind: map[Kind]string{
			KindTable: "table",
			KindView:  "view",
		},
		ParamN: func(int) string {
			return "?"
		},
		MaskFunc: func() string {
			return "?"
		},
		Schema:           models.Sqlite3Schema,
		GoType:           Sqlite3GoType,
		Tables:           Sqlite3Tables,
		TableColumns:     Sqlite3TableColumns,
		TableForeignKeys: Sqlite3ForeignKeys,
		TableIndexes:     Sqlite3Indexes,
		IndexColumns:     Sqlite3IndexColumns,
		QueryColumns:     Sqlite3QueryColumns,
	})
}

// unsignedRE is the unsigned regexp.
var unsignedRE = regexp.MustCompile(`\s*unsigned\*`)

// Sqlite3GoType parse a sqlite3 type into a Go type based on the column
// definition.
func Sqlite3GoType(ctx context.Context, typ string, nullable bool) (string, string, int, error) {
	// extract precision
	typ, prec, _, err := parsePrec(typ)
	if err != nil {
		return "", "", 0, err
	}
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
		goType, zero = gotpl.Int32(ctx), "0"
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
	case "timestamp", "datetime", "date", "timestamp with time zone", "time with time zone", "time without time zone", "timestamp without time zone":
		goType, zero = "Time", "Time{}"
	default:
		// case "varchar", "character", "varying character", "nchar", "native character", "nvarchar", "text", "clob", "datetime", "date", "time":
		goType, zero = "string", `""`
		if nullable {
			goType, zero = "sql.NullString", "sql.NullString{}"
		}
	}
	// if unsigned ...
	if intRE.MatchString(goType) && unsigned && goType == gotpl.Int32(ctx) {
		goType, zero = gotpl.Uint32(ctx), "0"
	}
	return goType, zero, prec, nil
}

// Sqlite3Tables returns the sqlite3 tables.
func Sqlite3Tables(ctx context.Context, db models.DB, schema string, kind string) ([]*models.Table, error) {
	// get tables
	rows, err := models.Sqlite3Tables(ctx, db, kind)
	if err != nil {
		return nil, err
	}
	// add manual pk info for sequences
	sequences, err := models.Sqlite3Sequences(ctx, db)
	if err != nil {
		return nil, err
	}
	// Add information about manual FK
	var tables []*models.Table
	for _, row := range rows {
		// Look for a match in the table name where it contains the autoincrement
		manualPk := true
		for _, seq := range sequences {
			if seq.TableName == row.TableName {
				manualPk = false
				break
			}
		}
		if manualPk {
			// check for primary key that is an integer
			cols, err := Sqlite3TableColumns(ctx, db, schema, row.TableName)
			if err != nil {
				return nil, err
			}
			for _, col := range cols {
				if col.IsPrimaryKey && strings.ToUpper(col.DataType) == "INTEGER" {
					manualPk = false
					break
				}
			}
		}
		tables = append(tables, &models.Table{
			TableName: row.TableName,
			Type:      row.Type,
			ManualPk:  manualPk,
		})
	}
	return tables, nil
}

// Sqlite3TableColumns returns the sqlite3 table column info.
func Sqlite3TableColumns(ctx context.Context, db models.DB, schema string, table string) ([]*models.Column, error) {
	return models.Sqlite3TableColumns(ctx, db, table)
}

// Sqlite3ForeignKeys returns the sqlite3 foreign key info.
func Sqlite3ForeignKeys(ctx context.Context, db models.DB, schema string, table string) ([]*models.ForeignKey, error) {
	return models.Sqlite3TableForeignKeys(ctx, db, table)
}

// Sqlite3Indexes returns the sqlite3 indexes info.
func Sqlite3Indexes(ctx context.Context, db models.DB, schema string, table string) ([]*models.Index, error) {
	return models.Sqlite3TableIndexes(ctx, db, table)
}

// Sqlite3IndexColumns returns the sqlite3 index column info.
func Sqlite3IndexColumns(ctx context.Context, db models.DB, schema string, table string, index string) ([]*models.IndexColumn, error) {
	return models.Sqlite3IndexColumns(ctx, db, index)
}

// Sqlite3QueryColumns parses a sqlite3 query and generates a type for it.
func Sqlite3QueryColumns(ctx context.Context, db models.DB, _ string, inspect []string) ([]*models.Column, error) {
	// create temporary view xoid
	xoid := "_xo_" + randomID()
	viewq := `CREATE TEMPORARY VIEW ` + xoid + ` AS ` + strings.Join(inspect, "\n")
	models.Logf(viewq)
	if _, err := db.ExecContext(ctx, viewq); err != nil {
		return nil, err
	}
	// load column information
	return Sqlite3TableColumns(ctx, db, "", xoid)
}
