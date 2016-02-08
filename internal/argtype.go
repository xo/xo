package internal

// ArgType is the type that specifies the command line arguments.
type ArgType struct {
	// Verbose enables verbose output.
	Verbose bool `arg:"-v,help:toggle verbose"`

	// Package is the name used to generate package headers. If not specified,
	// the name of the output directory will be used instead.
	Package string `arg:"-p,help:package name used in generated Go code"`

	// CustomTypePackage is the Go package name to use for unknown types.
	CustomTypePackage string `arg:"--custom-type-package,-C,help:Go package name to use for custom or unknown types"`

	// Out is the output path. If Out is a file, then that will be used as the
	// path. If Out is a directory, then the output file will be
	// Out/<$CWD>.xo.go
	Out string `arg:"-o,help:output path or file name"`

	// Suffix is the output suffix for filenames.
	Suffix string `arg:"-f,help:output file suffix"`

	// SingleFile when toggled changes behavior so that output is to one f ile.
	SingleFile bool `arg:"--single-file,help:toggle single file output"`

	// Schema is the name of the schema to query.
	Schema string `arg:"-s,help:database schema to generate Go types from"`

	// IncTypes are the types to include.
	//IncTypes []string `arg:"--include,-t,help:include types"`

	// ExcTypes are the types to exclude.
	//ExcTypes []string `arg:"--exclude,-x,help:exclude types"`

	// QueryMode toggles whether or not to parse a query from stdin.
	QueryMode bool `arg:"--enable-query-mode,-N,help:enable query parsing mode"`

	// Query is the passed query. If not specified, then os.Stdin will be used.
	// cli args take precedence over Stdin.
	Query string `arg:"-Q,help:sql query to generate Go type and func from"`

	// QueryType is the name to give to the Go type generated from the query.
	QueryType string `arg:"--query-type,-T,help:name for Go type generated from query"`

	// QueryFunc is the name to assign to the generated query func.
	QueryFunc string `arg:"--query-func,-F,help:comment for the generated Go func"`

	// QueryTrim enables triming whitespace on the supplied query.
	QueryTrim bool `arg:"--trim,-M,help:trim whitespace of query when generating Go code"`

	// QueryStrip enables stripping the '::<type> AS <name>' from queries.
	QueryStrip bool `arg:"--query-strip,-B,help:strip '::type AS name' from queries after inspection"`

	// TypeComment is the type comment for a query.
	QueryTypeComment string `arg:"--comment,help:comment for type generated from query"`

	// FuncComment is the func comment to provide the named query.
	QueryFuncComment string `arg:"--func-comment,help:comment for func generated for query"`

	// QueryParamDelimiter is the delimiter for parameterized values for a query.
	QueryParamDelimiter string `arg:"--query-delimiter,-D,help:delimiter for query parameters"`

	// QueryOnlyOne toggles the generated query code to expect only one result.
	QueryOnlyOne bool `arg:"--only-one,-1,help:toggle generated query to return only one result"`

	// NoExtra when toggled will not generate certain extras.
	//NoExtra bool `arg:"--no-extra,-Z,help:"disable extra code generation"`

	// DSN is the database string (ie, pgsql://user@blah:localhost:5432/dbname?args=)
	DSN string `arg:"positional,required,help:data source name"`

	// Path is the output path, as derived from Out.
	Path string `arg:"-"`

	// Filename is the output filename, as derived from Out.
	Filename string `arg:"-"`
}

// CustomTypePackage is a hack.
var CustomTypePackage = ""
