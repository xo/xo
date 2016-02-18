package main

//go:generate ./tpl.sh
//go:generate ./gen.sh models

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"

	// database drivers
	"github.com/alexflint/go-arg"

	"github.com/knq/xo/internal"
	_ "github.com/knq/xo/loaders"
	"github.com/knq/xo/models"
)

// args is the default command line arguments.
var args = &internal.ArgType{
	Suffix:              ".xo.go",
	Int32Type:           "int",
	Uint32Type:          "uint",
	QueryParamDelimiter: "%%",

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

func main() {
	var err error

	// parse args
	arg.MustParse(args)

	// process args
	err = processArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// open database
	loader, db, err := openDB(args.DSN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// set loader
	args.Loader = loader

	// load defs into type map
	if args.QueryMode {
		err = loader.ParseQuery(args, db)
	} else {
		err = loader.LoadSchemaTypes(args, db)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// add xo
	err = args.ExecuteTemplate(internal.XO, "xo_db", args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// output
	err = writeTypes(internal.TBufSlice(args.Generated))
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

			// error if not split was set, but destination is not a directory
			if !args.SingleFile {
				return errors.New("output path is not directory")
			}
		} else if _, ok := err.(*os.PathError); ok {
			// path error (ie, file doesn't exist yet)
			args.Path = path.Dir(args.Out)
			args.Filename = path.Base(args.Out)

			// error if split was set, but dest doesn't exist
			if !args.SingleFile {
				return errors.New("output path must be a directory and already exist when not writing to a single file")
			}
		} else {
			return err
		}
	}

	// check user template path
	if args.TemplatePath != "" {
		fi, err := os.Stat(args.TemplatePath)
		if err == nil && !fi.IsDir() {
			return errors.New("template path is not directory")
		} else if err != nil {
			return errors.New("template path must exist")
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

	// if query mode toggled, but no query, read Stdin.
	if args.QueryMode && args.Query == "" {
		buf, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		args.Query = string(buf)
	}

	// query mode parsing
	if args.Query != "" {
		args.QueryMode = true
	}

	// check that query type was specified
	if args.QueryMode && args.QueryType == "" {
		return errors.New("query type must be supplied for query parsing mode")
	}

	// query trim
	if args.QueryMode && args.QueryTrim {
		args.Query = strings.TrimSpace(args.Query)
	}

	// if verbose
	if args.Verbose {
		models.XOLog = func(s string, p ...interface{}) {
			fmt.Printf("SQL:\n%s\nPARAMS:\n%v\n\n", s, p)
		}
	}

	return nil
}

// openDB attempts to open a database connection.
func openDB(dsn string) (internal.Loader, *sql.DB, error) {
	var err error

	// parse dsn
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, nil, err
	}
	if u.Scheme == "" {
		return nil, nil, errors.New("invalid dsn")
	}

	var dbtype, connStr string
	var loader internal.Loader
	var ok bool

	// find schema loader
	for dbtype, loader = range internal.SchemaLoaders {
		if connStr, ok = loader.IsSupported(u); ok {
			break
		}
	}
	if !ok {
		return nil, nil, errors.New("unknown scheme in dsn")
	}

	// open database
	db, err := sql.Open(dbtype, connStr)
	if err != nil {
		return nil, nil, err
	}

	return loader, db, err
}

// files is a map of filenames to open file handles.
var files = map[string]*os.File{}

// getFile builds the filepath from the TBuf information, and retrieves the
// file from files. If the built filename is not already defined, then it calls
// the os.OpenFile with the correct parameters depending on the state of args.
func getFile(args *internal.ArgType, t *internal.TBuf) (*os.File, error) {
	var f *os.File
	var err error

	// determine filename
	var filename = strings.ToLower(t.Name) + args.Suffix
	if args.SingleFile {
		filename = args.Filename
	}
	filename = path.Join(args.Path, filename)

	// lookup file
	f, ok := files[filename]
	if ok {
		return f, nil
	}

	// default open mode
	mode := os.O_RDWR | os.O_CREATE | os.O_TRUNC

	// stat file to determine if file already exists
	fi, err := os.Stat(filename)
	if err == nil && fi.IsDir() {
		return nil, errors.New("filename cannot be directory")
	} else if _, ok = err.(*os.PathError); !ok && args.Append && t.Type != internal.XO {
		// file exists so append if append is set and not XO type
		mode = os.O_APPEND | os.O_WRONLY
	}

	// skip
	if t.Type == internal.XO && fi != nil {
		return nil, nil
	}

	// open file
	f, err = os.OpenFile(filename, mode, 0666)
	if err != nil {
		return nil, err
	}

	// file didn't originally exist, so add package header
	if fi == nil || !args.Append {
		err = args.TemplateSet().Execute(f, "xo_package.go.tpl", args)
		if err != nil {
			return nil, err
		}
	}

	// store file
	files[filename] = f

	return f, nil
}

// writeTypes writes the generated definitions.
func writeTypes(out internal.TBufSlice) error {
	var err error

	// sort segments
	sort.Sort(out)

	// loop, writing in order
	for _, t := range out {
		var f *os.File

		// skip when in append and type is XO
		if args.Append && t.Type == internal.XO {
			continue
		}

		// get file and filename
		f, err = getFile(args, &t)
		if err != nil {
			return err
		}

		// should only be nil when type == xo
		if f == nil {
			continue
		}

		// write segment
		if !args.Append || (t.Type != internal.Model && t.Type != internal.QueryModel) {
			_, err = t.Buf.WriteTo(f)
			if err != nil {
				return err
			}
		}
	}

	// build goimports parameters, closing files
	params := []string{"-w"}
	for k, f := range files {
		params = append(params, k)

		// close
		err = f.Close()
		if err != nil {
			return err
		}
	}

	// process written files with goimports
	return exec.Command("goimports", params...).Run()
}
