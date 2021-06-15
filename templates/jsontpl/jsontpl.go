package jsontpl

import (
	"embed"

	"github.com/xo/xo/templates"
)

//go:embed *.tpl
var Files embed.FS

func init() {
	templates.Register("json", &templates.TemplateSet{
		Files: Files,
	})
}
