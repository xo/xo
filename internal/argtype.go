package internal

import (
	"io/ioutil"
	"path"
	"text/template"

	"github.com/knq/xo/templates"
)

// ArgType is the type that specifies the command line arguments.
type ArgType struct {
	// Verbose enables verbose output.
	//Verbose bool `arg:"-v,help:toggle verbose"`

	// DSN is the database string (ie, pgsql://user@blah:localhost:5432/dbname?args=)
	DSN string `arg:"positional,required,help:data source name"`

	// Schema is the name of the schema to query.
	Schema string `arg:"-s,help:schema name to generate Go types for"`

	// Out is the output path. If Out is a file, then that will be used as the
	// path. If Out is a directory, then the output file will be
	// Out/<$CWD>.xo.go
	Out string `arg:"-o,help:output path or file name"`

	// Append toggles to append to the existing types.
	Append bool `arg:"-a,help:append to existing files"`

	// Suffix is the output suffix for filenames.
	Suffix string `arg:"-f,help:output file suffix"`

	// SingleFile when toggled changes behavior so that output is to one f ile.
	SingleFile bool `arg:"--single-file,help:toggle single file output"`

	// Package is the name used to generate package headers. If not specified,
	// the name of the output directory will be used instead.
	Package string `arg:"-p,help:package name used in generated Go code"`

	// CustomTypePackage is the Go package name to use for unknown types.
	CustomTypePackage string `arg:"--custom-type-package,-C,help:Go package name to use for custom or unknown types"`

	// Int32Type is the type to assign those discovered as int32 (ie, serial, integer, etc).
	Int32Type string `arg:"--int32-type,-i,help:Go type to assign to integers"`

	// Uint32Type is the type to assign those discovered as uint32.
	Uint32Type string `arg:"--uint32-type,-u,help:Go type to assign to unsigned integers"`

	// IncTypes are the types to include.
	InclTypes []string `arg:"--include,help:include type(s)"`

	// ExcTypes are the types to exclude.
	ExclTypes []string `arg:"--exclude,help:exclude type(s)"`

	// QueryMode toggles whether or not to parse a query from stdin.
	QueryMode bool `arg:"--query-mode,-N,help:enable query mode"`

	// Query is the passed query. If not specified, then os.Stdin will be used.
	// cli args take precedence over stdin.
	Query string `arg:"-Q,help:query to generate Go type and func from"`

	// QueryType is the name to give to the Go type generated from the query.
	QueryType string `arg:"--query-type,-T,help:query's generated Go type"`

	// QueryFunc is the name to assign to the generated query func.
	QueryFunc string `arg:"--query-func,-F,help:query's generated Go func name"`

	// QueryOnlyOne toggles the generated query code to expect only one result.
	QueryOnlyOne bool `arg:"--query-only-one,-1,help:toggle query's generated Go func to return only one result"`

	// QueryTrim enables triming whitespace on the supplied query.
	QueryTrim bool `arg:"--query-trim,-M,help:toggle trimming of query whitespace in generated Go code"`

	// QueryStrip enables stripping the '::<type> AS <name>' from queries.
	QueryStrip bool `arg:"--query-strip,-B,help:toggle stripping '::type AS name' from query in generated Go code"`

	// TypeComment is the type comment for a query.
	QueryTypeComment string `arg:"--query-type-comment,help:comment for query's generated Go type"`

	// FuncComment is the func comment to provide the named query.
	QueryFuncComment string `arg:"--query-func-comment,help:comment for query's generated Go func"`

	// QueryParamDelimiter is the delimiter for parameterized values for a query.
	QueryParamDelimiter string `arg:"--query-delimiter,-D,help:delimiter for query's embedded Go parameters"`

	// TemplatePath is the path to use the user supplied templates instead of
	// the built in versions.
	TemplatePath string `arg:"--template-path,help:user supplied template path"`

	// NoExtra when toggled will not generate certain extras.
	//NoExtra bool `arg:"--no-extra,-Z,help:"disable extra code generation"`

	// Path is the output path, as derived from Out.
	Path string `arg:"-"`

	// Filename is the output filename, as derived from Out.
	Filename string `arg:"-"`

	// Loader is the schema loader
	Loader Loader `arg:"-"`

	// templateSet is the set of templates to use for generating data.
	templateSet *TemplateSet `arg:"-"`

	// Generated is the generated templates after a run.
	Generated []TBuf `arg:"-"`

	// KnownTypeMap is the collection of known Go types.
	KnownTypeMap map[string]bool `arg:"-"`

	// ShortNameTypeMap is the collection of Go style short names for types, mainly
	// used for use with declaring a func receiver on a type.
	ShortNameTypeMap map[string]string `arg:"-"`
}

// UserTemplateLoader loads templates from the specified name
func (a *ArgType) TemplateLoader(name string) ([]byte, error) {
	// no template path specified
	if a.TemplatePath == "" {
		return templates.Asset(name)
	}

	return ioutil.ReadFile(path.Join(a.TemplatePath, name))
}

// TemplateSet retrieves the created template set.
func (a *ArgType) TemplateSet() *TemplateSet {
	if a.templateSet == nil {
		a.templateSet = &TemplateSet{
			funcs: a.NewTemplateFuncs(),
			l:     a.TemplateLoader,
			tpls:  map[string]*template.Template{},
		}
	}

	return a.templateSet
}
