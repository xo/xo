// Package loader loads query and schema information from mysql, oracle,
// postgres, sqlite3 and sqlserver databases.
package loader

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/kenshaw/snaker"
	"github.com/xo/xo/models"
	xo "github.com/xo/xo/types"
)

// loaders are registered database loaders.
var loaders = make(map[string]Loader)

// Register registers a database loader.
func Register(typ string, loader Loader) {
	loaders[typ] = loader
}

// Flags returns the additional flags for the loaders.
//
// These should be added to the invocation context for any call to a loader
// func.
func Flags() []xo.FlagSet {
	var types []string
	for typ := range loaders {
		types = append(types, typ)
	}
	sort.Strings(types)
	var flags []xo.FlagSet
	for _, typ := range types {
		l := loaders[typ]
		if l.Flags == nil {
			continue
		}
		for _, flag := range l.Flags() {
			flags = append(flags, xo.FlagSet{
				Type: typ,
				Name: string(flag.ContextKey),
				Flag: flag,
			})
		}
	}
	return flags
}

// Loader loads type information from a database.
type Loader struct {
	Type             string
	Mask             string
	Flags            func() []xo.Flag
	Schema           func(context.Context, models.DB) (string, error)
	Enums            func(context.Context, models.DB, string) ([]*models.Enum, error)
	EnumValues       func(context.Context, models.DB, string, string) ([]*models.EnumValue, error)
	Procs            func(context.Context, models.DB, string) ([]*models.Proc, error)
	ProcParams       func(context.Context, models.DB, string, string) ([]*models.ProcParam, error)
	Tables           func(context.Context, models.DB, string, string) ([]*models.Table, error)
	TableColumns     func(context.Context, models.DB, string, string) ([]*models.Column, error)
	TableSequences   func(context.Context, models.DB, string, string) ([]*models.Sequence, error)
	TableForeignKeys func(context.Context, models.DB, string, string) ([]*models.ForeignKey, error)
	TableIndexes     func(context.Context, models.DB, string, string) ([]*models.Index, error)
	IndexColumns     func(context.Context, models.DB, string, string, string) ([]*models.IndexColumn, error)
	ViewCreate       func(context.Context, models.DB, string, string, []string) (sql.Result, error)
	ViewSchema       func(context.Context, models.DB, string) (string, error)
	ViewTruncate     func(context.Context, models.DB, string, string) (sql.Result, error)
	ViewDrop         func(context.Context, models.DB, string, string) (sql.Result, error)
	ViewStrip        func([]string, []string) ([]string, []string, []string, error)
}

// get retrieves the database connection, loader, and schema name from the
// context.
func get(ctx context.Context) (*sql.DB, *Loader, string, error) {
	typ, _ := ctx.Value(xo.DriverKey).(string)
	l, ok := loaders[typ]
	if !ok {
		return nil, nil, "", fmt.Errorf("no database loader available for %q", typ)
	}
	db, _ := ctx.Value(xo.DbKey).(*sql.DB)
	schema, _ := ctx.Value(xo.SchemaKey).(string)
	return db, &l, schema, nil
}

// NthParam returns a 0-based func to generate the nth param placeholder for
// database queries.
func NthParam(ctx context.Context) (func(int) string, error) {
	_, l, _, err := get(ctx)
	if err != nil {
		return nil, err
	}
	mask := "?"
	if l.Mask != "" {
		mask = l.Mask
	}
	if !strings.Contains(mask, "%d") {
		return func(int) string {
			return mask
		}, nil
	}
	return func(i int) string {
		return fmt.Sprintf(mask, i+1)
	}, nil
}

// Schema loads the active schema name from the context.
func Schema(ctx context.Context) (string, error) {
	db, l, _, err := get(ctx)
	if err != nil {
		return "", err
	}
	return l.Schema(ctx, db)
}

// Enums returns the database enums.
func Enums(ctx context.Context) ([]*models.Enum, error) {
	db, l, schema, err := get(ctx)
	if err != nil {
		return nil, err
	}
	if l.Enums != nil {
		return l.Enums(ctx, db, schema)
	}
	return nil, nil
}

// EnumValues returns the database enum values.
func EnumValues(ctx context.Context, enum string) ([]*models.EnumValue, error) {
	db, l, schema, err := get(ctx)
	if err != nil {
		return nil, err
	}
	return l.EnumValues(ctx, db, schema, enum)
}

// Procs returns the database procs.
func Procs(ctx context.Context) ([]*models.Proc, error) {
	db, l, schema, err := get(ctx)
	if err != nil {
		return nil, err
	}
	if l.Procs != nil {
		return l.Procs(ctx, db, schema)
	}
	return nil, nil
}

// ProcParams returns the database proc params.
func ProcParams(ctx context.Context, id string) ([]*models.ProcParam, error) {
	db, l, schema, err := get(ctx)
	if err != nil {
		return nil, err
	}
	if l.ProcParams != nil {
		return l.ProcParams(ctx, db, schema, id)
	}
	return nil, nil
}

// Tables returns the database tables.
func Tables(ctx context.Context, typ string) ([]*models.Table, error) {
	db, l, schema, err := get(ctx)
	if err != nil {
		return nil, err
	}
	return l.Tables(ctx, db, schema, typ)
}

// TableColumns returns the database table columns.
func TableColumns(ctx context.Context, table string) ([]*models.Column, error) {
	db, l, schema, err := get(ctx)
	if err != nil {
		return nil, err
	}
	return l.TableColumns(ctx, db, schema, table)
}

// TableSequences returns the database table sequences.
func TableSequences(ctx context.Context, table string) ([]*models.Sequence, error) {
	db, l, schema, err := get(ctx)
	if err != nil {
		return nil, err
	}
	return l.TableSequences(ctx, db, schema, table)
}

// TableForeignKeys returns the database table foreign keys.
func TableForeignKeys(ctx context.Context, table string) ([]*models.ForeignKey, error) {
	db, l, schema, err := get(ctx)
	if err != nil {
		return nil, err
	}
	return l.TableForeignKeys(ctx, db, schema, table)
}

// TableIndexes returns the database table indexes.
func TableIndexes(ctx context.Context, table string) ([]*models.Index, error) {
	db, l, schema, err := get(ctx)
	if err != nil {
		return nil, err
	}
	return l.TableIndexes(ctx, db, schema, table)
}

// IndexColumns returns the database index columns.
func IndexColumns(ctx context.Context, table, index string) ([]*models.IndexColumn, error) {
	db, l, schema, err := get(ctx)
	if err != nil {
		return nil, err
	}
	return l.IndexColumns(ctx, db, schema, table, index)
}

// ViewCreate creates a introspection view of a query.
func ViewCreate(ctx context.Context, id string, query []string) (sql.Result, error) {
	db, l, schema, err := get(ctx)
	if err != nil {
		return nil, err
	}
	return l.ViewCreate(ctx, db, schema, id, query)
}

// ViewSchema returns the schema that the introspection view was created in.
func ViewSchema(ctx context.Context, id string) (string, error) {
	db, l, _, err := get(ctx)
	if err != nil {
		return "", err
	}
	if l.ViewSchema != nil {
		return l.ViewSchema(ctx, db, id)
	}
	return "", nil
}

// ViewTruncate truncates the introspection view.
func ViewTruncate(ctx context.Context, id string) (sql.Result, error) {
	db, l, schema, err := get(ctx)
	if err != nil {
		return nil, err
	}
	if l.ViewTruncate != nil {
		return l.ViewTruncate(ctx, db, schema, id)
	}
	return nil, nil
}

// ViewDrop drops the introspection view.
func ViewDrop(ctx context.Context, id string) (sql.Result, error) {
	db, l, schema, err := get(ctx)
	if err != nil {
		return nil, err
	}
	return l.ViewDrop(ctx, db, schema, id)
}

// ViewStrip post processes the query and inspected query, altering as
// necessary and building a set of comments for the query.
func ViewStrip(ctx context.Context, query, inspect []string) ([]string, []string, []string, error) {
	_, l, _, err := get(ctx)
	if err != nil {
		return nil, nil, nil, err
	}
	if l.ViewStrip != nil {
		return l.ViewStrip(query, inspect)
	}
	return query, inspect, make([]string, len(query)), nil
}

// schemaType returns Go type and zero for a type, removing a "<schema>."
// prefix when the type is determined to be in the same package.
func schemaType(typ string, nullable bool, schema string) (string, string) {
	if strings.HasPrefix(typ, schema+".") {
		// in the same schema, so chop off
		typ = typ[len(schema)+1:]
	}
	if nullable {
		typ = "null_" + typ
	}
	s := snaker.SnakeToCamelIdentifier(typ)
	return s, s + "{}"
}

// intRE matches Go int types.
var intRE = regexp.MustCompile(`^int(8|16|32|64)?$`)
