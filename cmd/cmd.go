// Package cmd provides xo command-line application logic.
package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/kenshaw/snaker"
	"github.com/spf13/cobra"
	"github.com/xo/dburl"
	"github.com/xo/dburl/passfile"
	"github.com/xo/xo/loader"
	"github.com/xo/xo/models"
	"github.com/xo/xo/templates"
	xo "github.com/xo/xo/types"
	"github.com/yookoala/realpath"
)

// Args contains command-line arguments.
type Args struct {
	// Verbose enables verbose output.
	Verbose bool
	// LoaderParams are database loader parameters.
	LoaderParams LoaderParams
	// TemplateParams are template parameters.
	TemplateParams TemplateParams
	// QueryParams are query parameters.
	QueryParams QueryParams
	// SchemaParams are schema parameters.
	SchemaParams SchemaParams
	// OutParams are out parameters.
	OutParams OutParams
}

// NewArgs creates args for the provided template names.
func NewArgs(name string, names ...string) *Args {
	// default args
	return &Args{
		LoaderParams: LoaderParams{
			Flags: make(map[xo.ContextKey]*xo.Value),
		},
		TemplateParams: TemplateParams{
			Type:  xo.NewValue("string", name, "template type", names...),
			Flags: make(map[xo.ContextKey]*xo.Value),
		},
		SchemaParams: SchemaParams{
			FkMode:  xo.NewValue("string", "smart", "foreign key resolution mode", "smart", "parent", "field", "key"),
			Include: xo.NewValue("glob", "", "include types"),
			Exclude: xo.NewValue("glob", "", "exclude types"),
		},
	}
}

// LoaderParams are loader parameters.
type LoaderParams struct {
	// Schema is the name of the database schema.
	Schema string
	// Flags are additional loader flags.
	Flags map[xo.ContextKey]*xo.Value
}

// TemplateParams are template parameters.
type TemplateParams struct {
	// Type is the name of the template.
	Type *xo.Value
	// Src is the src directory of the template.
	Src string
	// Flags are additional template flags.
	Flags map[xo.ContextKey]*xo.Value
}

// QueryParams are query parameters.
type QueryParams struct {
	// Query is the query to introspect.
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
	// AllowNulls enables results to have null types.
	AllowNulls bool
}

// SchemaParams are schema parameters.
type SchemaParams struct {
	// FkMode is the foreign resolution mode.
	FkMode *xo.Value
	// Include allows the user to specify which types should be included. Can
	// match multiple types via regex patterns.
	//
	// - When unspecified, all types are included.
	// - When specified, only types match will be included.
	// - When a type matches an exclude entry and an include entry,
	//   the exclude entry will take precedence.
	Include *xo.Value
	// Exclude allows the user to specify which types should be skipped. Can
	// match multiple types via regex patterns.
	//
	// When unspecified, all types are included in the schema.
	Exclude *xo.Value
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
	// Single when true changes behavior so that output is to one file.
	Single string
	// Debug toggles direct writing of files to disk, skipping post processing.
	Debug bool
}

// Run runs the code generation.
func Run(ctx context.Context, name, version string, cmdArgs ...string) error {
	dir := parseArg("--src", "-d", cmdArgs)
	template := parseArg("--template", "-t", cmdArgs)
	ts, err := NewTemplateSet(ctx, dir, template)
	if err != nil {
		return err
	}
	// create args
	args := NewArgs(ts.Target(), ts.Targets()...)
	// create root command
	cmd, err := RootCommand(ctx, name, version, ts, args, cmdArgs...)
	if err != nil {
		return err
	}
	// execute
	return cmd.Execute()
}

// NewTemplateSet creates a new templates set.
func NewTemplateSet(ctx context.Context, dir, template string) (*templates.Set, error) {
	// build template ts
	ts := templates.NewDefaultTemplateSet(ctx)
	switch {
	case dir == "" && template == "":
		// show all default templates
		if err := ts.LoadDefaults(ctx); err != nil {
			return nil, err
		}
	case template != "":
		// only load the selected default template
		if err := ts.LoadDefault(ctx, template); err != nil {
			return nil, err
		}
		ts.Use(template)
	default:
		// load specified template
		s := snaker.SnakeToCamel(filepath.Base(dir))
		s = strings.ReplaceAll(strings.ToLower(s), "_", "-")
		// add template
		var err error
		if s, err = ts.Add(ctx, s, os.DirFS(dir), false); err != nil {
			return nil, err
		}
		// use
		ts.Use(s)
	}
	return ts, nil
}

// RootCommand creates the root command.
func RootCommand(ctx context.Context, name, version string, ts *templates.Set, args *Args, cmdargs ...string) (*cobra.Command, error) {
	// command
	cmd := &cobra.Command{
		Use:     name,
		Version: version,
		Short:   name + ", the templated code generator for databases.",
	}
	// general config
	_ = cmd.Flags().StringP("config", "c", "", "config file")
	cmd.PersistentFlags().BoolVarP(&args.Verbose, "verbose", "v", false, "enable verbose output")
	cmd.SetVersionTemplate("{{ .Name }} {{ .Version }}\n")
	cmd.InitDefaultHelpCmd()
	cmd.SetArgs(cmdargs)
	cmd.SilenceErrors, cmd.SilenceUsage = true, true
	// add sub commands
	var subCmds []*cobra.Command
	for _, f := range []func(context.Context, *templates.Set, *Args) (*cobra.Command, error){
		QueryCommand,
		SchemaCommand,
		DumpCommand,
	} {
		c, err := f(ctx, ts, args)
		if err != nil {
			return nil, err
		}
		subCmds = append(subCmds, c)
	}
	cmd.AddCommand(subCmds...)
	return cmd, nil
}

// QueryCommand builds the query command.
func QueryCommand(ctx context.Context, ts *templates.Set, args *Args) (*cobra.Command, error) {
	// query command
	cmd := &cobra.Command{
		Use:   "query <database url>",
		Short: "Generate code for a database query from a template.",
		RunE:  Exec(ctx, "query", ts, args),
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	databaseFlags(cmd, args)
	outFlags(cmd, args)
	flags.StringVarP(&args.QueryParams.Query, "query", "Q", "", "custom database query (uses stdin if not provided)")
	flags.StringVarP(&args.QueryParams.Type, "type", "T", "", "type name")
	flags.StringVar(&args.QueryParams.TypeComment, "type-comment", "", "type comment")
	flags.StringVarP(&args.QueryParams.Func, "func", "F", "", "func name")
	flags.StringVar(&args.QueryParams.FuncComment, "func-comment", "", "func comment")
	flags.BoolVarP(&args.QueryParams.Trim, "trim", "M", false, "enable trimming whitespace")
	flags.BoolVarP(&args.QueryParams.Strip, "strip", "B", false, "enable stripping type casts")
	flags.BoolVarP(&args.QueryParams.One, "one", "1", false, "enable returning single (only one) result")
	flags.BoolVarP(&args.QueryParams.Flat, "flat", "l", false, "enable returning unstructured (flat) values")
	flags.BoolVarP(&args.QueryParams.Exec, "exec", "X", false, "enable exec (disables query introspection)")
	flags.BoolVarP(&args.QueryParams.Interpolate, "interpolate", "I", false, "enable interpolation of embedded params")
	flags.StringVarP(&args.QueryParams.Delimiter, "delimiter", "L", "%%", "delimiter used for embedded params")
	flags.StringVarP(&args.QueryParams.Fields, "fields", "Z", "", "override field names for results")
	flags.BoolVarP(&args.QueryParams.AllowNulls, "allow-nulls", "U", false, "allow result fields with NULL values")
	if err := templateFlags(cmd, ts, true, args); err != nil {
		return nil, err
	}
	return cmd, nil
}

// SchemaCommand builds the schema command.
func SchemaCommand(ctx context.Context, ts *templates.Set, args *Args) (*cobra.Command, error) {
	// schema command
	cmd := &cobra.Command{
		Use:   "schema <database url>",
		Short: "Generate code for a database schema from a template.",
		RunE:  Exec(ctx, "schema", ts, args),
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	databaseFlags(cmd, args)
	outFlags(cmd, args)
	flags.VarP(args.SchemaParams.FkMode, "fk-mode", "k", args.SchemaParams.FkMode.Desc())
	flags.VarP(args.SchemaParams.Include, "include", "i", args.SchemaParams.Include.Desc())
	flags.VarP(args.SchemaParams.Exclude, "exclude", "e", args.SchemaParams.Exclude.Desc())
	flags.BoolVarP(&args.SchemaParams.UseIndexNames, "use-index-names", "j", false, "use index names as defined in schema for generated code")
	if err := templateFlags(cmd, ts, true, args); err != nil {
		return nil, err
	}
	if err := loaderFlags(cmd, args); err != nil {
		return nil, err
	}
	return cmd, nil
}

// DumpCommand builds the dump command.
func DumpCommand(ctx context.Context, ts *templates.Set, args *Args) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "dump <dir>",
		Short: "Dump template to path.",
		RunE: func(cmd *cobra.Command, v []string) error {
			// set template
			ts.Use(args.TemplateParams.Type.AsString())
			// get template src
			src, err := ts.Src()
			if err != nil {
				return err
			}
			// ensure out dir exists
			if err := checkDir(v[0]); err != nil {
				return err
			}
			// dump
			return fs.WalkDir(src, ".", func(n string, d fs.DirEntry, err error) error {
				switch {
				case err != nil:
					return err
				case d.IsDir():
					return os.MkdirAll(filepath.Join(v[0], n), 0o755)
				}
				buf, err := fs.ReadFile(src, n)
				if err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(v[0], n), buf, 0o644)
			})
		},
	}
	if err := templateFlags(cmd, ts, false, args); err != nil {
		return nil, err
	}
	cmd.Args = cobra.ExactArgs(1)
	cmd.SetUsageTemplate(cmd.UsageTemplate() + "\nArgs:\n  <dir>  out directory\n\n")
	return cmd, nil
}

// databaseFlags adds database flags to the command.
func databaseFlags(cmd *cobra.Command, args *Args) {
	cmd.Flags().StringVarP(&args.LoaderParams.Schema, "schema", "s", "", "database schema name")
	cmd.Args = cobra.ExactArgs(1)
	cmd.SetUsageTemplate(cmd.UsageTemplate() + "\nArgs:\n  <database url>  database url (e.g., postgres://user:pass@localhost:port/dbname, mysql://... )\n\n")
}

// outFlags adds out flags to the command.
func outFlags(cmd *cobra.Command, args *Args) {
	cmd.Flags().StringVarP(&args.OutParams.Out, "out", "o", "models", "out path")
	cmd.Flags().BoolVarP(&args.OutParams.Debug, "debug", "D", false, "debug generated code (writes generated code to disk without post processing)")
	cmd.Flags().StringVarP(&args.OutParams.Single, "single", "S", "", "output all contents to the specified file")
}

// loaderFlags adds database loader flags to the command.
func loaderFlags(cmd *cobra.Command, args *Args) error {
	for _, flag := range loader.Flags() {
		if err := flag.Add(cmd, args.LoaderParams.Flags); err != nil {
			return err
		}
	}
	return nil
}

// templateFlags adds template flags to the command.
func templateFlags(cmd *cobra.Command, ts *templates.Set, extra bool, args *Args) error {
	cmd.Flags().VarP(args.TemplateParams.Type, "template", "t", args.TemplateParams.Type.Desc())
	if extra {
		cmd.Flags().StringVarP(&args.TemplateParams.Src, "src", "d", "", "template source directory")
		for _, name := range ts.Targets() {
			for _, flag := range ts.Flags(name) {
				if err := flag.Add(cmd, args.TemplateParams.Flags); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func parseArg(short, full string, args []string) (s string) {
	defer func() {
		s = strings.TrimSpace(s)
	}()
	for i := range args {
		switch s := strings.TrimSpace(args[i]); {
		case s == short, s == full:
			if i < len(args)-1 {
				return args[i+1]
			}
		case strings.HasPrefix(s, short):
			return strings.TrimPrefix(s, short)
		case strings.HasPrefix(s, full):
			return strings.TrimPrefix(s, full)
		}
	}
	return ""
}

// Exec handles the execution for query and schema.
func Exec(ctx context.Context, mode string, ts *templates.Set, args *Args) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, cmdargs []string) error {
		// setup args
		if err := checkArgs(cmd, mode, ts, args); err != nil {
			return err
		}
		// set template
		ts.Use(args.TemplateParams.Type.AsString())
		// build context
		ctx := BuildContext(ctx, args)
		// enable verbose output for sql queries
		if args.Verbose {
			models.SetLogger(func(str string, v ...interface{}) {
				s, z := "SQL: %s\n", []interface{}{str}
				if len(v) != 0 {
					s, z = s+"PARAMS: %v\n", append(z, v)
				}
				fmt.Printf(s+"\n", z...)
			})
		}
		// open database
		var err error
		if ctx, err = open(ctx, cmdargs[0], args.LoaderParams.Schema); err != nil {
			return err
		}
		// load
		set, err := load(ctx, mode, ts, args)
		if err != nil {
			return err
		}
		return Generate(ctx, mode, ts, set, args)
	}
}

// Generate generates the XO files with the provided templates, data, and arguments.
func Generate(ctx context.Context, mode string, ts *templates.Set, set *xo.Set, args *Args) error {
	// create set context
	ctx = ts.NewContext(ctx, mode)
	if err := displayErrors(ts); err != nil {
		return err
	}
	// preprocess
	ts.Pre(ctx, args.OutParams.Out, mode, set)
	if err := displayErrors(ts); err != nil {
		return err
	}
	// process
	ts.Process(ctx, args.OutParams.Out, mode, set)
	if err := displayErrors(ts); err != nil {
		return err
	}
	// post
	if !args.OutParams.Debug {
		ts.Post(ctx, mode)
		if err := displayErrors(ts); err != nil {
			return err
		}
	}
	// dump
	ts.Dump(args.OutParams.Out)
	if err := displayErrors(ts); err != nil {
		return err
	}
	return nil
}

// checkArgs sets up and checks args.
func checkArgs(cmd *cobra.Command, mode string, ts *templates.Set, args *Args) error {
	// check template is available for the mode
	if err := ts.For(mode); err != nil {
		return err
	}
	// check --src and --template are exclusive
	if cmd.Flags().Lookup("src").Changed && cmd.Flags().Lookup("template").Changed {
		return errors.New("--src and --template cannot be used together")
	}
	// read query string from stdin if not provided via --query
	if mode == "query" && args.QueryParams.Query == "" {
		buf, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		args.QueryParams.Query = string(bytes.TrimRight(buf, "\r\n"))
	}
	// check out path
	if args.OutParams.Out != "" {
		var err error
		if args.OutParams.Out, err = realpath.Realpath(args.OutParams.Out); err != nil {
			return err
		}
		if err := checkDir(args.OutParams.Out); err != nil {
			return err
		}
	}
	return nil
}

// BuildContext builds a context for the mode and template.
func BuildContext(ctx context.Context, args *Args) context.Context {
	// add loader flags
	for k, v := range args.LoaderParams.Flags {
		ctx = context.WithValue(ctx, k, v.Interface())
	}
	// add template flags
	for k, v := range args.TemplateParams.Flags {
		ctx = context.WithValue(ctx, k, v.Interface())
	}
	// add out
	ctx = context.WithValue(ctx, xo.OutKey, args.OutParams.Out)
	ctx = context.WithValue(ctx, xo.SingleKey, args.OutParams.Single)
	return ctx
}

// open opens a connection to the database, returning a context for use in
// template generation.
func open(ctx context.Context, urlstr, schema string) (context.Context, error) {
	v, err := user.Current()
	if err != nil {
		return nil, err
	}
	// parse dsn
	u, err := dburl.Parse(urlstr)
	if err != nil {
		return nil, err
	}
	// open database
	db, err := passfile.OpenURL(u, v.HomeDir, "xopass")
	if err != nil {
		return nil, err
	}
	// add driver to context
	ctx = context.WithValue(ctx, xo.DriverKey, u.Driver)
	// add db to context
	ctx = context.WithValue(ctx, xo.DbKey, db)
	// determine schema
	if schema == "" {
		if schema, err = loader.Schema(ctx); err != nil {
			return nil, err
		}
	}
	// add schema to context
	ctx = context.WithValue(ctx, xo.SchemaKey, schema)
	return ctx, nil
}

// load loads a set of queries or schemas.
func load(ctx context.Context, mode string, ts *templates.Set, args *Args) (*xo.Set, error) {
	f := LoadSchema
	if mode == "query" {
		f = LoadQuery
	}
	set := new(xo.Set)
	if err := f(ctx, set, args); err != nil {
		return nil, err
	}
	return set, nil
}

// displayErrors displays collected errors from the set.
func displayErrors(ts *templates.Set) error {
	if errors := ts.Errors(); len(errors) != 0 {
		for _, err := range errors {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
		return fmt.Errorf("%d errors encountered", len(errors))
	}
	return nil
}

// checkDir checks that dir exists.
func checkDir(dir string) error {
	if !isDir(dir) {
		return fmt.Errorf("%s must exist and must be a directory", dir)
	}
	return nil
}

// isDir determines if dir is a directory.
func isDir(dir string) bool {
	if fi, err := os.Stat(dir); err == nil {
		return fi.IsDir()
	}
	return false
}
