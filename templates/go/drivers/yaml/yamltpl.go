package yamltpl

import (
	"context"
	"embed"

	"github.com/xo/xo/templates"
	xo "github.com/xo/xo/types"
)

func init() {
	templates.Register("yaml", &templates.TemplateSet{
		Files:   Files,
		FileExt: ".xo.yaml",
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

// Files are the embedded JSON templates.
//
//go:embed *.tpl
var Files embed.FS
