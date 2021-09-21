package jsontpl

import (
	"context"
	"embed"

	"github.com/xo/xo/templates"
	xo "github.com/xo/xo/types"
)

func init() {
	templates.Register("json", &templates.TemplateSet{
		Files:   Files,
		FileExt: ".xo.json",
		Flags: []xo.Flag{
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
	UglyKey   xo.ContextKey = "ugly"
)

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

// Files are the embedded JSON templates.
//
//go:embed *.tpl
var Files embed.FS
