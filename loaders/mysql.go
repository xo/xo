package loaders

import (
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"

	"github.com/knq/xo/internal"
	"github.com/knq/xo/models"
)

func init() {
	internal.SchemaLoaders["mysql"] = internal.TypeLoader{
		Schemes:         []string{"mysql", "mariadb", "maria", "percona", "aurora", "my"},
		ProcessDSN:      MyProcessDSN,
		ParamN:          func(int) string { return "?" },
		MaskFunc:        func() string { return "?" },
		ProcessRelkind:  MyRelkind,
		Schema:          MySchema,
		ParseType:       MyParseType,
		EnumList:        models.MyEnums,
		EnumValueList:   MyEnumValues,
		ProcList:        models.MyProcs,
		ProcParamList:   models.MyProcParams,
		TableList:       models.MyTables,
		ColumnList:      models.MyTableColumns,
		ForeignKeyList:  models.MyTableForeignKeys,
		IndexList:       models.MyTableIndexes,
		IndexColumnList: models.MyIndexColumns,
		QueryColumnList: MyQueryColumns,
	}
}

// MyProcessDSN processes a mysql DSN.
func MyProcessDSN(u *url.URL, protocol string) string {
	// build host or domain socket
	host := u.Host
	dbname := u.Path[1:]
	if protocol == "unix" {
		host = path.Join(u.Host, path.Dir(u.Path))
		dbname = path.Base(u.Path)
	} else if !strings.Contains(host, ":") {
		// append default port
		host = host + ":3306"
	}

	// build user/pass
	userinfo := ""
	if un := u.User.Username(); len(un) > 0 {
		userinfo = un
		if up, ok := u.User.Password(); ok {
			userinfo = userinfo + ":" + up
		}
	}

	// build params
	params := u.Query().Encode()
	if len(params) > 0 {
		params = "?" + params
	}

	// format
	return fmt.Sprintf(
		"%s@%s(%s)/%s%s",
		userinfo,
		protocol,
		host,
		dbname,
		params,
	)
}

// MySchema retrieves the name of the current schema.
func MySchema(args *internal.ArgType) (string, error) {
	var err error

	// sql query
	const sqlstr = `SELECT SCHEMA()`

	var schema string

	// run query
	models.XOLog(sqlstr)
	err = args.DB.QueryRow(sqlstr).Scan(&schema)
	if err != nil {
		return "", err
	}

	return schema, nil
}

// MyRelkind returns the mysql string representation for RelType.
func MyRelkind(relType internal.RelType) string {
	var s string
	switch relType {
	case internal.Table:
		s = "BASE TABLE"
	case internal.View:
		s = "VIEW"
	default:
		panic("unsupported RelType")
	}
	return s
}

// MyParseType parse a mysql type into a Go type based on the column
// definition.
func MyParseType(args *internal.ArgType, dt string, nullable bool) (int, string, string) {
	precision := 0
	nilType := "nil"
	asSlice := false
	unsigned := false

	// extract unsigned
	if strings.HasSuffix(dt, " unsigned") {
		unsigned = true
		dt = dt[:len(dt)-len(" unsigned")]
	}

	// extract length
	if loc := internal.LenRE.FindStringIndex(dt); len(loc) != 0 {
		precision, _ = strconv.Atoi(dt[loc[0]+1 : loc[1]-1])
		dt = dt[:loc[0]]
	}

	var typ string
	switch dt {
	case "bool", "boolean":
		nilType = "false"
		typ = "bool"
		if nullable {
			nilType = "sql.NullBool{}"
			typ = "sql.NullBool"
		}

	case "char", "varchar", "tinytext", "text", "mediumtext", "longtext":
		nilType = `""`
		typ = "string"
		if nullable {
			nilType = "sql.NullString{}"
			typ = "sql.NullString"
		}

	case "tinyint", "smallint", "mediumint":
		nilType = "0"
		typ = "int16"
		if nullable {
			nilType = "sql.NullInt64{}"
			typ = "sql.NullInt64"
		}
	case "int", "integer":
		nilType = "0"
		typ = args.Int32Type
		if nullable {
			nilType = "sql.NullInt64{}"
			typ = "sql.NullInt64"
		}
	case "bigint":
		nilType = "0"
		typ = "int64"
		if nullable {
			nilType = "sql.NullInt64{}"
			typ = "sql.NullInt64"
		}

	case "decimal", "float":
		nilType = "0.0"
		typ = "float32"
		if nullable {
			nilType = "sql.NullFloat64{}"
			typ = "sql.NullFloat64"
		}
	case "double":
		nilType = "0.0"
		typ = "float64"
		if nullable {
			nilType = "sql.NullFloat64{}"
			typ = "sql.NullFloat64"
		}

	case "binary", "varbinary", "tinyblob", "blob", "mediumblob", "longblob":
		asSlice = true
		typ = "byte"

	case "timestamp", "datetime":
		typ = "*time.Time"
		if nullable {
			nilType = "pq.NullTime{}"
			typ = "pq.NullTime"
		}

	default:
		if strings.HasPrefix(dt, args.Schema+".") {
			// in the same schema, so chop off
			typ = internal.SnakeToCamel(dt[len(args.Schema)+1:])
			nilType = typ + "(0)"
		} else {
			typ = internal.SnakeToCamel(dt)
			nilType = typ + "{}"
		}
	}

	// special case for []slice
	if typ == "string" && asSlice {
		return precision, "StringSlice{}", "StringSlice"
	}

	// correct type if slice
	if asSlice {
		typ = "[]" + typ
		nilType = "nil"
	}

	// add 'u' as prefix to type if its unsigned
	// FIXME: this needs to be tested properly...
	if unsigned && internal.IntRE.MatchString(typ) {
		typ = "u" + typ
	}

	return precision, nilType, typ
}

// MyEnumValues loads the enum values.
func MyEnumValues(db models.XODB, schema string, enum string) ([]*models.EnumValue, error) {
	var err error

	// load enum vals
	res, err := models.MyEnumValues(db, schema, enum)
	if err != nil {
		return nil, err
	}

	// process enum vals
	enumVals := []*models.EnumValue{}
	for i, ev := range strings.Split(res.EnumValues[1:len(res.EnumValues)-1], "','") {
		enumVals = append(enumVals, &models.EnumValue{
			EnumValue:  ev,
			ConstValue: i + 1,
		})
	}

	return enumVals, nil
}

// MyQueryColumns parses the query and generates a type for it.
func MyQueryColumns(args *internal.ArgType, inspect []string) ([]*models.Column, error) {
	var err error

	// create temporary view xoid
	xoid := "_xo_" + internal.GenRandomID()
	viewq := `CREATE VIEW ` + xoid + ` AS (` + strings.Join(inspect, "\n") + `)`
	models.XOLog(viewq)
	_, err = args.DB.Exec(viewq)
	if err != nil {
		return nil, err
	}

	// load columns
	cols, err := models.MyTableColumns(args.DB, args.Schema, xoid)

	// drop inspect view
	dropq := `DROP VIEW ` + xoid
	models.XOLog(dropq)
	_, _ = args.DB.Exec(dropq)

	// load column information
	return cols, err
}
