// +build oracle

package loaders

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	_ "gopkg.in/rana/ora.v4"

	"github.com/knq/xo/internal"
	"github.com/knq/xo/models"
)

func init() {
	internal.SchemaLoaders["ora"] = internal.TypeLoader{
		ParamN:         func(i int) string { return fmt.Sprintf(":%d", i+1) },
		MaskFunc:       func() string { return ":%d" },
		ProcessRelkind: OrRelkind,
		Schema:         OrSchema,
		ParseType:      OrParseType,
		//EnumList:        models.OrEnums,
		//EnumValueList:   OrEnumValues,
		//ProcList:      models.OrProcs,
		//ProcParamList: models.OrProcParams,
		TableList:       models.OrTables,
		ColumnList:      models.OrTableColumns,
		ForeignKeyList:  models.OrTableForeignKeys,
		IndexList:       models.OrTableIndexes,
		IndexColumnList: models.OrIndexColumns,
		QueryColumnList: OrQueryColumns,
	}
}

// OrRelkind returns the oracle string representation for RelType.
func OrRelkind(relType internal.RelType) string {
	var s string
	switch relType {
	case internal.Table:
		s = "TABLE"
	case internal.View:
		s = "VIEW"
	default:
		panic("unsupported RelType")
	}
	return s
}

// OrSchema retrieves the name of the current schema.
func OrSchema(args *internal.ArgType) (string, error) {
	var err error

	// sql query
	const sqlstr = `SELECT LOWER(SYS_CONTEXT('USERENV', 'CURRENT_SCHEMA')) FROM dual`

	var schema string

	// run query
	models.XOLog(sqlstr)
	err = args.DB.QueryRow(sqlstr).Scan(&schema)
	if err != nil {
		return "", err
	}

	return schema, nil
}

// OrLenRE is a regexp that matches lengths.
var OrLenRE = regexp.MustCompile(`\([0-9]+\)`)

// OrParseType parse a oracle type into a Go type based on the column
// definition.
func OrParseType(args *internal.ArgType, dt string, nullable bool) (int, string, string) {
	nilVal := "nil"

	dt = strings.ToLower(dt)

	// extract precision
	dt, precision, scale := args.ParsePrecision(dt)

	var typ string
	// strip remaining length (on things like timestamp)
	switch OrLenRE.ReplaceAllString(dt, "") {
	case "char", "nchar", "varchar", "varchar2", "nvarchar2",
		"long",
		"clob", "nclob",
		"rowid":
		nilVal = `""`
		typ = "string"

	case "shortint":
		nilVal = "0"
		typ = "int16"
		if nullable {
			nilVal = "sql.NullInt64{}"
			typ = "sql.NullInt64"
		}
	case "integer":
		nilVal = "0"
		typ = args.Int32Type
		if nullable {
			nilVal = "sql.NullInt64{}"
			typ = "sql.NullInt64"
		}
	case "longinteger":
		nilVal = "0"
		typ = "int64"
		if nullable {
			nilVal = "sql.NullInt64{}"
			typ = "sql.NullInt64"
		}

	case "float", "shortdecimal":
		nilVal = "0.0"
		typ = "float32"
		if nullable {
			nilVal = "sql.NullFloat64{}"
			typ = "sql.NullFloat64"
		}

	case "number", "decimal":
		nilVal = "0.0"
		if 0 < precision && precision < 18 && scale > 0 {
			typ = "float64"
			if nullable {
				nilVal = "sql.NullFloat64{}"
				typ = "sql.NullFloat64"
			}
		} else if 0 < precision && precision <= 19 && scale == 0 {
			typ = "int64"
			if nullable {
				nilVal = "sql.NullInt64{}"
				typ = "sql.NullInt64"
			}
		} else {
			nilVal = `""`
			typ = "string"
		}

	case "blob", "long raw", "raw":
		typ = "[]byte"

	case "date", "timestamp", "timestamp with time zone":
		typ = "time.Time"
		nilVal = "time.Time{}"

	default:
		// bail
		fmt.Fprintf(os.Stderr, "error: unknown type %q\n", dt)
		os.Exit(1)
	}

	// special case for bool
	if typ == "int" && precision == 1 {
		nilVal = "false"
		typ = "bool"
		if nullable {
			nilVal = "sql.NullBool{}"
			typ = "sql.NullBool"
		}
	}

	return precision, nilVal, typ
}

// OrQueryColumns parses the query and generates a type for it.
func OrQueryColumns(args *internal.ArgType, inspect []string) ([]*models.Column, error) {
	var err error

	// create temporary view xoid
	xoid := "XO$" + internal.GenRandomID()
	viewq := `CREATE GLOBAL TEMPORARY TABLE ` + xoid + ` ` +
		`ON COMMIT PRESERVE ROWS ` +
		`AS ` + strings.Join(inspect, "\n")
	models.XOLog(viewq)
	_, err = args.DB.Exec(viewq)
	if err != nil {
		return nil, err
	}

	// load columns
	cols, err := models.OrTableColumns(args.DB, args.Schema, xoid)

	// drop inspect view
	dropq := `DROP TABLE ` + xoid
	models.XOLog(dropq)
	_, _ = args.DB.Exec(dropq)

	// load column information
	return cols, err
}
