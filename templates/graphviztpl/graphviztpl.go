package graphviztpl

import (
	"embed"

	"github.com/xo/xo/templates"
)

//go:embed *.tpl
var Files embed.FS

func init() {
	templates.Register("graphviz", &templates.TemplateSet{
		Files: Files,
	})
}
