// Package loader contains database schema, type and query loaders.
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
var loaders = make(map[string]*Loader)

// Register registers a database loader.
func Register(loader *Loader) {
	loaders[loader.Driver] = loader
}

// Get retrieves a database loader for the provided driver name.
func Get(driver string) *Loader {
	return loaders[driver]
}

// Flags returns the additional driver flags for the loaders.
//
// These should be added to the invocation context for any call to a loader
// func.
func Flags() []xo.FlagSet {
	var drivers []string
	for driver := range loaders {
		drivers = append(drivers, driver)
	}
	sort.Strings(drivers)
	var flags []xo.FlagSet
	for _, driver := range drivers {
		l := loaders[driver]
		if l.Flags == nil {
			continue
		}
		for _, flag := range l.Flags() {
			flags = append(flags, xo.FlagSet{
				Type: driver,
				Name: string(flag.ContextKey),
				Flag: flag,
			})
		}
	}
	return flags
}

// Loader loads type information from a database.
type Loader struct {
	Driver           string
	Mask             string
	Flags            func() []xo.Flag
	GoType           func(context.Context, xo.Datatype) (string, string, error)
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
	ViewStrip        func([]string) ([]string, []string)
	ViewCreate       func(context.Context, models.DB, string, string, []string) (sql.Result, error)
	ViewSchema       func(context.Context, models.DB, string) (string, error)
	ViewTruncate     func(context.Context, models.DB, string, string) (sql.Result, error)
	ViewDrop         func(context.Context, models.DB, string, string) (sql.Result, error)
}

// NthParam returns the 0-based Nth param for the Loader.
func (l *Loader) NthParam(i int) string {
	mask := l.Mask
	if mask == "" {
		return "?"
	}
	if strings.Contains(mask, "%d") {
		return fmt.Sprintf(mask, i+1)
	}
	return mask
}

// SchemaName loads the active schema name for a database.
func (l *Loader) SchemaName(ctx context.Context, db models.DB) (string, error) {
	if l.Schema != nil {
		return l.Schema(ctx, db)
	}
	return "", nil
}

// Int32 returns int32 from the context.
func Int32(ctx context.Context) string {
	s, _ := ctx.Value(xo.Int32Key).(string)
	return s
}

// Uint32 returns uint32 from the context.
func Uint32(ctx context.Context) string {
	s, _ := ctx.Value(xo.Uint32Key).(string)
	return s
}

// SchemaGoType returns Go type and zero for a type, removing a "<schema>."
// prefix when the type is determined to be in the same package.
func SchemaGoType(ctx context.Context, typ string, nullable bool) (string, string) {
	if _, schema, _ := xo.DriverSchemaNthParam(ctx); strings.HasPrefix(typ, schema+".") {
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
