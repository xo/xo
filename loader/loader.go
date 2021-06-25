// Package loader contains database schema, type and query loaders.
package loader

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/kenshaw/snaker"
	"github.com/xo/xo/models"
	"github.com/xo/xo/templates"
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
func Flags() []FlagSet {
	var drivers []string
	for driver := range loaders {
		drivers = append(drivers, driver)
	}
	sort.Strings(drivers)
	var flags []FlagSet
	for _, driver := range drivers {
		l := loaders[driver]
		if l.Flags == nil {
			continue
		}
		for _, flag := range l.Flags() {
			flags = append(flags, FlagSet{
				Driver: driver,
				Name:   string(flag.ContextKey),
				Flag:   flag,
			})
		}
	}
	return flags
}

// FlagSet is a set of flags for a driver.
type FlagSet struct {
	Driver string
	Name   string
	Flag   Flag
}

// Flag is a option flag.
type Flag struct {
	ContextKey  xo.ContextKey
	Desc        string
	PlaceHolder string
	Default     string
	Short       rune
	Value       interface{}
	Enums       []string
}

// Loader loads type information from a database.
type Loader struct {
	Driver           string
	Mask             string
	Flags            func() []Flag
	GoType           func(context.Context, xo.Datatype) (string, string, error)
	Schema           func(context.Context, models.DB) (string, error)
	Enums            func(context.Context, models.DB, string) ([]*models.Enum, error)
	EnumValues       func(context.Context, models.DB, string, string) ([]*models.EnumValue, error)
	Procs            func(context.Context, models.DB, string) ([]*models.Proc, error)
	ProcParams       func(context.Context, models.DB, string, string) ([]*models.ProcParam, error)
	Tables           func(context.Context, models.DB, string, string) ([]*models.Table, error)
	TableColumns     func(context.Context, models.DB, string, string) ([]*models.Column, error)
	TableSequences   func(context.Context, models.DB, string) ([]*models.Sequence, error)
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

// CtxLoader returns loader from the context.
func CtxLoader(ctx context.Context) *Loader {
	l, _ := ctx.Value(xo.LoaderKey).(*Loader)
	return l
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

// intRE matches Go int types.
var intRE = regexp.MustCompile(`^int(32|64)?$`)

// parsePrec parses "type[ (precision[,scale])]" strings returning the parsed
// precision and scale.
func parsePrec(typ string) (string, int, int, error) {
	typ, prec, scale := strings.ToLower(typ), -1, -1
	if m := precRE.FindStringIndex(typ); m != nil {
		s := typ[m[0]+1 : m[1]-1]
		if i := strings.LastIndex(s, ","); i != -1 {
			var err error
			if scale, err = strconv.Atoi(strings.TrimSpace(s[i+1:])); err != nil {
				return "", 0, 0, fmt.Errorf("could not parse scale: %w", err)
			}
			s = s[:i]
		}
		// extract precision
		var err error
		if prec, err = strconv.Atoi(strings.TrimSpace(s)); err != nil {
			return "", 0, 0, fmt.Errorf("could not parse precision: %w", err)
		}
		typ = typ[:m[0]]
	}
	return strings.TrimSpace(typ), prec, scale, nil
}

// precRE is the regexp that matches "(precision[,scale])" definitions in a
// database.
var precRE = regexp.MustCompile(`\(([0-9]+)(\s*,\s*[0-9]+\s*)?\)$`)

// schemaGoType returns Go type and zero for a type, removing a "<schema>."
// prefix when the type is determined to be in the same package.
func schemaGoType(ctx context.Context, typ string) (string, string) {
	if schema := templates.Schema(ctx); strings.HasPrefix(typ, schema+".") {
		// in the same schema, so chop off
		typ = typ[len(schema)+1:]
	}
	s := snaker.SnakeToCamelIdentifier(typ)
	return s, s + "{}"
}
