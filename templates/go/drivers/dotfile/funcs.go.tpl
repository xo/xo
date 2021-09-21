// Package funcs provides custom template funcs.
package funcs

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/kenshaw/snaker"
	"github.com/xo/xo/templates/dottpl"
	xo "github.com/xo/xo/types"
)

// Init intializes the custom template funcs.
func Init(ctx context.Context) (template.FuncMap, error) {
	driver, schema, _ := xo.DriverSchemaNthParam(ctx)
	// parse row template
	row, err := template.New("row").Parse(dottpl.Row(ctx))
	if err != nil {
		return nil, err
	}
	funcs := &Funcs{
		driver:    driver,
		schema:    schema,
		defaults:  dottpl.Defaults(ctx),
		bold:      dottpl.Bold(ctx),
		color:     dottpl.Color(ctx),
		row:       row,
		direction: dottpl.Direction(ctx),
	}
	return funcs.FuncMap(), nil
}

// Funcs is a set of template funcs.
type Funcs struct {
	driver    string
	schema    string
	defaults  []string
	bold      bool
	color     string
	row       *template.Template
	direction bool
}

// FuncMap returns the func map.
func (f *Funcs) FuncMap() template.FuncMap {
	return template.FuncMap{
		"schema":    f.schemafn,
		"defaults":  f.defaultsfn,
		"header":    f.header,
		"row":       f.rowfn,
		"edge":      f.edge,
		"quotes":    quotes,
		"normalize": normalize,
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

func (f *Funcs) edge(table xo.Table, fkey xo.ForeignKey, i int) string {
	node, toNode := f.schemafn(table.Name), f.schemafn(fkey.RefTable)
	row, toRow := quotes(fkey.Fields[i].Name), quotes(fkey.RefFields[i].Name)
	var dirFrom, dirTo string
	if f.direction {
		dirFrom, dirTo = ":e", ":w"
	}
	// "table":"col":e -> "reftable":"refcol":w
	return fmt.Sprintf("%s:%s%s -> %s:%s%s", node, row, dirFrom, toNode, toRow, dirTo)
}

// schemafn takes a series of names and joins them with the schema name.
func (f *Funcs) schemafn(names ...string) string {
	s, n := f.schema, strings.Join(names, ".")
	switch {
	case s == "" && n == "":
		return ""
	case f.driver == "sqlite3":
		return quotes(n)
	}
	return quotes(s + "." + n)
}

func (f *Funcs) defaultsfn() []string {
	return f.defaults
}

func quotes(v string) string {
	return fmt.Sprintf("%q", v)
}

func normalize(v string) string {
	return snaker.CamelToSnakeIdentifier(snaker.ForceCamelIdentifier(v))
}
