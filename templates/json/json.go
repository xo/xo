//go:build xotpl

package json

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"text/template"

	xo "github.com/xo/xo/types"
)

// Init registers the template.
func Init(ctx context.Context, f func(xo.TemplateType)) error {
	f(xo.TemplateType{
		Modes: []string{"query", "schema"},
		Flags: []xo.Flag{
			{
				ContextKey: FileKey,
				Type:       "string",
				Desc:       "output file",
				Default:    "xo.xo.json",
				Hidden:     true,
			},
			{
				ContextKey: IndentKey,
				Type:       "string",
				Desc:       "indent spacing",
				Default:    "  ",
			},
			{
				ContextKey: UglyKey,
				Type:       "bool",
				Desc:       "disable indentation",
				Default:    "false",
			},
		},
		Funcs: func(ctx context.Context, _ string) (template.FuncMap, error) {
			return template.FuncMap{
				// json marshals v as json.
				"json": func(v interface{}) (string, error) {
					buf := new(bytes.Buffer)
					enc := json.NewEncoder(buf)
					if !Ugly(ctx) {
						enc.SetIndent("", Indent(ctx))
					}
					if err := enc.Encode(v); err != nil {
						return "", err
					}
					return strings.TrimSpace(buf.String()), nil
				},
			}, nil
		},
		Process: func(ctx context.Context, _ string, set *xo.Set, emit func(xo.Template)) error {
			emit(xo.Template{
				Src:  "xo.xo.json.tpl",
				Dest: File(ctx),
				Data: set,
			})
			return nil
		},
	})
	return nil
}

// Context keys.
const (
	FileKey   xo.ContextKey = "file"
	IndentKey xo.ContextKey = "indent"
	UglyKey   xo.ContextKey = "ugly"
)

// File returns file from the context.
func File(ctx context.Context) string {
	s, _ := ctx.Value(FileKey).(string)
	return s
}

// Indent returns indent from the context.
func Indent(ctx context.Context) string {
	s, _ := ctx.Value(IndentKey).(string)
	return s
}

// Ugly returns ugly from the context.
func Ugly(ctx context.Context) bool {
	b, _ := ctx.Value(UglyKey).(bool)
	return b
}
