package loaders

import (
	"regexp"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"github.com/knq/xo/internal"
	"github.com/knq/xo/models"
)

func init() {
	internal.SchemaLoaders["sqlite3"] = internal.TypeLoader{
		ProcessRelkind: SqRelkind,
		ParamN:         func(int) string { return "?" },
		MaskFunc:       func() string { return "?" },
		ParseType:      SqParseType,
		TableList:      SqTables,
		ColumnList:     SqTableColumns,
		ForeignKeyList: func(db models.XODB, schema string, table string) ([]*models.ForeignKey, error) {
			return models.SqTableForeignKeys(db, table)
		},
		IndexList: func(db models.XODB, schema string, table string) ([]*models.Index, error) {
			return models.SqTableIndexes(db, table)
		},
		IndexColumnList: func(db models.XODB, schema string, table string, index string) ([]*models.IndexColumn, error) {
			return models.SqIndexColumns(db, index)
		},
		QueryColumnList: SqQueryColumns,
	}
}

// SqRelkind returns the sqlite string representation for RelType.
func SqRelkind(relType internal.RelType) string {
	var s string
	switch relType {
	case internal.Table:
		s = "table"
	case internal.View:
		s = "view"
	default:
		panic("unsupported RelType")
	}
	return s
}

var uRE = regexp.MustCompile(`\s*unsigned\*`)

// SqParseType parse a sqlite type into a Go type based on the column
// definition.
func SqParseType(args *internal.ArgType, dt string, nullable bool) (int, string, string) {
	precision := 0
	nilVal := "nil"
	unsigned := false

	dt = strings.ToLower(dt)

	// extract precision
	dt, precision, _ = args.ParsePrecision(dt)

	if uRE.MatchString(dt) {
		unsigned = true
		uRE.ReplaceAllString(dt, "")
	}

	var typ string
	switch dt {
	case "bool", "boolean":
		nilVal = "false"
		typ = "bool"
		if nullable {
			nilVal = "sql.NullBool{}"
			typ = "sql.NullBool"
		}

	case "int", "integer", "tinyint", "smallint", "mediumint", "bigint":
		nilVal = "0"
		typ = args.Int32Type
		if nullable {
			nilVal = "sql.NullInt64{}"
			typ = "sql.NullInt64"
		}

	case "numeric", "real", "double", "float", "decimal":
		nilVal = "0.0"
		typ = "float64"
		if nullable {
			nilVal = "sql.NullFloat64{}"
			typ = "sql.NullFloat64"
		}

	case "blob":
		typ = "[]byte"

	case "timestamp", "datetime", "date", "timestamp with time zone", "time with time zone", "time without time zone", "timestamp without time zone":
		nilVal = "xoutil.SqTime{}"
		typ = "xoutil.SqTime"

	default:
		// case "varchar", "character", "varying character", "nchar", "native character", "nvarchar", "text", "clob", "datetime", "date", "time":
		nilVal = `""`
		typ = "string"
		if nullable {
			nilVal = "sql.NullString{}"
			typ = "sql.NullString"
		}
	}

	// if unsigned ...
	if internal.IntRE.MatchString(typ) && unsigned {
		typ = "u" + typ
	}

	return precision, nilVal, typ
}

// SqTables returns the sqlite tables with the manual PK information added.
// ManualPk is true when the table's primary key is not autoincrement.
func SqTables(db models.XODB, schema string, relkind string) ([]*models.Table, error) {
	var err error

	// get the tables
	rows, err := models.SqTables(db, relkind)
	if err != nil {
		return nil, err
	}

	// get the SQL for the Autoincrement detection
	autoIncrements, err := models.SqAutoIncrements(db)
	if err != nil {
		// Set it to an empty set on error.
		autoIncrements = []*models.SqAutoIncrement{}
	}

	// Add information about manual FK.
	var tables []*models.Table
	for _, row := range rows {
		manualPk := true
		// Look for a match in the table name where it contains the autoincrement
		// keyword for the given table in the SQL.
		for _, autoInc := range autoIncrements {
			lSQL := strings.ToLower(autoInc.SQL)
			if autoInc.TableName == row.TableName && strings.Contains(lSQL, "autoincrement") {
				manualPk = false
			} else {
				cols, err := SqTableColumns(db, schema, row.TableName)
				if err != nil {
					return nil, err
				}
				for _, col := range cols {
					if col.IsPrimaryKey == true {
						dt := strings.ToUpper(col.DataType)
						if dt == "INTEGER" {
							manualPk = false
						}
						break
					}
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

// SqTableColumns returns the sqlite table column info.
func SqTableColumns(db models.XODB, schema string, table string) ([]*models.Column, error) {
	var err error

	// grab
	rows, err := models.SqTableColumns(db, table)
	if err != nil {
		return nil, err
	}

	// fix columns
	var cols []*models.Column
	for _, row := range rows {
		cols = append(cols, &models.Column{
			FieldOrdinal: row.FieldOrdinal,
			ColumnName:   row.ColumnName,
			DataType:     row.DataType,
			NotNull:      row.NotNull,
			DefaultValue: row.DefaultValue,
			IsPrimaryKey: row.PkColIndex != 0,
		})
	}

	return cols, nil
}

// SqQueryColumns parses a sqlite query and generates a type for it.
func SqQueryColumns(args *internal.ArgType, inspect []string) ([]*models.Column, error) {
	var err error

	// create temporary view xoid
	xoid := "_xo_" + internal.GenRandomID()
	viewq := `CREATE TEMPORARY VIEW ` + xoid + ` AS ` + strings.Join(inspect, "\n")
	models.XOLog(viewq)
	_, err = args.DB.Exec(viewq)
	if err != nil {
		return nil, err
	}

	// load column information
	return SqTableColumns(args.DB, "", xoid)
}
