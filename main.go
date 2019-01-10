package main

//go:generate ./tpl.sh
//go:generate ./gen.sh models

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/knq/snaker"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"

	"github.com/alexflint/go-arg"
	"gopkg.in/yaml.v2"

	"github.com/JLightning/xo/internal"
	_ "github.com/JLightning/xo/loaders"
	"github.com/JLightning/xo/models"
	"github.com/xo/dburl"
	_ "github.com/xo/xoutil"
)

func main() {
	// circumvent all logic to just determine if xo was built with oracle
	// support
	if len(os.Args) == 2 && os.Args[1] == "--has-oracle-support" {
		var out int
		if _, ok := internal.SchemaLoaders["ora"]; ok {
			out = 1
		}

		fmt.Fprintf(os.Stdout, "%d", out)
		return
	}

	var err error

	// get defaults
	internal.Args = internal.NewDefaultArgs()
	args := internal.Args

	// parse args
	arg.MustParse(args)

	parseXoConfigFile(args)

	// process args
	err = processArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// open database
	err = openDB(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer args.DB.Close()

	// load schema name
	if args.Schema == "" {
		args.Schema, err = args.Loader.SchemaName(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	}

	// load defs into type map
	if args.QueryMode {
		err = args.Loader.ParseQuery(args)
	} else {
		err = args.Loader.LoadSchema(args)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// add xo
	//err = args.ExecuteTemplate(internal.XOTemplate, "xo_db", "", args)
	//if err != nil {
	//	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	//	os.Exit(1)
	//}

	err = args.ExecuteTemplate(internal.PaginationTemplate, "pagination", "", args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	err = args.ExecuteTemplate(internal.PaginationSchemaTemplate, "pagination", "", args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	err = args.ExecuteTemplate(internal.ScalarTemplate, "scalar", "", args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	err = args.ExecuteTemplate(internal.SchemaGraphQLScalarTemplate, "scalar", "", args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	err = args.ExecuteTemplate(internal.WireTemplate, "wire", "", args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// output
	err = writeTypes(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func parseXoConfigFile(args *internal.ArgType) {
	data, err := ioutil.ReadFile("xo_config.yml")
	if err != nil {
		return
	}
	err = yaml.Unmarshal(data, &internal.XoConfig)
	if err != nil {
		return
	}
}

// processArgs processs cli args.
func processArgs(args *internal.ArgType) error {
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

	// escape all
	if args.EscapeAll {
		args.EscapeSchemaName = true
		args.EscapeTableNames = true
		args.EscapeColumnNames = true
	}

	// if verbose
	if args.Verbose {
		models.XOLog = func(s string, p ...interface{}) {
			fmt.Printf("SQL:\n%s\nPARAMS:\n%v\n\n", s, p)
		}
	}

	if args.EntitiesPkg == "" {
		log.Fatal("--entities-pkg: entities package is required")
	}

	return nil
}

// openDB attempts to open a database connection.
func openDB(args *internal.ArgType) error {
	var err error

	// parse dsn
	u, err := dburl.Parse(args.DSN)
	if err != nil {
		return err
	}

	// save driver type
	args.LoaderType = u.Driver

	// grab loader
	var ok bool
	args.Loader, ok = internal.SchemaLoaders[u.Driver]
	if !ok {
		return errors.New("unsupported database type")
	}

	// open database connection
	args.DB, err = sql.Open(u.Driver, u.DSN)
	if err != nil {
		return err
	}

	return nil
}

// files is a map of filenames to open file handles.
var files = map[string]*os.File{}

// getFile builds the filepath from the TBuf information, and retrieves the
// file from files. If the built filename is not already defined, then it calls
// the os.OpenFile with the correct parameters depending on the state of args.
func getFile(args *internal.ArgType, t *internal.TBuf) (*os.File, error) {
	var f *os.File
	var err error

	oldArgPkg := args.Package

	// determine filename
	var filename = strings.ToLower(snaker.CamelToSnake(t.Name))
	if t.TemplateType == internal.SchemaGraphQLTemplate || t.TemplateType == internal.SchemaGraphQLEnumTemplate || t.TemplateType == internal.SchemaGraphQLScalarTemplate || t.TemplateType == internal.PaginationSchemaTemplate {
		filename += ".graphql"
	} else if t.TemplateType == internal.GqlgenModelTemplate {
		filename += ".yml"
	} else if t.TemplateType == internal.WireTemplate {
		filename += ".go"
	} else {
		filename += args.Suffix
	}
	if t.TemplateType == internal.RepositoryTemplate || t.TemplateType == internal.IndexTemplate || t.TemplateType == internal.ForeignKeyTemplate {
		args.Package = "repositories"
		filename = "repositories/" + filename
	} else if t.TemplateType == internal.SchemaGraphQLTemplate || t.TemplateType == internal.SchemaGraphQLEnumTemplate || t.TemplateType == internal.SchemaGraphQLScalarTemplate || t.TemplateType == internal.PaginationSchemaTemplate {
		args.Package = "schema"
		filename = "graphql/schema/" + filename
	} else if t.TemplateType == internal.GqlgenModelTemplate {
		filename = "graphql/" + filename
	} else if t.TemplateType == internal.WireTemplate {
		args.Package = "main"
	} else {
		args.Package = "entities"
		filename = "entities/" + filename
	}
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
	if err == nil {
		if fi.IsDir() {
			return nil, errors.New("filename cannot be directory")
		} else if strings.HasSuffix(filename, "wire.go") {
			fmt.Println("skip wire.go")
			return nil, nil
		}
	} else if _, ok = err.(*os.PathError); !ok && args.Append && t.TemplateType != internal.XOTemplate {
		// file exists so append if append is set and not XO type
		mode = os.O_APPEND | os.O_WRONLY
	}

	// skip
	if t.TemplateType == internal.XOTemplate && fi != nil {
		return nil, nil
	}

	// open file
	f, err = os.OpenFile(filename, mode, 0666)
	if err != nil {
		return nil, err
	}

	// file didn't originally exist, so add package header
	if fi == nil || !args.Append {
		// add build tags
		if args.Tags != "" {
			f.WriteString(`// +build ` + args.Tags + "\n\n")
		}

		if strings.HasSuffix(filename, ".go") {
			if strings.HasSuffix(filename, "wire.go") {
				if _, err = f.WriteString("//+build wireinject\n\npackage main"); err != nil {
					return nil, err
				}
			} else {
				// execute
				err = args.TemplateSet().Execute(f, "xo_package.go.tpl", args)
				if err != nil {
					return nil, err
				}
			}
		} else if strings.HasSuffix(filename, ".yml") {
			err = args.TemplateSet().Execute(f, "gqlgen.yml.tpl", args)
			if err != nil {
				return nil, err
			}
		}
	}

	args.Package = oldArgPkg

	// store file
	files[filename] = f

	return f, nil
}

// writeTypes writes the generated definitions.
func writeTypes(args *internal.ArgType) error {
	var err error

	out := internal.TBufSlice(args.Generated)

	// sort segments
	sort.Sort(out)

	// loop, writing in order
	for _, t := range out {
		var f *os.File

		// skip when in append and type is XO
		if args.Append && t.TemplateType == internal.XOTemplate {
			continue
		}

		// check if generated template is only whitespace/empty
		bufStr := strings.TrimSpace(t.Buf.String())
		if len(bufStr) == 0 {
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
		if !args.Append || (t.TemplateType != internal.TypeTemplate && t.TemplateType != internal.QueryTypeTemplate) {
			_, err = t.Buf.WriteTo(f)
			if err != nil {
				return err
			}
		}
	}

	// build goimports parameters, closing files
	params := []string{"-w"}
	for k, f := range files {
		if strings.HasSuffix(f.Name(), ".go") {
			params = append(params, k)
		}

		// close
		err = f.Close()
		if err != nil {
			return err
		}
	}

	fmt.Println("--- Repositories: ")
	for _, v := range args.NewTemplateFuncs()["reponames"].(func() []string)() {
		if !strings.Contains(v, "Rlts") {
			fmt.Println(v)
		}
	}

	fmt.Println("--- Rlts Repositories: ")
	for _, v := range args.NewTemplateFuncs()["reponames"].(func() []string)() {
		if strings.Contains(v, "Rlts") {
			fmt.Println(v)
		}
	}

	//err = tryMergeGqlgenYml(args)
	if err != nil {
		return err
	}

	// process written files with goimports
	return exec.Command("goimports", params...).Run()
}

func tryMergeGqlgenYml(args *internal.ArgType) error {
	importFile := args.Path + "/graphql/gqlgen_import.yml"
	mainFile := args.Path + "/graphql/gqlgen.yml"
	if _, err := os.Stat(importFile); err == nil {
		data, err := ioutil.ReadFile(importFile)
		if err != nil {
			return err
		}
		var value struct {
			Models map[string]struct {
				Model string
			}
		}
		err = yaml.Unmarshal(data, &value)

		// open file
		f, err := os.OpenFile(mainFile, os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			return err
		}

		var keys []string
		for k := range value.Models {
			keys = append(keys, k)
		}

		sort.Strings(keys)
		for _, k := range keys {
			v := value.Models[k]
			_, err = f.WriteString(fmt.Sprintf("  %s:\n    model: %s\n", k, v.Model))
			if err != nil {
				return err
			}
		}
	}
	return nil
}
