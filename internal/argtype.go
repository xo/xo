package internal

// ArgType is the type that specifies the command line arguments.
type ArgType struct {
	// Verbose enables verbose output.
	Verbose bool `arg:"-v,help:toggle verbose"`

	// Package is the name of the package to generate. If not specified, the
	// name of the output directory will be used instead.
	Package string `arg:"-p,help:output package name"`

	// Out is the output path. If Out is a file, then that will be used as the
	// path. If Out is a directory, then the output file will be
	// Out/<$CWD>.xo.go
	Out string `arg:"-o,help:output file"`

	// Split when toggled splits the output to individual files per type.
	Split bool `arg:"help:split output to one file per type"`

	// Schema is the name of the schema to query.
	Schema string `arg:"-s,help:use specified schema"`

	// CustomTypePackage is the Go package name to use for unknown sql types.
	CustomTypePackage string `arg:"--custom-type-package,-C,help:package to use for custom or unknown types"`

	// Suffix is the output suffix for filenames.
	Suffix string `arg:"-f,help:output file suffix"`

	// Magic toggles magic field renaming.
	Magic bool `arg:"help:toggle magic"`

	// Tables limits the query to the specified tables only.
	Tables []string `arg:"-t,help:limit to specified tables"`

	// DSN is the database string (ie, pgsql://user@blah:localhost:5432/dbname?args=)
	DSN string `arg:"positional,required,help:data source name"`

	// Path is the output path, as derived from Out.
	Path string `arg:"-"`

	// Filename is the output filename, as derived from Out.
	Filename string `arg:"-"`
}

// CustomTypePackage is a hack.
var CustomTypePackage = ""
