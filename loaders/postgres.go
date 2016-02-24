package loaders

import (
	"regexp"
	"strconv"
	"strings"

	_ "github.com/lib/pq"

	"github.com/knq/xo/internal"
	"github.com/knq/xo/models"
)

func init() {
	internal.SchemaLoaders["postgres"] = internal.TypeLoader{
		Schemes:        []string{"postgres", "postgresql", "pgsql", "pg"},
		ProcessRelkind: PgRelkind,
		Schema:         func(*internal.ArgType) (string, error) { return "public", nil },
		ParseType:      PgParseType,
		EnumList:       models.PgEnums,
		EnumValueList:  models.PgEnumValues,
		ProcList:       models.PgProcs,
		ProcParamList:  models.PgProcParams,
		TableList:      models.PgTables,
		ColumnList:     models.PgTableColumns,
		ForeignKeyList: models.PgTableForeignKeys,
		IndexList:      models.PgTableIndexes,
		IndexColumnList: func(db models.XODB, schema string, table string, index string) ([]*models.IndexColumn, error) {
			return models.PgIndexColumns(db, schema, index)
		},
		QueryStrip:      PgQueryStrip,
		QueryColumnList: PgQueryColumns,
	}
}

// PgRelkind returns the postgres string representation for RelType.
func PgRelkind(relType internal.RelType) string {
	var s string
	switch relType {
	case internal.Table:
		s = "r"
	case internal.View:
		s = "v"
	default:
		panic("unsupported RelType")
	}
	return s
}

// PgParseType parse a postgres type into a Go type based on the column
// definition.
func PgParseType(args *internal.ArgType, dt string, nullable bool) (int, string, string) {
	precision := 0
	nilType := "nil"
	asSlice := false

	// handle SETOF
	if strings.HasPrefix(dt, "SETOF ") {
		_, _, t := PgParseType(args, dt[len("SETOF "):], false)
		return 0, "nil", "[]" + t
	}

	// determine if it's a slice
	if strings.HasSuffix(dt, "[]") {
		dt = dt[:len(dt)-2]
		asSlice = true
	}

	// extract length
	if loc := internal.LenRE.FindStringIndex(dt); len(loc) != 0 {
		precision, _ = strconv.Atoi(dt[loc[0]+1 : loc[1]-1])
		dt = dt[:loc[0]]
	}

	var typ string
	switch dt {
	case "boolean":
		nilType = "false"
		typ = "bool"
		if nullable {
			nilType = "sql.NullBool{}"
			typ = "sql.NullBool"
		}

	case "character", "character varying", "text":
		nilType = `""`
		typ = "string"
		if nullable {
			nilType = "sql.NullString{}"
			typ = "sql.NullString"
		}

	case "smallint":
		nilType = "0"
		typ = "int16"
		if nullable {
			nilType = "sql.NullInt64{}"
			typ = "sql.NullInt64"
		}
	case "integer":
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

	case "smallserial":
		nilType = "0"
		typ = "uint16"
		if nullable {
			nilType = "sql.NullInt64{}"
			typ = "sql.NullInt64"
		}
	case "serial":
		nilType = "0"
		typ = args.Uint32Type
		if nullable {
			nilType = "sql.NullInt64{}"
			typ = "sql.NullInt64"
		}
	case "bigserial":
		nilType = "0"
		typ = "uint64"
		if nullable {
			nilType = "sql.NullInt64{}"
			typ = "sql.NullInt64"
		}

	case "real":
		nilType = "0.0"
		typ = "float32"
		if nullable {
			nilType = "sql.NullFloat64{}"
			typ = "sql.NullFloat64"
		}
	case "double precision":
		nilType = "0.0"
		typ = "float64"
		if nullable {
			nilType = "sql.NullFloat64{}"
			typ = "sql.NullFloat64"
		}

	case "bytea":
		asSlice = true
		typ = "byte"

	case "timestamp with time zone":
		typ = "*time.Time"
		if nullable {
			nilType = "pq.NullTime{}"
			typ = "pq.NullTime"
		}

	case "time with time zone", "time without time zone", "timestamp without time zone":
		nilType = "0"
		typ = "int64"
		if nullable {
			nilType = "sql.NullInt64{}"
			typ = "sql.NullInt64"
		}

	case "interval":
		typ = "*time.Duration"

	case `"char"`, "bit":
		// FIXME: this needs to actually be tested ...
		// i think this should be 'rune' but I don't think database/sql
		// supports 'rune' as a type?
		//
		// this is mainly here because postgres's pg_catalog.* meta tables have
		// this as a type.
		//typ = "rune"
		nilType = `uint8(0)`
		typ = "uint8"

	case `"any"`, "bit varying":
		asSlice = true
		typ = "byte"

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

	return precision, nilType, typ
}

// pgQueryStripRE is the regexp to match the '::type AS name' portion in a query,
// which is a quirk/requirement of generating queries as is done in this
// package.
var pgQueryStripRE = regexp.MustCompile(`(?i)::[a-z][a-z0-9_\.]+\s+AS\s+[a-z][a-z0-9_\.]+`)

// PgQueryStrip strips stuff.
func PgQueryStrip(query []string, queryComments []string) {
	for i, l := range query {
		pos := pgQueryStripRE.FindStringIndex(l)
		if pos != nil {
			query[i] = l[:pos[0]] + l[pos[1]:]
			queryComments[i+1] = l[pos[0]:pos[1]]
		} else {
			queryComments[i+1] = ""
		}
	}
}

// PgQueryColumns parses the query and generates a type for it.
func PgQueryColumns(args *internal.ArgType, inspect []string) ([]*models.Column, error) {
	var err error

	// create temporary view xoid
	xoid := "_xo_" + internal.GenRandomID()
	viewq := `CREATE TEMPORARY VIEW ` + xoid + ` AS (` + strings.Join(inspect, "\n") + `)`
	models.XOLog(viewq)
	_, err = args.DB.Exec(viewq)
	if err != nil {
		return nil, err
	}

	// query to determine schema name where temporary view was created
	var nspq = `SELECT n.nspname ` +
		`FROM pg_class c ` +
		`JOIN pg_namespace n ON n.oid = c.relnamespace ` +
		`WHERE n.nspname LIKE 'pg_temp%' AND c.relname = $1`

	// run query
	var schema string
	models.XOLog(nspq, xoid)
	err = args.DB.QueryRow(nspq, xoid).Scan(&schema)
	if err != nil {
		return nil, err
	}

	// load column information
	return models.PgTableColumns(args.DB, schema, xoid)
}
