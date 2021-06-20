// Package cmd contains the primary logic of the xo command-line application.
package cmd

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/xo/dburl"
	"github.com/xo/dburl/passfile"
	"github.com/xo/xo/loader"
	"github.com/xo/xo/models"
	"github.com/xo/xo/templates"
	"github.com/yookoala/realpath"
)

// Run runs the code generation.
func Run(ctx context.Context, name, version string) error {
	// build args
	ctx, args, cmd, err := NewArgs(ctx, name, version)
	if err != nil {
		return err
	}
	// build generator
	var g Generator
	switch cmd {
	case "schema":
		var err error
		if ctx, err = Open(ctx, args.DbParams.DSN, args.DbParams.Schema); err != nil {
			return err
		}
		g = NewSchemaGenerator()
	case "query":
		var err error
		if ctx, err = Open(ctx, args.DbParams.DSN, args.DbParams.Schema); err != nil {
			return err
		}
		g = NewQueryGenerator()
	case "dump":
		g = NewNoopGenerator()
	}
	// exec
	if err := g.Exec(ctx, args); err != nil {
		return err
	}
	// process
	if err := g.Process(ctx, args); err != nil {
		return err
	}
	// write
	f := templates.Write
	switch {
	case args.OutParams.Debug:
		f = templates.WriteFiles
	case cmd == "dump":
		f = templates.WriteRaw
	}
	if err := f(ctx); err != nil {
		return err
	}
	// do error ouutput
	errors, err := templates.Errors(ctx)
	if err != nil {
		return err
	}
	for _, err := range errors {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}
	if len(errors) != 0 {
		return fmt.Errorf("%d errors encountered", len(errors))
	}
	return nil
}

// Args contains command-line arguments.
type Args struct {
	// Verbose enables verbose output.
	Verbose bool
	// DbParams are database parameters.
	DbParams DbParams
	// TemplateParams are template parameters.
	TemplateParams TemplateParams
	// QueryParams are query parameters.
	QueryParams QueryParams
	// SchemaParams are schema parameters.
	SchemaParams SchemaParams
	// OutParams are out parameters.
	OutParams OutParams
}

// NewArgs creates the command args.
func NewArgs(ctx context.Context, name, version string) (context.Context, *Args, string, error) {
	// kingpin settings
	kingpin.UsageTemplate(kingpin.CompactUsageTemplate)
	// create args
	args := &Args{
		DbParams: DbParams{
			Flags: make(map[loader.ContextKey]interface{}),
		},
		TemplateParams: TemplateParams{
			Flags: make(map[templates.ContextKey]interface{}),
		},
	}
	// general
	kingpin.Flag("verbose", "enable verbose output").Short('v').Default("false").BoolVar(&args.Verbose)
	// database
	dc := func(cmd *kingpin.CmdClause) {
		cmd.Flag("schema", "database schema name").Short('s').PlaceHolder("<name>").StringVar(&args.DbParams.Schema)
		cmd.Arg("DSN", "data source name").Required().StringVar(&args.DbParams.DSN)
	}
	// template
	tc := func(cmd *kingpin.CmdClause) {
		types := templates.Types()
		desc := strings.Join(types, ", ")
		cmd.Flag("template", "template type ("+desc+"; default: go)").Short('t').Default("go").EnumVar(&args.TemplateParams.Type, types...)
		cmd.Flag("suffix", "file extension suffix for generated files (otherwise set by template type)").Short('f').PlaceHolder("<ext>").StringVar(&args.TemplateParams.Suffix)
	}
	// out
	oc := func(cmd *kingpin.CmdClause) {
		cmd.Flag("out", "out path (default: models)").Short('o').Default("models").PlaceHolder("models").StringVar(&args.OutParams.Out)
		cmd.Flag("append", "enable append mode").Short('a').BoolVar(&args.OutParams.Append)
		cmd.Flag("single", "enable single file output").Short('S').PlaceHolder("<file>").StringVar(&args.OutParams.Single)
		cmd.Flag("debug", "debug generated code (writes generated code to disk without post processing)").Short('D').BoolVar(&args.OutParams.Debug)
	}
	// additonal loader flags
	lf := func(cmd *kingpin.CmdClause) {
		for _, flag := range loader.Flags() {
			f := cmd.Flag(flag.Driver+"-"+flag.Name, flag.Flag.Desc).
				PlaceHolder(flag.Flag.PlaceHolder).
				Short(flag.Flag.Short).
				Default(flag.Flag.Default)
			switch flag.Flag.Value.(type) {
			case bool:
				args.DbParams.Flags[flag.Flag.ContextKey] = newBool(f, args.DbParams.Flags[flag.Flag.ContextKey])
			case string:
				args.DbParams.Flags[flag.Flag.ContextKey] = newString(f, args.DbParams.Flags[flag.Flag.ContextKey], flag.Flag.Enums)
			case []string:
				args.DbParams.Flags[flag.Flag.ContextKey] = newStrings(f, args.DbParams.Flags[flag.Flag.ContextKey], flag.Flag.Enums)
			}
		}
	}
	// additional templates flags
	tf := func(cmd *kingpin.CmdClause) {
		cmd.Flag("src", "template source directory").Short('d').PlaceHolder("<path>").StringVar(&args.TemplateParams.Src)
		for _, flag := range templates.Flags() {
			f := cmd.Flag(flag.Type+"-"+flag.Name, flag.Flag.Desc).
				PlaceHolder(flag.Flag.PlaceHolder).
				Short(flag.Flag.Short).
				Default(flag.Flag.Default)
			switch flag.Flag.Value.(type) {
			case bool:
				args.TemplateParams.Flags[flag.Flag.ContextKey] = newBool(f, args.TemplateParams.Flags[flag.Flag.ContextKey])
			case string:
				args.TemplateParams.Flags[flag.Flag.ContextKey] = newString(f, args.TemplateParams.Flags[flag.Flag.ContextKey], flag.Flag.Enums)
			case []string:
				args.TemplateParams.Flags[flag.Flag.ContextKey] = newStrings(f, args.TemplateParams.Flags[flag.Flag.ContextKey], flag.Flag.Enums)
			}
		}
	}
	// query command
	queryCmd := kingpin.Command("query", "Generate code for a database custom query from a template.")
	dc(queryCmd)
	tc(queryCmd)
	oc(queryCmd)
	queryCmd.Flag("query", "custom database query (uses stdin if not provided)").Short('Q').PlaceHolder(`""`).StringVar(&args.QueryParams.Query)
	queryCmd.Flag("type", "type name").Short('T').PlaceHolder("<name>").StringVar(&args.QueryParams.Type)
	queryCmd.Flag("type-comment", "type comment").PlaceHolder(`""`).StringVar(&args.QueryParams.TypeComment)
	queryCmd.Flag("func", "func name").Short('F').PlaceHolder("<name>").StringVar(&args.QueryParams.Func)
	queryCmd.Flag("func-comment", "func comment").PlaceHolder(`""`).StringVar(&args.QueryParams.FuncComment)
	queryCmd.Flag("trim", "enable trimming whitespace").Short('M').BoolVar(&args.QueryParams.Trim)
	queryCmd.Flag("strip", "enable stripping type casts").Short('B').BoolVar(&args.QueryParams.Strip)
	queryCmd.Flag("one", "enable returning single (only one) result").Short('1').BoolVar(&args.QueryParams.One)
	queryCmd.Flag("flat", "enable returning unstructured values").Short('l').BoolVar(&args.QueryParams.Flat)
	queryCmd.Flag("exec", "enable exec (no introspection performed)").Short('X').BoolVar(&args.QueryParams.Exec)
	queryCmd.Flag("interpolate", "enable interpolation of embedded params").Short('I').BoolVar(&args.QueryParams.Interpolate)
	queryCmd.Flag("delimiter", "delimiter used for embedded params (default: %%)").Short('L').PlaceHolder("%%").Default("%%").StringVar(&args.QueryParams.Delimiter)
	queryCmd.Flag("fields", "override field names for results").Short('Z').PlaceHolder("<field>").StringVar(&args.QueryParams.Fields)
	queryCmd.Flag("allow-nulls", "allow result fields with NULL values").Short('U').BoolVar(&args.QueryParams.AllowNulls)
	tf(queryCmd)
	// schema command
	schemaCmd := kingpin.Command("schema", "Generate code for a database schema from a template.")
	dc(schemaCmd)
	tc(schemaCmd)
	oc(schemaCmd)
	schemaCmd.Flag("fk-mode", "foreign key resolution mode (smart, parent, field, key; default: smart)").Short('k').Default("smart").EnumVar(&args.SchemaParams.FkMode, "smart", "parent", "field", "key")
	schemaCmd.Flag("ignore", "fields to exclude from generated code").Short('I').PlaceHolder("<field>").StringsVar(&args.SchemaParams.Ignore)
	schemaCmd.Flag("use-index-names", "use index names as defined in schema for generated code").Short('j').BoolVar(&args.SchemaParams.UseIndexNames)
	tf(schemaCmd)
	lf(schemaCmd)
	// dump command
	dumpCmd := kingpin.Command("dump", "Dump internal templates to path.")
	tc(dumpCmd)
	dumpCmd.Arg("out", "out path").Required().StringVar(&args.OutParams.Out)
	// add --version flag
	kingpin.Flag("version", "display version and exit").PreAction(func(*kingpin.ParseContext) error {
		fmt.Fprintln(os.Stdout, name, version)
		os.Exit(0)
		return nil
	}).Bool()
	cmd := kingpin.Parse()
	// add loader flags
	for key, v := range args.DbParams.Flags {
		// deref the interface (should always be a pointer to a type)
		ctx = context.WithValue(ctx, key, reflect.ValueOf(v).Elem().Interface())
	}
	// add gen type
	ctx = context.WithValue(ctx, templates.GenTypeKey, cmd)
	// add template type
	ctx = context.WithValue(ctx, templates.TemplateTypeKey, args.TemplateParams.Type)
	// add suffix
	ctx = context.WithValue(ctx, templates.SuffixKey, args.TemplateParams.Suffix)
	// add template flags
	for key, v := range args.TemplateParams.Flags {
		// deref the interface (should always be a pointer to a type)
		ctx = context.WithValue(ctx, key, reflect.ValueOf(v).Elem().Interface())
	}
	// add src to context
	if args.TemplateParams.Src != "" {
		d, err := realpath.Realpath(args.TemplateParams.Src)
		if err != nil {
			return nil, nil, "", err
		}
		fi, err := os.Stat(d)
		switch {
		case err != nil:
			return nil, nil, "", err
		case !fi.IsDir():
			return nil, nil, "", fmt.Errorf("src path %s is not a directory", d)
		}
		ctx = context.WithValue(ctx, templates.SrcKey, os.DirFS(d))
	}
	// add out to context
	if args.OutParams.Out != "" {
		out, err := realpath.Realpath(args.OutParams.Out)
		if err != nil {
			return nil, nil, "", err
		}
		fi, err := os.Stat(out)
		switch {
		case err != nil && os.IsNotExist(err):
			return nil, nil, "", fmt.Errorf("%s must exist and must be a directory", out)
		case err != nil:
			return nil, nil, "", err
		case !fi.IsDir():
			return nil, nil, "", fmt.Errorf("out path %s is not a directory", out)
		}
		ctx = context.WithValue(ctx, templates.OutKey, args.OutParams.Out)
	}
	// enable verbose output for sql queries
	if args.Verbose {
		models.SetLogger(func(s string, v ...interface{}) {
			fmt.Printf("SQL:\n%s\nPARAMS:\n%v\n\n", s, v)
		})
	}
	return ctx, args, cmd, nil
}

// DbParams are database parameters.
type DbParams struct {
	// Schema is the name of the database schema.
	Schema string
	// DSN is the database string (ie, postgres://user:pass@host:5432/dbname?args=)
	DSN string
	// Flags are additional loader flags.
	Flags map[loader.ContextKey]interface{}
}

// TemplateParams are template parameters.
type TemplateParams struct {
	// Type is the name of the template.
	Type string
	// Suffix is the file extension suffix.
	Suffix string
	// Src is the src directory of the template.
	Src string
	// Esc indicates which types to escape (none, schema, table, column, all).
	Esc []string
	// Flags are additional template flags.
	Flags map[templates.ContextKey]interface{}
}

// QueryParams are query parameters.
type QueryParams struct {
	// Query is the passed query.
	//
	// If not specified, then os.Stdin will be used.
	Query string
	// Type is the type name.
	Type string
	// TypeComment is the type comment.
	TypeComment string
	// Func is the func name.
	Func string
	// FuncComment is the func comment.
	FuncComment string
	// Trim enables triming whitespace.
	Trim bool
	// Strip enables stripping the '::<type> AS <name>' in queries.
	Strip bool
	// One toggles the generated code to expect only one result.
	One bool
	// Flat toggles the generated code to return all scanned values directly.
	Flat bool
	// Exec toggles the generated code to do a db exec.
	Exec bool
	// Interpolate enables interpolation.
	Interpolate bool
	// Delimiter is the delimiter for parameterized values.
	Delimiter string
	// Fields are the fields to scan the result to.
	Fields string
	// AllowNulls toggles results can contain null types.
	AllowNulls bool
}

// SchemaParams are schema parameters.
type SchemaParams struct {
	// FkMode is the foreign resolution mode.
	FkMode string
	// Ignore allows the user to specify field names which should not be
	// handled by xo in the generated code.
	Ignore []string
	// UseIndexNames toggles using index names.
	//
	// This is not enabled by default, because index names are often generated
	// using database design software which often gives non-descriptive names
	// to indexes (for example, 'authors__b124214__u_idx' instead of the more
	// descriptive 'authors_title_idx').
	UseIndexNames bool
}

// OutParams are out parameters.
type OutParams struct {
	// Out is the out path.
	Out string
	// Append toggles to append to the existing types.
	Append bool
	// Single when true changes behavior so that output is to one file.
	Single string
	// Debug toggles direct writing of files to disk, skipping post processing.
	Debug bool
}

// Open opens a connection to the database, returning a context for use in the
// application logic.
func Open(ctx context.Context, dsn, schema string) (context.Context, error) {
	// parse dsn
	v, err := dburl.Parse(dsn)
	if err != nil {
		return nil, err
	}
	// grab loader
	l := loader.Get(v.Driver)
	if l == nil {
		return nil, fmt.Errorf("no database loader available for %q", v.Driver)
	}
	// open database
	db, err := passfile.OpenURL(v, "xopass")
	if err != nil {
		return nil, err
	}
	// determine schema
	if schema == "" {
		if schema, err = l.SchemaName(ctx, db); err != nil {
			return nil, err
		}
	}
	// add db to context
	ctx = context.WithValue(ctx, DbKey, db)
	// add loader to context
	ctx = context.WithValue(ctx, LoaderKey, l)
	// add driver to context
	ctx = context.WithValue(ctx, templates.DriverKey, v.Driver)
	// add schema to context
	ctx = context.WithValue(ctx, templates.SchemaKey, schema)
	// add nth-func to context
	ctx = context.WithValue(ctx, templates.NthParamKey, l.NthParam)
	return ctx, nil
}

// ContextKey is a context key.
type ContextKey string

// Context key values.
const (
	DbKey     ContextKey = "db"
	LoaderKey ContextKey = "loader"
)

// DbLoaderSchema returns the database, loader, and schema name from the context.
func DbLoaderSchema(ctx context.Context) (models.DB, *loader.Loader, string) {
	db, _ := ctx.Value(DbKey).(models.DB)
	l, _ := ctx.Value(LoaderKey).(*loader.Loader)
	schema, _ := ctx.Value(templates.SchemaKey).(string)
	return db, l, schema
}

// Generator is the shared interface for generators.
type Generator interface {
	Exec(context.Context, *Args) error
	Process(context.Context, *Args) error
}

// NoopGenerator is a no op generator.
type NoopGenerator struct{}

// NewNoopGenerator creates a new noop generator.
func NewNoopGenerator() *NoopGenerator {
	return &NoopGenerator{}
}

// Exec satisfies the Generator interface.
func (*NoopGenerator) Exec(context.Context, *Args) error {
	return nil
}

// Process satisfies the Generator interface.
func (*NoopGenerator) Process(ctx context.Context, args *Args) error {
	return nil
}

// newBool creates a new bool when v is nil, otherwise it converts v and returns.
func newBool(f *kingpin.FlagClause, v interface{}) *bool {
	if v == nil {
		b := false
		f.BoolVar(&b)
		return &b
	}
	b := v.(*bool)
	f.BoolVar(b)
	return b
}

// newString creates a new string when v is nil, otherwise it converts v and returns.
func newString(f *kingpin.FlagClause, v interface{}, enums []string) *string {
	if v == nil {
		s := ""
		if len(enums) != 0 {
			f.EnumVar(&s, enums...)
		} else {
			f.StringVar(&s)
		}
		return &s
	}
	s := v.(*string)
	if len(enums) != 0 {
		f.EnumVar(s, enums...)
	} else {
		f.StringVar(s)
	}
	return s
}

// newStrings creates a new string when v is nil, otherwise it converts v and returns.
func newStrings(f *kingpin.FlagClause, v interface{}, enums []string) *[]string {
	switch {
	case v == nil && enums == nil:
		var s []string
		f.StringsVar(&s)
		return &s
	case v != nil && enums == nil:
		s := v.(*[]string)
		f.StringsVar(s)
		return s
	case v == nil:
		var s []string
		f.EnumsVar(&s, enums...)
		return &s
	}
	s := v.(*[]string)
	f.EnumsVar(s, enums...)
	return s
}
