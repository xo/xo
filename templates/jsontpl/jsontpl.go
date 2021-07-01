package jsontpl

import (
	"context"
	"embed"
	"text/template"

	"github.com/xo/xo/templates"
	xo "github.com/xo/xo/types"
)

func init() {
	templates.Register("json", &templates.TemplateSet{
		Files:   Files,
		FileExt: ".xo.json",
		Flags: []templates.Flag{
			{
				ContextKey: IndentKey,
				Desc:       "indentation",
				Default:    "  ",
				Value:      "",
			},
		},
		Funcs: func(ctx context.Context) (template.FuncMap, error) {
			return NewFuncs(ctx).FuncMap(), nil
		},
		FileName: func(ctx context.Context, tpl *templates.Template) string {
			return tpl.Name
		},
		Process: func(ctx context.Context, _ bool, set *templates.TemplateSet, v *xo.XO) error {
			return set.Emit(ctx, &templates.Template{
				Name:     "xo",
				Template: "xo",
				Data:     v,
			})
		},
		Order: []string{"xo"},
	})
}

// Context keys.
const (
	IndentKey xo.ContextKey = "indent"
)

// Indent returns the indent from the context.
func Indent(ctx context.Context) string {
	s, _ := ctx.Value(IndentKey).(string)
	return s
}

// Files are the embedded JSON templates.
//
//go:embed *.tpl
var Files embed.FS
