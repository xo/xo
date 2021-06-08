package loader

import (
	"context"
	"fmt"
	"strings"

	"github.com/xo/xo/models"
	"github.com/xo/xo/templates/gotpl"
)

func init() {
	Register(&Loader{
		Driver: "sqlserver",
		Kind: map[Kind]string{
			KindTable: "U",
			KindView:  "V",
		},
		ParamN: func(i int) string {
			return fmt.Sprintf("@p%d", i+1)
		},
		MaskFunc: func() string {
			return "@p%d"
		},
		Schema:           models.SqlserverSchema,
		GoType:           SqlserverGoType,
		Tables:           SqlserverTables,
		TableColumns:     models.SqlserverTableColumns,
		TableForeignKeys: models.SqlserverTableForeignKeys,
		TableIndexes:     models.SqlserverTableIndexes,
		IndexColumns:     models.SqlserverIndexColumns,
		QueryColumns:     SqlserverQueryColumns,
	})
}

// SqlserverGoType parse a mssql type into a Go type based on the column
// definition.
func SqlserverGoType(ctx context.Context, typ string, nullable bool) (string, string, int, error) {
	// extract precision
	typ, prec, _, err := parsePrec(typ)
	if err != nil {
		return "", "", 0, err
	}
	var goType, zero string
	switch typ {
	case "tinyint", "bit":
		goType, zero = "bool", "false"
		if nullable {
			goType, zero = "sql.NullBool", "sql.NullBool{}"
		}
	case "char", "varchar", "text", "nchar", "nvarchar", "ntext", "smallmoney", "money":
		goType, zero = "string", `""`
		if nullable {
			goType, zero = "sql.NullString", "sql.NullString{}"
		}
	case "smallint":
		goType, zero = "int16", "0"
		if nullable {
			goType, zero = "sql.NullInt64", "sql.NullInt64{}"
		}
	case "int":
		goType, zero = gotpl.Int32(ctx), "0"
		if nullable {
			goType, zero = "sql.NullInt64", "sql.NullInt64{}"
		}
	case "bigint":
		goType, zero = "int64", "0"
		if nullable {
			goType, zero = "sql.NullInt64", "sql.NullInt64{}"
		}
	case "smallserial":
		goType, zero = "uint16", "0"
		if nullable {
			goType, zero = "sql.NullInt64", "sql.NullInt64{}"
		}
	case "serial":
		goType, zero = gotpl.Uint32(ctx), "0"
		if nullable {
			goType, zero = "sql.NullInt64", "sql.NullInt64{}"
		}
	case "bigserial":
		goType, zero = "uint64", "0"
		if nullable {
			goType, zero = "sql.NullInt64", "sql.NullInt64{}"
		}
	case "real":
		goType, zero = "float32", "0.0"
		if nullable {
			goType, zero = "sql.NullFloat64", "sql.NullFloat64{}"
		}
	case "numeric", "decimal":
		goType, zero = "float64", "0.0"
		if nullable {
			goType, zero = "sql.NullFloat64", "sql.NullFloat64{}"
		}
	case "binary", "varbinary":
		goType, zero = "[]byte", "nil"
	case "datetime", "datetime2", "timestamp":
		goType, zero = "time.Time", "time.Time{}"
	case "time with time zone", "time without time zone", "timestamp without time zone":
		goType, zero = "int64", "0"
		if nullable {
			goType, zero = "sql.NullInt64", "sql.NullInt64{}"
		}
	case "interval":
		goType, zero = "*time.Duration", "nil"
	default:
		goType, zero = schemaGoType(ctx, typ)
	}
	return goType, zero, prec, nil
}

// SqlserverTables returns the sqlserver tables with the manual PK information added.
// ManualPk is true when the table's primary key is not an identity.
func SqlserverTables(ctx context.Context, db models.DB, schema string, relkind string) ([]*models.Table, error) {
	// get the tables
	rows, err := models.SqlserverTables(ctx, db, schema, relkind)
	if err != nil {
		return nil, err
	}
	// get the tables that have Identity included
	identities, err := models.SqlserverIdentities(ctx, db, schema)
	if err != nil {
		// set it to an empty set on error
		identities = []*models.SqlserverIdentity{}
	}
	// add information about manual fk
	var tables []*models.Table
	for _, row := range rows {
		manualPk := true
		// look for a match in the table name where it contains the identity
		for _, identity := range identities {
			if identity.TableName == row.TableName {
				manualPk = false
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

// SqlserverQueryColumns parses the query and generates a type for it.
func SqlserverQueryColumns(ctx context.Context, db models.DB, schema string, inspect []string) ([]*models.Column, error) {
	// process inspect -- cannot have 'order by' in a CREATE VIEW
	ins := []string{}
	for _, l := range inspect {
		if !strings.HasPrefix(strings.ToUpper(l), "ORDER BY ") {
			ins = append(ins, l)
		}
	}
	// create temporary view xoid
	xoid := "_xo_" + randomID()
	viewq := `CREATE VIEW ` + xoid + ` AS ` + strings.Join(ins, "\n")
	models.Logf(viewq)
	if _, err := db.ExecContext(ctx, viewq); err != nil {
		return nil, err
	}
	// load columns
	cols, err := models.SqlserverTableColumns(ctx, db, schema, xoid)
	// drop inspect view
	dropq := `DROP VIEW ` + xoid
	models.Logf(dropq)
	_, _ = db.ExecContext(ctx, dropq)
	// load column information
	return cols, err
}
