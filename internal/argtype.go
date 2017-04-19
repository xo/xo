package internal

import "database/sql"

// ArgType is the type that specifies the command line arguments.
type ArgType struct {
	// Verbose enables verbose output.
	Verbose bool `arg:"-v,help:toggle verbose"`

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

	// IgnoreFields allows the user to specify field names which should not be
	// handled by xo in the generated code.
	IgnoreFields []string `arg:"--ignore-fields,help:fields to exclude from the generated Go code types"`

	// ForeignKeyMode is the foreign key mode for generating foreign key names.
	ForeignKeyMode *FkMode `arg:"--fk-mode,-k,help:sets mode for naming foreign key funcs in generated Go code [values: <smart|parent|field|key>]"`

	// UseIndexNames toggles using index names.
	//
	// This is not enabled by default, because index names are often generated
	// using database design software which has the nasty habit of giving
	// non-helpful names to indexes as things like 'authors__b124214__u_idx'
	// instead of 'authors_title_idx'.
	UseIndexNames bool `arg:"--use-index-names,-j,help:use index names as defined in schema for generated Go code"`

	// UseReversedEnumConstNames toggles using reversed enum names.
	UseReversedEnumConstNames bool `arg:"--use-reversed-enum-const-names,-R,help:use reversed enum names for generated consts in Go code"`

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

	// QueryStrip enables stripping the '::<type> AS <name>' from supplied query.
	QueryStrip bool `arg:"--query-strip,-B,help:toggle stripping type casts from query in generated Go code"`

	// QueryInterpolate enables interpolation in generated query.
	QueryInterpolate bool `arg:"--query-interpolate,-I,help:toggle query interpolation in generated Go code"`

	// TypeComment is the type comment for a query.
	QueryTypeComment string `arg:"--query-type-comment,help:comment for query's generated Go type"`

	// FuncComment is the func comment to provide the named query.
	QueryFuncComment string `arg:"--query-func-comment,help:comment for query's generated Go func"`

	// QueryParamDelimiter is the delimiter for parameterized values for a query.
	QueryParamDelimiter string `arg:"--query-delimiter,-D,help:delimiter for query's embedded Go parameters"`

	// QueryFields are the fields to scan the result to.
	QueryFields string `arg:"--query-fields,-Z,help:comma separated list of field names to scan query's results to the query's associated Go type"`

	// QueryAllowNulls indicates that custom query results can contain null types.
	QueryAllowNulls bool `arg:"--query-allow-nulls,-U,help:use query column NULL state"`

	// EscapeAll toggles escaping schema, table, and column names in SQL queries.
	EscapeAll bool `arg:"--escape-all,-X,help:escape all names in SQL queries"`

	// EscapeSchemaName toggles escaping schema name in SQL queries.
	EscapeSchemaName bool `arg:"--escape-schema,-z,help:escape schema name in SQL queries"`

	// EscapeTableNames toggles escaping table names in SQL queries.
	EscapeTableNames bool `arg:"--escape-table,-y,help:escape table names in SQL queries"`

	// EscapeColumnNames toggles escaping column names in SQL queries.
	EscapeColumnNames bool `arg:"--escape-column,-x,help:escape column names in SQL queries"`

	// EnablePostgresOIDs toggles postgres oids.
	EnablePostgresOIDs bool `arg:"--enable-postgres-oids,help:enable postgres oids"`

	// NameConflictSuffix is the suffix used when a name conflicts with a scoped Go variable.
	NameConflictSuffix string `arg:"--name-conflict-suffix,-w,help:suffix to append when a name conflicts with a Go variable"`

	// TemplatePath is the path to use the user supplied templates instead of
	// the built in versions.
	TemplatePath string `arg:"--template-path,help:user supplied template path"`

	// Tags is the list of build tags to add to generated Go files.
	Tags string `arg:"--tags,help:build tags to add to package header"`

	// Path is the output path, as derived from Out.
	Path string `arg:"-"`

	// Filename is the output filename, as derived from Out.
	Filename string `arg:"-"`

	// LoaderType is the loader type.
	LoaderType string `arg:"-"`

	// Loader is the schema loader.
	Loader Loader `arg:"-"`

	// DB is the opened database handle.
	DB *sql.DB `arg:"-"`

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

// NewDefaultArgs returns the default arguments.
func NewDefaultArgs() *ArgType {
	fkMode := FkModeSmart

	return &ArgType{
		Suffix:              ".xo.go",
		Int32Type:           "int",
		Uint32Type:          "uint",
		ForeignKeyMode:      &fkMode,
		QueryParamDelimiter: "%%",
		NameConflictSuffix:  "Val",

		// KnownTypeMap is the collection of known Go types.
		KnownTypeMap: map[string]bool{
			"bool":        true,
			"string":      true,
			"byte":        true,
			"rune":        true,
			"int":         true,
			"int16":       true,
			"int32":       true,
			"int64":       true,
			"uint":        true,
			"uint8":       true,
			"uint16":      true,
			"uint32":      true,
			"uint64":      true,
			"float32":     true,
			"float64":     true,
			"Slice":       true,
			"StringSlice": true,
		},

		// ShortNameTypeMap is the collection of Go style short names for types, mainly
		// used for use with declaring a func receiver on a type.
		ShortNameTypeMap: map[string]string{
			"bool":        "b",
			"string":      "s",
			"byte":        "b",
			"rune":        "r",
			"int":         "i",
			"int16":       "i",
			"int32":       "i",
			"int64":       "i",
			"uint":        "u",
			"uint8":       "u",
			"uint16":      "u",
			"uint32":      "u",
			"uint64":      "u",
			"float32":     "f",
			"float64":     "f",
			"Slice":       "s",
			"StringSlice": "ss",
		},
	}
}

// Description provides the description for the command output.
func (a *ArgType) Description() string {
	return `xo is a command line utility to generate Go code from a database schema.
`
}

// Args are the application arguments.
var Args *ArgType
