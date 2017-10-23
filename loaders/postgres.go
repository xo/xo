package loaders

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	_ "github.com/lib/pq"

	"github.com/knq/snaker"
	"github.com/knq/xo/internal"
	"github.com/knq/xo/models"
)

func init() {
	internal.SchemaLoaders["postgres"] = internal.TypeLoader{
		ProcessRelkind: PgRelkind,
		Schema:         func(*internal.ArgType) (string, error) { return "public", nil },
		ParseType:      PgParseType,
		EnumList:       models.PgEnums,
		EnumValueList:  models.PgEnumValues,
		ProcList:       models.PgProcs,
		ProcParamList:  models.PgProcParams,
		TableList:      PgTables,
		ColumnList: func(db models.XODB, schema string, table string) ([]*models.Column, error) {
			return models.PgTableColumns(db, schema, table, internal.Args.EnablePostgresOIDs)
		},
		ForeignKeyList:  models.PgTableForeignKeys,
		IndexList:       models.PgTableIndexes,
		IndexColumnList: PgIndexColumns,
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
	nilVal := "nil"
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

	// extract precision
	dt, precision, _ = args.ParsePrecision(dt)

	var typ string
	switch dt {
	case "boolean":
		nilVal = "false"
		typ = "bool"
		if nullable {
			nilVal = "sql.NullBool{}"
			typ = "sql.NullBool"
		}

	case "character", "character varying", "text", "money", "inet":
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
	case "integer":
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
	case "numeric", "double precision":
		nilVal = "0.0"
		typ = "float64"
		if nullable {
			nilVal = "sql.NullFloat64{}"
			typ = "sql.NullFloat64"
		}

	case "bytea":
		asSlice = true
		typ = "byte"

	case "date", "timestamp with time zone", "time with time zone", "time without time zone", "timestamp without time zone":
		nilVal = "time.Time{}"
		typ = "time.Time"
		if nullable {
			nilVal = "pq.NullTime{}"
			typ = "pq.NullTime"
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
		nilVal = `uint8(0)`
		typ = "uint8"

	case `"any"`, "bit varying":
		asSlice = true
		typ = "byte"

	case "hstore":
		typ = "hstore.Hstore"

	case "uuid":
		nilVal = "uuid.New()"
		typ = "uuid.UUID"

	default:
		if strings.HasPrefix(dt, args.Schema+".") {
			// in the same schema, so chop off
			typ = snaker.SnakeToCamelIdentifier(dt[len(args.Schema)+1:])
			nilVal = typ + "(0)"
		} else {
			typ = snaker.SnakeToCamelIdentifier(dt)
			nilVal = typ + "{}"
		}
	}

	// special case for []slice
	if typ == "string" && asSlice {
		return precision, "StringSlice{}", "StringSlice"
	}

	// correct type if slice
	if asSlice {
		typ = "[]" + typ
		nilVal = "nil"
	}

	return precision, nilVal, typ
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

// PgTables returns the Postgres tables with the manual PK information added.
// ManualPk is true when the table does not have a sequence defined.
func PgTables(db models.XODB, schema string, relkind string) ([]*models.Table, error) {
	var err error

	// get the tables
	rows, err := models.PgTables(db, schema, relkind)
	if err != nil {
		return nil, err
	}

	// Get the tables that have a sequence defined.
	sequences, err := models.PgSequences(db, schema)
	if err != nil {
		// Set it to an empty set on error.
		sequences = []*models.Sequence{}
	}

	// Add information about manual FK.
	var tables []*models.Table
	for _, row := range rows {
		manualPk := true
		// Look for a match in the table name where it contains the sequence
		for _, sequence := range sequences {
			if sequence.TableName == row.TableName {
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
	return models.PgTableColumns(args.DB, schema, xoid, false)
}

// PgIndexColumns returns the column list for an index.
func PgIndexColumns(db models.XODB, schema string, table string, index string) ([]*models.IndexColumn, error) {
	var err error

	// load columns
	cols, err := models.PgIndexColumns(db, schema, index)
	if err != nil {
		return nil, err
	}

	// load col order
	colOrd, err := models.PgGetColOrder(db, schema, index)
	if err != nil {
		return nil, err
	}

	// build schema name used in errors
	s := schema
	if s != "" {
		s = s + "."
	}

	// put cols in order using colOrder
	ret := []*models.IndexColumn{}
	for _, v := range strings.Split(colOrd.Ord, " ") {
		cid, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("could not convert %s%s index %s column %s to int", s, table, index, v)
		}

		// find column
		found := false
		var c *models.IndexColumn
		for _, ic := range cols {
			if cid == ic.Cid {
				found = true
				c = ic
				break
			}
		}

		// sanity check
		if !found {
			return nil, fmt.Errorf("could not find %s%s index %s column id %d", s, table, index, cid)
		}

		ret = append(ret, c)
	}

	return ret, nil
}
