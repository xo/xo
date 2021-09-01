//go:build tools

package main

import (
	"bytes"
	"context"
	_ "embed"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"text/template"

	"github.com/kenshaw/snaker"
	_ "github.com/lib/pq"
	"github.com/xo/dburl/passfile"
	"github.com/xo/xo/_examples/pgcatalog/pgtypes"
	"mvdan.cc/gofumpt/format"
)

func main() {
	dsn := flag.String("dsn", "pg://", "dsn")
	out := flag.String("out", "pgtypes.go", "out")
	flag.Parse()
	if err := run(context.Background(), *dsn, *out); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, dsn, out string) error {
	// open
	db, err := passfile.Open(dsn, "xopass")
	if err != nil {
		return err
	}
	// retrieve types
	types, err := pgtypes.GetPgTypes(ctx, db)
	if err != nil {
		return err
	}
	sort.Slice(types, func(i, j int) bool {
		return types[i].Name < types[j].Name
	})
	buf, err := gen(types)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(out, buf, 0644)
}

// gen generates the file.
func gen(types []*pgtypes.PgType) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := tpl.Execute(buf, types); err != nil {
		return nil, err
	}
	return format.Source(buf.Bytes(), format.Options{
		ExtraRules: true,
	})
}

// tpl is the type template.
var tpl = template.Must(template.New("types.go.tpl").Funcs(template.FuncMap{
	"identifier": snaker.ForceCamelIdentifier,
}).Parse(string(typesGo)))

//go:embed types.go.tpl
var typesGo []byte
