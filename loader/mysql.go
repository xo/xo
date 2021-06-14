package loader

import (
	"context"
	"strings"

	"github.com/xo/xo/models"
	"github.com/xo/xo/templates/gotpl"
)

func init() {
	Register(&Loader{
		Driver: "mysql",
		Kind: map[Kind]string{
			KindTable: "BASE TABLE",
			KindView:  "VIEW",
		},
		ParamN: func(int) string {
			return "?"
		},
		MaskFunc: func() string {
			return "?"
		},
		Schema:           models.MysqlSchema,
		GoType:           MysqlGoType,
		Enums:            models.MysqlEnums,
		EnumValues:       MysqlEnumValues,
		Procs:            models.MysqlProcs,
		ProcParams:       models.MysqlProcParams,
		Tables:           MysqlTables,
		TableColumns:     models.MysqlTableColumns,
		TableForeignKeys: models.MysqlTableForeignKeys,
		TableIndexes:     models.MysqlTableIndexes,
		IndexColumns:     models.MysqlIndexColumns,
		QueryColumns:     MysqlQueryColumns,
	})
}

// MysqlGoType parse a mysql type into a Go type based on the column
// definition.
func MysqlGoType(ctx context.Context, typ string, nullable bool) (string, string, int, error) {
	// extract precision
	typ, prec, _, err := parsePrec(typ)
	if err != nil {
		return "", "", 0, err
	}
	// extract unsigned
	unsigned := false
	if strings.HasSuffix(typ, " unsigned") {
		typ, unsigned = typ[:len(typ)-len(" unsigned")], true
	}
	var goType, zero string
	switch typ {
	case "bit":
		switch {
		case prec == 1 && !nullable:
			goType, zero = "bool", "false"
		case prec == 1 && nullable:
			goType, zero = "sql.NullBool", "sql.NullBool{}"
		case prec <= 8 && !nullable:
			goType, zero = "uint8", "0"
		case prec <= 16 && !nullable:
			goType, zero = "uint16", "0"
		case prec <= 32 && !nullable:
			goType, zero = "uint32", "0"
		case nullable:
			goType, zero = "sql.NullInt64", "sql.NullInt64{}"
		default:
			goType, zero = "uint64", "0"
		}
	case "bool", "boolean":
		goType, zero = "bool", "false"
		if nullable {
			goType, zero = "sql.NullBool", "sql.NullBool{}"
		}
	case "char", "varchar", "tinytext", "text", "mediumtext", "longtext":
		goType, zero = "string", `""`
		if nullable {
			goType, zero = "sql.NullString", "sql.NullString{}"
		}
	case "tinyint":
		// force tinyint(1) as bool
		switch {
		case prec == 1 && !nullable:
			goType, zero = "bool", "false"
		case prec == 1 && nullable:
			goType, zero = "sql.NullBool", "sql.NullBool{}"
		case nullable:
			goType, zero = "sql.NullInt64", "sql.NullInt64{}"
		default:
			goType, zero = "int8", "0"
		}
	case "smallint", "year":
		goType, zero = "int16", "0"
		if nullable {
			goType, zero = "sql.NullInt64", "sql.NullInt64{}"
		}
	case "mediumint", "int", "integer":
		goType, zero = gotpl.Int32(ctx), "0"
		if nullable {
			goType, zero = "sql.NullInt64", "sql.NullInt64{}"
		}
	case "bigint":
		goType, zero = "int64", "0"
		if nullable {
			goType, zero = "sql.NullInt64", "sql.NullInt64{}"
		}
	case "float":
		goType, zero = "float32", "0.0"
		if nullable {
			goType, zero = "sql.NullFloat64", "sql.NullFloat64{}"
		}
	case "decimal", "double":
		goType, zero = "float64", "0.0"
		if nullable {
			goType, zero = "sql.NullFloat64", "sql.NullFloat64{}"
		}
	case "binary", "varbinary", "tinyblob", "blob", "mediumblob", "longblob":
		goType, zero = "[]byte", "nil"
	case "timestamp", "datetime", "date":
		goType, zero = "time.Time", "time.Time{}"
		if nullable {
			goType, zero = "sql.NullTime", "sql.NullTime{}"
		}
	case "time":
		// time is not supported by the MySQL driver. Can parse the string to time.Time in the user code.
		goType, zero = "string", `""`
		if nullable {
			goType, zero = "sql.NullString", "sql.NullString{}"
		}
	default:
		goType, zero = schemaGoType(ctx, typ)
	}
	// if unsigned ...
	if intRE.MatchString(goType) && unsigned && goType == gotpl.Int32(ctx) {
		goType, zero = gotpl.Uint32(ctx), "0"
	}
	return goType, zero, prec, nil
}

// MysqlEnumValues loads the enum values.
func MysqlEnumValues(ctx context.Context, db models.DB, schema string, enum string) ([]*models.EnumValue, error) {
	// load enum vals
	res, err := models.MysqlEnumValues(ctx, db, schema, enum)
	if err != nil {
		return nil, err
	}
	// process enum vals
	values := []*models.EnumValue{}
	for i, val := range strings.Split(res.EnumValues[1:len(res.EnumValues)-1], "','") {
		values = append(values, &models.EnumValue{
			EnumValue:  val,
			ConstValue: i + 1,
		})
	}
	return values, nil
}

// MysqlTables returns the mysql tables.
func MysqlTables(ctx context.Context, db models.DB, schema string, kind string) ([]*models.Table, error) {
	// get tables
	rows, err := models.MysqlTables(ctx, db, schema, kind)
	if err != nil {
		return nil, err
	}
	// add manual pk info for sequences
	sequences, err := models.MysqlSequences(ctx, db, schema)
	if err != nil {
		return nil, err
	}
	var tables []*models.Table
	for _, row := range rows {
		// Look for a match in the table name where it contains the autoincrement
		manualPk := true
		for _, seq := range sequences {
			if seq.TableName == row.TableName {
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

// MysqlQueryColumns parses the query and generates a type for it.
func MysqlQueryColumns(ctx context.Context, db models.DB, schema string, inspect []string) ([]*models.Column, error) {
	// create temporary view xoid
	xoid := "_xo_" + randomID()
	viewq := `CREATE VIEW ` + xoid + ` AS (` + strings.Join(inspect, "\n") + `)`
	models.Logf(viewq)
	if _, err := db.ExecContext(ctx, viewq); err != nil {
		return nil, err
	}
	// load columns
	cols, err := models.MysqlTableColumns(ctx, db, schema, xoid)
	// drop inspect view
	dropq := `DROP VIEW ` + xoid
	models.Logf(dropq)
	_, _ = db.ExecContext(ctx, dropq)
	// load column information
	return cols, err
}
