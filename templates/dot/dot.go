//go:build xotpl

package dot

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"text/template"

	"github.com/kenshaw/snaker"
	xo "github.com/xo/xo/types"
)

// Init registers the template.
func Init(ctx context.Context, f func(xo.TemplateType)) error {
	f(xo.TemplateType{
		Modes: []string{"schema"},
		Flags: []xo.Flag{
			{
				ContextKey: DefaultsKey,
				Type:       "string",
				Desc:       "default statements",
				Default:    "node [shape=none, margin=0]",
			},
			{
				ContextKey: BoldKey,
				Type:       "bool",
				Desc:       "bold header row",
				Default:    "false",
			},
			{
				ContextKey: ColorKey,
				Type:       "string",
				Desc:       "header color",
				Default:    "lightblue",
			},
			{
				ContextKey: RowKey,
				Type:       "string",
				Desc:       "row value template",
				Default:    "{{ .Name }}: {{ .Type.Type }}",
			},
			{
				ContextKey: DirectionKey,
				Type:       "bool",
				Desc:       "enable edge directions",
				Default:    "true",
			},
		},
		Funcs: NewFuncs,
		Process: func(ctx context.Context, _ string, set *xo.Set, emit func(xo.Template)) error {
			if len(set.Schemas) == 0 {
				return errors.New("dot template must be passed at least one schema")
			}
			for _, schema := range set.Schemas {
				emit(xo.Template{
					Partial:  "dot",
					Dest:     "xo.xo.dot",
					SortName: schema.Name,
					Data:     schema,
				})
			}
			return nil
		},
	})
	return nil
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

// NewFuncs creates a set of template funcs for the context.
func NewFuncs(ctx context.Context, _ string) (template.FuncMap, error) {
	driver, _, schema := xo.DriverDbSchema(ctx)
	// parse row template
	row, err := template.New("row").Parse(Row(ctx))
	if err != nil {
		return nil, err
	}
	funcs := &Funcs{
		driver:    driver,
		schema:    schema,
		defaults:  Defaults(ctx),
		bold:      Bold(ctx),
		color:     Color(ctx),
		row:       row,
		direction: Direction(ctx),
	}
	return funcs.FuncMap(), nil
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

// Context keys.
var (
	DefaultsKey  xo.ContextKey = "defaults"
	DirectionKey xo.ContextKey = "direction"
	BoldKey      xo.ContextKey = "bold"
	RowKey       xo.ContextKey = "row"
	ColorKey     xo.ContextKey = "color"
)

// Defaults returns defaults from the context.
func Defaults(ctx context.Context) []string {
	s, _ := ctx.Value(DefaultsKey).([]string)
	return s
}

// Bold returns bold from the context.
func Bold(ctx context.Context) bool {
	b, _ := ctx.Value(BoldKey).(bool)
	return b
}

// Color returns color from the context.
func Color(ctx context.Context) string {
	s, _ := ctx.Value(ColorKey).(string)
	return s
}

// Row returns row from the context.
func Row(ctx context.Context) string {
	s, _ := ctx.Value(RowKey).(string)
	return s
}

// Direction returns direction from the context.
func Direction(ctx context.Context) bool {
	b, _ := ctx.Value(DirectionKey).(bool)
	return b
}
