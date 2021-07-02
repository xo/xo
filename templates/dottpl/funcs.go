package dottpl

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/xo/xo/templates"
	xo "github.com/xo/xo/types"
)

// Funcs is a set of template funcs.
type Funcs struct {
	schema string
	driver string

	defaults  []string
	bold      bool
	color     string
	row       *template.Template
	direction bool
}

// NewFuncs creates a new Funcs
func NewFuncs(ctx context.Context) (*Funcs, error) {
	// parse row template
	row, err := template.New("row").Parse(Row(ctx))
	if err != nil {
		return nil, err
	}
	return &Funcs{
		schema:    templates.Schema(ctx),
		driver:    templates.Driver(ctx),
		defaults:  Defaults(ctx),
		bold:      Bold(ctx),
		color:     Color(ctx),
		row:       row,
		direction: Direction(ctx),
	}, err
}

// FuncMap returns the func map.
func (f *Funcs) FuncMap() template.FuncMap {
	return template.FuncMap{
		"schema": f.schemafn,

		"defaults": f.defaultsfn,
		"header":   f.header,
		"row":      f.rowfn,
		"edge":     f.edge,

		"quotes": f.quotes,
	}
}

func (f *Funcs) header(text string) string {
	color := fmt.Sprintf(" bgcolor=%q", f.color)
	if f.color == "" {
		color = ""
	}
	if f.bold {
		text = "<B>" + text + "</B>"
	}
	return fmt.Sprintf("<td%s>%s</td>", color, text)
}

func (f *Funcs) rowfn(field xo.Field) string {
	buf := new(bytes.Buffer)
	if err := f.row.Funcs(f.FuncMap()).Execute(buf, field); err != nil {
		return fmt.Sprintf("[[ error: %v ]]", err)
	}
	return fmt.Sprintf(`<td align="left" PORT=%q>%s</td>`, field.Name, buf.String())
}

func (f *Funcs) edge(node, row, toNode, toRow string) string {
	dirFrom, dirTo := ":e", ":w"
	if !f.direction {
		dirFrom, dirTo = "", ""
	}
	// "table":"col":e -> "reftable":"refcol":w
	return fmt.Sprintf("%s:%s%s -> %s:%s%s",
		node, row, dirFrom,
		toNode, toRow, dirTo)
}

// schemafn takes a series of names and joins them with the schema name.
func (f *Funcs) schemafn(names ...string) string {
	s, n := f.schema, strings.Join(names, ".")
	switch {
	case s == "" && n == "":
		return ""
	case f.driver == "sqlite3":
		return f.quotes(n)
	}
	return f.quotes(s + "." + n)
}

func (f *Funcs) defaultsfn() []string {
	return f.defaults
}

func (f *Funcs) quotes(v interface{}) string {
	return fmt.Sprintf("%q", v)
}
