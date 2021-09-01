package dottpl

import (
	"context"
	"embed"
	"errors"

	"github.com/xo/xo/templates"
	xo "github.com/xo/xo/types"
)

func init() {
	templates.Register("dot", &templates.TemplateSet{
		Files:   Files,
		For:     []string{"schema"},
		FileExt: ".xo.dot",
		Flags: []xo.Flag{
			{
				ContextKey:  DefaultsKey,
				Desc:        "default statements (default: node [shape=none, margin=0])",
				PlaceHolder: `""`,
				Default:     "node [shape=none, margin=0]",
				Value:       []string{},
			},
			{
				ContextKey: BoldKey,
				Desc:       "bold header row",
				Default:    "false",
				Value:      false,
			},
			{
				ContextKey:  ColorKey,
				Desc:        "header color (default: lightblue)",
				PlaceHolder: `""`,
				Default:     "lightblue",
				Value:       "",
			},
			{
				ContextKey:  RowKey,
				Desc:        "row value template (default:  {{ .Name }}: {{ .Datatype.Type }})",
				Default:     "{{ .Name }}: {{ .Datatype.Type }}",
				PlaceHolder: `""`,
				Value:       "",
			},
			{
				ContextKey: DirectionKey,
				Desc:       "enable edge directions",
				Default:    "true",
				Value:      true,
			},
		},
		FileName: func(ctx context.Context, tpl *templates.Template) string {
			return tpl.Name
		},
		Process: func(ctx context.Context, _ bool, set *templates.TemplateSet, v *xo.XO) error {
			if len(v.Schemas) == 0 {
				return errors.New("dot template must be passed at least one schema")
			}
			for _, schema := range v.Schemas {
				if err := set.Emit(ctx, &templates.Template{
					Name:     "xo",
					Template: "xo",
					Data:     schema,
				}); err != nil {
					return err
				}
			}
			return nil
		},
		Order: []string{"xo"},
	})
}

// Context keys.
const (
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

// Files are the embedded dot templates.
//
//go:embed *.tpl
var Files embed.FS
