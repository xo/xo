package loaders

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	_ "github.com/denisenkom/go-mssqldb"

	"github.com/knq/xo/internal"
	"github.com/knq/xo/models"
)

func init() {
	internal.SchemaLoaders["mssql"] = internal.TypeLoader{
		Schemes:        []string{"mssql", "sqlserver", "ms"},
		ProcessDSN:     MsProcessDSN,
		MaskFunc:       func() string { return "$%d" },
		ProcessRelkind: MsRelkind,
		Schema:         MsSchema,
		ParseType:      MsParseType,
		//EnumList:       models.MsEnums,
		//EnumValueList:  models.MsEnumValues,
		//ProcList:       models.MsProcs,
		//ProcParamList:  models.MsProcParams,
		TableList:       models.MsTables,
		ColumnList:      models.MsTableColumns,
		ForeignKeyList:  models.MsTableForeignKeys,
		IndexList:       models.MsTableIndexes,
		IndexColumnList: models.MsIndexColumns,
		QueryColumnList: MsQueryColumns,
	}
}

// MsProcessDSN processes a mssql DSN.
func MsProcessDSN(u *url.URL, protocol string) string {
	var err error

	// build host or domain socket
	host := u.Host
	port := 1433
	dbname := u.Path[1:]

	// read port if present
	pos := strings.Index(host, ":")
	if pos != -1 {
		port, err = strconv.Atoi(host[pos+1:])
		if err != nil {
			panic("invalid port")
		}
		host = host[:pos]
	}

	// format dsn
	dsn := fmt.Sprintf("server=%s;port=%d;database=%s", host, port, dbname)

	// add user/pass
	if user := u.User.Username(); len(user) > 0 {
		dsn = dsn + ";user id=" + user
	}
	if pass, ok := u.User.Password(); ok {
		dsn = dsn + ";password=" + pass
	}

	// add params
	for k, v := range u.Query() {
		dsn = dsn + ";" + k + "=" + v[0]
	}

	return dsn
}

// MsSchema retrieves the name of the current schema.
func MsSchema(args *internal.ArgType) (string, error) {
	var err error

	// sql query
	const sqlstr = `SELECT SCHEMA_NAME()`

	var schema string

	// run query
	models.XOLog(sqlstr)
	err = args.DB.QueryRow(sqlstr).Scan(&schema)
	if err != nil {
		return "", err
	}

	return schema, nil
}

// MsRelkind returns the postgres string representation for RelType.
func MsRelkind(relType internal.RelType) string {
	var s string
	switch relType {
	case internal.Table:
		s = "U"
	case internal.View:
		s = "V"
	default:
		panic("unsupported RelType")
	}
	return s
}

// MsParseType parse a postgres type into a Go type based on the column
// definition.
func MsParseType(args *internal.ArgType, dt string, nullable bool) (int, string, string) {
	precision := 0
	nilVal := "nil"

	// extract precision
	dt, precision, _ = args.ParsePrecision(dt)

	var typ string
	switch dt {
	case "tinyint", "bit":
		nilVal = "false"
		typ = "bool"
		if nullable {
			nilVal = "sql.NullBool{}"
			typ = "sql.NullBool"
		}

	case "char", "varchar", "text", "nchar", "nvarchar", "ntext", "smallmoney", "money":
		nilVal = `""`
		typ = "string"
		if nullable {
			nilVal = "sql.NullString{}"
			typ = "sql.NullString"
		}

	case "smallint":
		nilVal = "0"
		typ = "int16"
		if nullable {
			nilVal = "sql.NullInt64{}"
			typ = "sql.NullInt64"
		}
	case "int":
		nilVal = "0"
		typ = args.Int32Type
		if nullable {
			nilVal = "sql.NullInt64{}"
			typ = "sql.NullInt64"
		}
	case "bigint":
		nilVal = "0"
		typ = "int64"
		if nullable {
			nilVal = "sql.NullInt64{}"
			typ = "sql.NullInt64"
		}

	case "smallserial":
		nilVal = "0"
		typ = "uint16"
		if nullable {
			nilVal = "sql.NullInt64{}"
			typ = "sql.NullInt64"
		}
	case "serial":
		nilVal = "0"
		typ = args.Uint32Type
		if nullable {
			nilVal = "sql.NullInt64{}"
			typ = "sql.NullInt64"
		}
	case "bigserial":
		nilVal = "0"
		typ = "uint64"
		if nullable {
			nilVal = "sql.NullInt64{}"
			typ = "sql.NullInt64"
		}

	case "real":
		nilVal = "0.0"
		typ = "float32"
		if nullable {
			nilVal = "sql.NullFloat64{}"
			typ = "sql.NullFloat64"
		}
	case "numeric", "decimal":
		nilVal = "0.0"
		typ = "float64"
		if nullable {
			nilVal = "sql.NullFloat64{}"
			typ = "sql.NullFloat64"
		}

	case "binary", "varbinary":
		typ = "[]byte"

	case "datetime", "datetime2", "timestamp":
		nilVal = "time.Time{}"
		typ = "time.Time"

	case "time with time zone", "time without time zone", "timestamp without time zone":
		nilVal = "0"
		typ = "int64"
		if nullable {
			nilVal = "sql.NullInt64{}"
			typ = "sql.NullInt64"
		}

	case "interval":
		typ = "*time.Duration"

	default:
		if strings.HasPrefix(dt, args.Schema+".") {
			// in the same schema, so chop off
			typ = internal.SnakeToCamel(dt[len(args.Schema)+1:])
			nilVal = typ + "(0)"
		} else {
			typ = internal.SnakeToCamel(dt)
			nilVal = typ + "{}"
		}
	}

	return precision, nilVal, typ
}

// MsQueryColumns parses the query and generates a type for it.
func MsQueryColumns(args *internal.ArgType, inspect []string) ([]*models.Column, error) {
	var err error

	// process inspect -- cannot have 'order by' in a CREATE VIEW
	ins := []string{}
	for _, l := range inspect {
		if !strings.HasPrefix(strings.ToUpper(l), "ORDER BY ") {
			ins = append(ins, l)
		}
	}

	// create temporary view xoid
	xoid := "_xo_" + internal.GenRandomID()
	viewq := `CREATE VIEW ` + xoid + ` AS ` + strings.Join(ins, "\n")
	models.XOLog(viewq)
	_, err = args.DB.Exec(viewq)
	if err != nil {
		return nil, err
	}

	// load columns
	cols, err := models.MsTableColumns(args.DB, args.Schema, xoid)

	// drop inspect view
	dropq := `DROP VIEW ` + xoid
	models.XOLog(dropq)
	_, _ = args.DB.Exec(dropq)

	// load column information
	return cols, err
}
