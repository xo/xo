package main

//go:generate go-bindata -pkg templates -prefix templates/ -o templates/tpls.go -ignore .go$ -ignore .swp$ -nometadata -nomemcopy templates/...

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path"
	"sort"

	// database drivers
	"github.com/alexflint/go-arg"
	_ "github.com/lib/pq"

	"github.com/knq/xo/internal"
	"github.com/knq/xo/loaders"
	"github.com/knq/xo/templates"
)

// args is the default command line arguments.
var args = internal.ArgType{
	Schema: "public",
	Suffix: ".xo.go",
	Magic:  true,
}

// schemaLoaders are the various schema loaders.
var schemaLoaders = map[string]loaders.Loader{
	"postgres": loaders.PgLoadTypes,
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
	scheme, db, err := openDB(args.DSN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// create initial type map
	typeMap := map[string]*bytes.Buffer{}
	err = templates.Tpls["xo_db.go.tpl"].Execute(loaders.GetBuf(typeMap, "xo_db"), args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// load defs into type map
	err = schemaLoaders[scheme](&args, db, typeMap)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// output
	err = writeTypes(typeMap)
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

			// error if split was set, but destination is not a directory
			if args.Split {
				return errors.New("supplied path is not directory")
			}
		} else if _, ok := err.(*os.PathError); ok {
			// path error (ie, file doesn't exist yet)
			args.Path = path.Dir(args.Out)
			args.Filename = path.Base(args.Out)

			// error if split was set, but dest doesn't exist
			if args.Split {
				return errors.New("must supply a directory that exists when using split")
			}
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

	// FIXME: hack to do something quickly
	internal.CustomTypePackage = args.CustomTypePackage

	return nil
}

// openDB attempts to open a database connection.
// TODO: check that this is compatible with all sql drivers.
func openDB(dsn string) (string, *sql.DB, error) {
	var err error

	// parse dsn
	u, err := url.Parse(dsn)
	if err != nil {
		return "", nil, err
	}

	// check and fix pgsql scheme
	if u.Scheme == "" {
		return "", nil, errors.New("invalid dsn")
	} else if u.Scheme == "pgsql" {
		u.Scheme = "postgres"
	}

	// open database
	db, err := sql.Open(u.Scheme, u.String())
	if err != nil {
		return "", nil, err
	}

	return u.Scheme, db, err
}

// writeTypes writes the generated definitions.
func writeTypes(out map[string]*bytes.Buffer) error {
	var err error

	// get and sort keys
	i := 0
	keys := make([]string, len(out))
	for k := range out {
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
			err = templates.Tpls["xo_package.go.tpl"].Execute(f, args)
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
