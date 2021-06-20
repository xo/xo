package loader

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/xo/xo/models"
	"github.com/xo/xo/templates/gotpl"
)

func init() {
	Register(&Loader{
		Driver:           "postgres",
		Mask:             "$%d",
		Flags:            PostgresFlags,
		GoType:           PostgresGoType,
		Schema:           models.PostgresSchema,
		Enums:            models.PostgresEnums,
		EnumValues:       models.PostgresEnumValues,
		Procs:            models.PostgresProcs,
		ProcParams:       models.PostgresProcParams,
		Tables:           models.PostgresTables,
		TableColumns:     PostgresTableColumns,
		TableSequences:   models.PostgresTableSequences,
		TableForeignKeys: models.PostgresTableForeignKeys,
		TableIndexes:     models.PostgresTableIndexes,
		IndexColumns:     PostgresIndexColumns,
		ViewStrip:        PostgresViewStrip,
		ViewCreate:       models.PostgresViewCreate,
		ViewSchema:       models.PostgresViewSchema,
		ViewDrop:         models.PostgresViewDrop,
	})
}

// PostgresFlags returnss the postgres flags.
func PostgresFlags() []Flag {
	return []Flag{
		{
			ContextKey: OidsKey,
			Desc:       "enable postgres OIDs",
			Default:    "false",
			Value:      false,
		},
	}
}

// PostgresGoType parse a type into a Go type based on the database type definition.
func PostgresGoType(ctx context.Context, typ string, nullable bool) (string, string, int, error) {
	// SETOF -> []T
	if strings.HasPrefix(typ, "SETOF ") {
		goType, _, _, err := PostgresGoType(ctx, typ[len("SETOF "):], false)
		if err != nil {
			return "", "", 0, err
		}
		return "[]" + goType, "nil", 0, nil
	}
	// determine if it's a slice
	asSlice := false
	if strings.HasSuffix(typ, "[]") {
		typ, asSlice = typ[:len(typ)-2], true
	}
	// extract precision
	typ, prec, _, err := parsePrec(typ)
	if err != nil {
		return "", "", 0, err
	}
	var goType, zero string
	switch typ {
	case "boolean":
		goType, zero = "bool", "false"
		if nullable {
			goType, zero = "sql.NullBool", "sql.NullBool{}"
		}
	case "bpchar", "char", "character varying", "character", "inet", "money", "text":
		goType, zero = "string", `""`
		if nullable {
			goType, zero = "sql.NullString", "sql.NullString{}"
		}
	case "smallint":
		goType, zero = "int16", "0"
		if nullable {
			goType, zero = "sql.NullInt64", "sql.NullInt64{}"
		}
	case "integer":
		goType, zero = gotpl.Int32(ctx), "0"
		if nullable {
			goType, zero = "sql.NullInt64", "sql.NullInt64{}"
		}
	case "bigint":
		goType, zero = "int64", "0"
		if nullable {
			goType, zero = "sql.NullInt64", "sql.NullInt64{}"
		}
	case "real":
		goType, zero = "float32", "0.0"
		if nullable {
			goType, zero = "sql.NullFloat64", "sql.NullFloat64{}"
		}
	case "double precision", "numeric":
		goType, zero = "float64", "0.0"
		if nullable {
			goType, zero = "sql.NullFloat64", "sql.NullFloat64{}"
		}
	case "date", "timestamp with time zone", "time with time zone", "time without time zone", "timestamp without time zone":
		goType, zero = "time.Time", "time.Time{}"
		if nullable {
			goType, zero = "sql.NullTime", "sql.NullTime{}"
		}
	case "bit":
		goType, zero = "uint8", "0"
		if nullable {
			goType, zero = "*uint8", "nil"
		}
	case "any", "bit varying", "bytea", "interval", "json", "jsonb", "xml":
		// TODO: write custom type for interval marshaling
		// TODO: marshalling for json types
		goType, zero, asSlice = "byte", "nil", true
	case "hstore":
		goType, zero = "hstore.Hstore", "nil"
	case "uuid":
		goType, zero = "uuid.UUID", "uuid.UUID{}"
		if nullable {
			goType, zero = "*uuid.UUID", "nil"
		}
	default:
		goType, zero = schemaGoType(ctx, typ)
	}
	// handle slices
	switch {
	case asSlice && goType == "string":
		return "StringSlice", "StringSlice{}", prec, nil
	case asSlice:
		return "[]" + goType, "nil", 0, nil
	}
	return goType, zero, prec, nil
}

// PostgresTableColumns returns the columns for a table.
func PostgresTableColumns(ctx context.Context, db models.DB, schema string, table string) ([]*models.Column, error) {
	return models.PostgresTableColumns(ctx, db, schema, table, EnableOids(ctx))
}

// PostgresIndexColumns returns the column list for an index.
//
// FIXME: rewrite this using SQL exclusively using OVER
func PostgresIndexColumns(ctx context.Context, db models.DB, schema string, table string, index string) ([]*models.IndexColumn, error) {
	// load columns
	cols, err := models.PostgresIndexColumns(ctx, db, schema, index)
	if err != nil {
		return nil, err
	}
	// load col order
	colOrd, err := models.PostgresGetColOrder(ctx, db, schema, index)
	if err != nil {
		return nil, err
	}
	// build schema name used in errors
	s := schema
	if s != "" {
		s += "."
	}
	// put cols in order using colOrder
	var ret []*models.IndexColumn
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
				found, c = true, ic
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

// PostgresViewStrip strips '::type AS name' in queries.
func PostgresViewStrip(query []string) ([]string, []string) {
	comments := make([]string, len(query))
	for i, line := range query {
		if pos := stripRE.FindStringIndex(line); pos != nil {
			query[i] = line[:pos[0]] + line[pos[1]:]
			comments[i] = line[pos[0]:pos[1]]
		}
	}
	return query, comments
}

// stripRE is the regexp to match the '::type AS name' portion in a query,
// which is a quirk/requirement of generating queries for postgres.
var stripRE = regexp.MustCompile(`(?i)::[a-z][a-z0-9_\.]+\s+AS\s+[a-z][a-z0-9_\.]+`)

// OidsKey is the oids context key.
const OidsKey ContextKey = "oids"

// EnableOids returns the EnableOids value from the context.
func EnableOids(ctx context.Context) bool {
	b, _ := ctx.Value(OidsKey).(bool)
	return b
}
