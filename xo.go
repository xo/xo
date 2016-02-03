package main

//go:generate go-bindata -pkg templates -prefix templates/ -o templates/templates.go -ignore .go$ -ignore .swp$ -nometadata -nomemcopy templates/...

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"

	// database drivers
	"github.com/gedex/inflector"
	_ "github.com/lib/pq"
	"github.com/serenize/snaker"

	"github.com/alexflint/go-arg"
	"github.com/knq/xo/models"
	"github.com/knq/xo/templates"
)

// args contains the command line arguments.
var args = struct {
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
}{
	Schema: "public",
	Suffix: ".xo.go",
	Magic:  true,
}

func main() {
	var err error

	// parse args
	arg.MustParse(&args)

	// process args
	err = processArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// open database
	db, err := openDB(args.DSN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// load database defs
	defs, err := loadDefs(db)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// output
	err = writeDefs(defs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// processArgs processs cli args.
func processArgs() error {
	var err error

	// get working directory
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// determine out path
	if args.Out == "" {
		args.Path = cwd
	} else {
		// determine what to do with Out
		fi, err := os.Stat(args.Out)
		if err == nil && fi.IsDir() {
			// out is directory
			args.Path = args.Out
		} else if err == nil && !fi.IsDir() {
			// file exists (will truncate later)
			args.Path = path.Dir(args.Out)
			args.Filename = path.Base(args.Out)
		} else if _, ok := err.(*os.PathError); ok {
			// path error (ie, file doesn't exist yet)
			args.Path = path.Dir(args.Out)
			args.Filename = path.Base(args.Out)
		} else {
			return err
		}
	}

	// fix path
	if args.Path == "." {
		args.Path = cwd
	}

	// determine package name
	if args.Package == "" {
		args.Package = path.Base(args.Path)
	}

	// determine filename if not previously set
	if args.Filename == "" {
		args.Filename = args.Package + args.Suffix
	}

	return nil
}

// openDB attempts to open a database connection.
// TODO: check that this is compatible with all sql drivers.
func openDB(dsn string) (*sql.DB, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}

	// check and fix scheme
	if u.Scheme == "" {
		return nil, errors.New("invalid dsn")
	} else if u.Scheme == "pgsql" {
		u.Scheme = "postgres"
	}

	return sql.Open(u.Scheme, u.String())
}

var (
	// tracking of types encountered
	knownTypeMap = map[string]bool{}

	// template func map
	tplFuncMap = template.FuncMap{
		"inc": func(i int) int {
			return i + 1
		},
		"fields": func(fields []*models.Column, pkField string) string {
			str := ""
			i := 0
			for _, col := range fields {
				if col.Field == pkField {
					continue
				}

				if i != 0 {
					str = str + ", "
				}
				str = str + col.ColumnName
				i++
			}

			return str
		},
		"values": func(fields []*models.Column, pkField string) string {
			str := ""
			i := 1
			for _, col := range fields {
				if col.Field == pkField {
					continue
				}

				if i != 1 {
					str = str + ", "
				}
				str = str + "$" + strconv.Itoa(i)
				i++
			}

			return str
		},
		"cols": func(fields []*models.Column, pkField string, prefix string) string {
			str := ""
			i := 0
			for _, col := range fields {
				if col.Field == pkField {
					continue
				}

				if i != 0 {
					str = str + ", "
				}
				str = str + prefix + "." + col.Field
				i++
			}

			return str
		},
		"count": func(fields []*models.Column, pkField string) int {
			i := 1
			for _, col := range fields {
				if col.Field == pkField {
					continue
				}

				i++
			}
			return i
		},
		"retype": func(typ string) string {
			if strings.Contains(typ, ".") {
				return typ
			}

			prefix := ""
			for strings.HasPrefix(typ, "[]") {
				typ = typ[2:]
				prefix = prefix + "[]"
			}

			if _, ok := knownTypeMap[typ]; !ok {
				pkg := args.CustomTypePackage
				if pkg != "" {
					pkg = pkg + "."
				}

				return prefix + pkg + typ
			}

			return prefix + typ
		},
		"reniltype": func(typ string) string {
			if strings.Contains(typ, ".") {
				return typ
			}

			if strings.HasSuffix(typ, "{}") {
				pkg := args.CustomTypePackage
				if pkg != "" {
					pkg = pkg + "."
				}

				return pkg + typ
			}

			return typ
		},
	}

	// templates
	tpls = map[string]*template.Template{}
)

func init() {
	// load template assets
	for _, n := range templates.AssetNames() {
		buf := templates.MustAsset(n)
		tpls[n] = template.Must(template.New(n).Funcs(tplFuncMap).Parse(string(buf)))
	}

	for _, typ := range []string{
		"bool", "string", "byte", "rune",
		"int", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64",
	} {
		knownTypeMap[typ] = true
	}
}

// EnumTemplate is a template item for a enum.
type EnumTemplate struct {
	Type       string
	TypeNative string
	Values     []*models.Enum
}

// TableTemplate is a template item for a table.
type TableTemplate struct {
	Type            string
	TableSchema     string
	TableName       string
	PrimaryKey      string
	PrimaryKeyField string
	Fields          []*models.Column
}

var (
	lenRE = regexp.MustCompile(`\([0-9]+\)$`)
)

// parseType calculates the go type based on a column definition.
func parseType(dt string, nullable bool) (int, string, string) {
	precision := 0
	nilType := "nil"
	asSlice := false

	// handle SETOF
	if strings.HasPrefix(dt, "SETOF ") {
		_, _, t := parseType(dt[len("SETOF "):], false)
		return 0, "nil", "[]" + t
	}

	// determine if it's a slice
	if strings.HasSuffix(dt, "[]") {
		dt = dt[:len(dt)-2]
		asSlice = true
	}

	// extract length
	if loc := lenRE.FindStringIndex(dt); len(loc) != 0 {
		precision, _ = strconv.Atoi(dt[loc[0]+1 : loc[1]-1])
		dt = dt[:loc[0]]
	}

	var typ string
	switch dt {
	case "boolean":
		nilType = "false"
		typ = "bool"
		if nullable {
			nilType = "sql.NullBool{}"
			typ = "sql.NullBool"
		}

	case "character", "character varying", "text":
		nilType = `""`
		typ = "string"
		if nullable {
			nilType = "sql.NullString{}"
			typ = "sql.NullString"
		}

	case "smallint":
		nilType = "0"
		typ = "int16"
		if nullable {
			nilType = "sql.NullInt64{}"
			typ = "sql.NullInt64"
		}
	case "integer":
		nilType = "0"
		typ = "int32"
		if nullable {
			nilType = "sql.NullInt64{}"
			typ = "sql.NullInt64"
		}
	case "bigint":
		nilType = "0"
		typ = "int64"
		if nullable {
			nilType = "sql.NullInt64{}"
			typ = "sql.NullInt64"
		}

	case "smallserial":
		nilType = "0"
		typ = "uint16"
		if nullable {
			nilType = "sql.NullInt64{}"
			typ = "sql.NullInt64"
		}
	case "serial":
		nilType = "0"
		typ = "uint32"
		if nullable {
			nilType = "sql.NullInt64{}"
			typ = "sql.NullInt64"
		}
	case "bigserial":
		nilType = "0"
		typ = "uint64"
		if nullable {
			nilType = "sql.NullInt64{}"
			typ = "sql.NullInt64"
		}

	case "real":
		nilType = "0.0"
		typ = "float32"
		if nullable {
			nilType = "sql.NullFloat64{}"
			typ = "sql.NullFloat64"
		}
	case "double precision":
		nilType = "0.0"
		typ = "float64"
		if nullable {
			nilType = "sql.NullFloat64{}"
			typ = "sql.NullFloat64"
		}

	case "bytea":
		asSlice = true
		typ = "byte"

	case "timestamp with time zone":
		typ = "*time.Time"

	case "time with time zone", "time without time zone", "timestamp without time zone":
		nilType = "0"
		typ = "int64"

	case "interval":
		typ = "*time.Duration"

	case `"char"`, "bit":
		// FIXME: this needs to actually be tested ...
		// i think this should be 'rune' but I don't think database/sql
		// supports 'rune' as a type?
		//
		// this is mainly here because postgres's pg_catalog.* meta tables have
		// this as a type.
		//typ = "rune"
		nilType = `uint8(0)`
		typ = "uint8"

	case `"any"`, "bit varying":
		asSlice = true
		typ = "byte"

	default:
		if strings.HasPrefix(dt, args.Schema+".") {
			// in the same schema, so chop off
			typ = snaker.SnakeToCamel(dt[len(args.Schema)+1:])
			nilType = typ + "(0)"
		} else {
			typ = snaker.SnakeToCamel(dt)
			nilType = typ + "{}"
		}
	}

	if asSlice {
		typ = "[]" + typ
		nilType = "nil"
	}

	return precision, nilType, typ
}

// getBuf retrieves the name from m, creating it if necessary.
func getBuf(m map[string]*bytes.Buffer, name string) *bytes.Buffer {
	buf, ok := m[name]
	if !ok {
		m[name] = new(bytes.Buffer)
		return m[name]
	}

	return buf
}

// loadEnums reads the enums from the database, writing the values to the
// typeMap and returning the created EnumTemplates.
func loadEnums(db *sql.DB, typeMap map[string]*bytes.Buffer) (map[string]*EnumTemplate, error) {
	var err error

	// load enums
	enums, err := models.EnumsBySchema(db, args.Schema)
	if err != nil {
		return nil, err
	}

	// process enums
	enumMap := make(map[string]*EnumTemplate)
	for _, e := range enums {
		// grab type
		typ := e.Type

		// set enum info
		e.Type = snaker.SnakeToCamel(typ)
		e.EnumType = snaker.SnakeToCamel(strings.ToLower(e.Value))
		knownTypeMap[e.EnumType] = true

		// set value in enum map if not present
		if _, ok := enumMap[typ]; !ok {
			enumMap[typ] = &EnumTemplate{
				Type:       e.Type,
				TypeNative: typ,
				Values:     make([]*models.Enum, 0),
			}
		}

		// append enum to template vals
		enumMap[typ].Values = append(enumMap[typ].Values, e)
	}

	// generate enum templates
	for typ, em := range enumMap {
		buf := getBuf(typeMap, strings.ToLower(snaker.SnakeToCamel(typ)))
		err = tpls["enum.go.tpl"].Execute(buf, em)
		if err != nil {
			return nil, err
		}
	}

	return enumMap, nil
}

// loadProcs reads the procs from the database, writing the values to the
// typeMap and returning the created ProcTemplates.
func loadProcs(db *sql.DB, typeMap map[string]*bytes.Buffer) (map[string]*models.Proc, error) {
	var err error

	// load procs
	procs, err := models.ProcsBySchema(db, args.Schema)
	if err != nil {
		return nil, err
	}

	// process procs
	procMap := make(map[string]*models.Proc)
	for _, p := range procs {
		p.Schema = args.Schema

		// fix the name if it starts with underscore
		name := p.Name
		if name[0:1] == "_" {
			name = name[1:]
		}

		// convert name and types to go values
		p.FuncName = snaker.SnakeToCamel(name)
		_, p.GoNilReturnType, p.GoReturnType = parseType(p.ReturnType, false)
		p.GoParameterTypes = make([]string, 0)

		if len(p.ParameterTypes) > 0 {
			// determine the go equivalent parameter types
			for _, typ := range strings.Split(p.ParameterTypes, ",") {
				_, _, pt := parseType(strings.TrimSpace(typ), false)
				p.GoParameterTypes = append(p.GoParameterTypes, pt)
			}
		}

		procMap[strings.ToLower("sp_"+p.FuncName)] = p
	}

	// generate proc templates
	for typ, pm := range procMap {
		buf := getBuf(typeMap, typ)
		err = tpls["proc.go.tpl"].Execute(buf, pm)
		if err != nil {
			return nil, err
		}
	}

	return procMap, nil
}

// loadTables loads the table definitions from the database, writing the
// resulting templates to typeMap and returning the created TableTemplates.
func loadTables(db *sql.DB, typeMap map[string]*bytes.Buffer) (map[string]*TableTemplate, error) {
	var err error

	// load columns
	cols, err := models.ColumnsByTableSchema(db, args.Schema)
	if err != nil {
		return nil, err
	}

	// process columns
	fieldMap := make(map[string]map[string]bool)
	tableMap := make(map[string]*TableTemplate)
	for _, c := range cols {
		typ := c.TableName

		// set col info
		c.Field = snaker.SnakeToCamel(c.ColumnName)
		c.Len, c.GoNilType, c.GoType = parseType(c.DataType, c.IsNullable)

		// set value in table map if not present
		if _, ok := tableMap[typ]; !ok {
			tableMap[typ] = &TableTemplate{
				Type:        inflector.Singularize(snaker.SnakeToCamel(typ)),
				TableSchema: args.Schema,
				TableName:   c.TableName,
				Fields:      make([]*models.Column, 0),
			}
		}

		// set primary key
		if c.IsPrimaryKey {
			tableMap[typ].PrimaryKey = c.ColumnName
			tableMap[typ].PrimaryKeyField = c.Field
		}

		// create field map if not already made
		if _, ok := fieldMap[typ]; !ok {
			fieldMap[typ] = make(map[string]bool)
		}

		// check fieldmap
		if _, ok := fieldMap[typ][c.ColumnName]; !ok {
			// append col to template fields
			tableMap[typ].Fields = append(tableMap[typ].Fields, c)
		}

		// set field map
		fieldMap[typ][c.ColumnName] = true
	}

	// generate table templates
	for typ, t := range tableMap {
		buf := getBuf(typeMap, strings.ToLower(snaker.SnakeToCamel(typ)))
		err = tpls["model.go.tpl"].Execute(buf, t)
		if err != nil {
			return nil, err
		}
	}

	return tableMap, nil
}

// loadDefs loads the table definitions from the database.
func loadDefs(db *sql.DB) (map[string]*bytes.Buffer, error) {
	var err error

	typeMap := make(map[string]*bytes.Buffer)

	// add base db type
	buf := getBuf(typeMap, strings.ToLower(args.Schema+"_db"))
	err = tpls["db.go.tpl"].Execute(buf, args)
	if err != nil {
		return nil, err
	}

	// load enums
	_, err = loadEnums(db, typeMap)
	if err != nil {
		return nil, err
	}

	// load tables
	tableMap, err := loadTables(db, typeMap)
	if err != nil {
		return nil, err
	}

	// load procs
	_, err = loadProcs(db, typeMap)
	if err != nil {
		return nil, err
	}

	tableMap = tableMap

	// loop over tables
	//_, err = loadIndexes(db, typeMap, tableMap)

	return typeMap, nil
}

// writeDefs writes the generated definitions.
func writeDefs(out map[string]*bytes.Buffer) error {
	var err error

	// get and sort keys
	keys := make([]string, len(out))
	i := 0
	for k, _ := range out {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	// loop over out and write in order
	var f *os.File
	goimportsParams := []string{"-w"}
	for _, k := range keys {
		// open file for writing
		if args.Split || f == nil {
			// determine filename
			var filename = k + args.Suffix
			if !args.Split {
				filename = args.Filename
			}
			filename = path.Join(args.Path, filename)

			// append filename to goimportsParams
			goimportsParams = append(goimportsParams, filename)

			// create file
			f, err = os.Create(filename)
			if err != nil {
				return err
			}

			// write package header
			err = tpls["package.go.tpl"].Execute(f, args)
			if err != nil {
				return err
			}
		}

		// write segment
		_, err = out[k].WriteTo(f)
		if err != nil {
			return err
		}

		// close file
		if args.Split {
			err = f.Close()
			if err != nil {
				return err
			}
		}
	}

	// close the last file
	if !args.Split {
		f.Close()
	}

	// process written files with goimports
	return exec.Command("goimports", goimportsParams...).Run()
}
