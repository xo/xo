package createdbtpl

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"text/template"

	"github.com/xo/xo/templates"
	xo "github.com/xo/xo/types"
)

func init() {
	formatterPath, _ := exec.LookPath("sql-formatter")
	var formatterOptions []string
	if formatterPath != "" {
		formatterOptions = []string{"-u", "-l={{ . }}", "-i=2", "--lines-between-queries=2"}
	}
	templates.Register("createdb", &templates.TemplateSet{
		Files:   Files,
		FileExt: ".xo.sql",
		Flags: []xo.Flag{
			{
				ContextKey:  FmtKey,
				Desc:        fmt.Sprintf("fmt command (default: %s)", formatterPath),
				Default:     formatterPath,
				PlaceHolder: "<path>",
				Value:       "",
			},
			{
				ContextKey:  FmtOptsKey,
				Desc:        fmt.Sprintf("fmt options (default: %s)", strings.Join(formatterOptions, ", ")),
				Default:     strings.Join(formatterOptions, ","),
				PlaceHolder: "<opts>",
				Value:       []string{},
			},
			{
				ContextKey: ConstraintKey,
				Desc:       "enable constraint name in output (postgres, mysql, sqlite3)",
				Default:    "false",
				Value:      false,
			},
			{
				ContextKey:  EscKey,
				Desc:        "escape mode (none, types, all; default: none)",
				PlaceHolder: "none",
				Default:     "none",
				Value:       "",
				Enums:       []string{"none", "types", "all"},
			},
			{
				ContextKey:  EngineKey,
				Desc:        "mysql table engine (default: InnoDB)",
				Default:     "InnoDB",
				PlaceHolder: `""`,
				Value:       "",
			},
		},
		Funcs: func(ctx context.Context) (template.FuncMap, error) {
			return NewFuncs(ctx).FuncMap(), nil
		},
		FileName: func(ctx context.Context, tpl *templates.Template) string {
			return tpl.Name
		},
		Post: func(ctx context.Context, buf []byte) ([]byte, error) {
			formatterPath, lang := Fmt(ctx), Lang(ctx)
			if formatterPath == "" {
				return buf, nil
			}
			// build options
			opts := FmtOpts(ctx)
			for i, o := range opts {
				tpl, err := template.New(fmt.Sprintf("option %d", i)).Parse(o)
				if err != nil {
					return nil, err
				}
				b := new(bytes.Buffer)
				if err := tpl.Execute(b, lang); err != nil {
					return nil, err
				}
				opts[i] = b.String()
			}
			// execute
			stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
			cmd := exec.Command(formatterPath, opts...)
			cmd.Stdin, cmd.Stdout, cmd.Stderr = bytes.NewReader(buf), stdout, stderr
			if err := cmd.Run(); err != nil {
				return nil, fmt.Errorf("unable to execute %s: %v: %s", formatterPath, err, stderr.String())
			}
			return append(bytes.TrimSuffix(stdout.Bytes(), []byte{'\n'}), '\n'), nil
		},
		Process: func(ctx context.Context, _ bool, set *templates.TemplateSet, v *xo.XO) error {
			if len(v.Schemas) == 0 {
				return errors.New("createdb template must be passed at least one schema")
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
	FmtKey        xo.ContextKey = "fmt"
	FmtOptsKey    xo.ContextKey = "fmt-opts"
	ConstraintKey xo.ContextKey = "constraint"
	EscKey        xo.ContextKey = "escape"
	EngineKey     xo.ContextKey = "engine"
)

// Fmt returns fmt from the context.
func Fmt(ctx context.Context) string {
	s, _ := ctx.Value(FmtKey).(string)
	return s
}

// FmtOpts returns fmt-opts from the context.
func FmtOpts(ctx context.Context) []string {
	v, _ := ctx.Value(FmtOptsKey).([]string)
	return v
}

// Constraint returns constraint from the context.
func Constraint(ctx context.Context) bool {
	b, _ := ctx.Value(ConstraintKey).(bool)
	return b
}

// Esc returns esc from the context.
func Esc(ctx context.Context, esc string) bool {
	v, _ := ctx.Value(EscKey).(string)
	return v == "all" || v == esc
}

// Engine returns engine from the context.
func Engine(ctx context.Context) string {
	s, _ := ctx.Value(EngineKey).(string)
	return s
}

// Lang returns the sql-formatter language to use from the context based on the
// context driver.
func Lang(ctx context.Context) string {
	driver, _, _ := xo.DriverSchemaNthParam(ctx)
	switch driver {
	case "postgres", "sqlite3":
		return "postgresql"
	case "mysql":
		return "mysql"
	case "sqlserver":
		return "tsql"
	case "oracle":
		return "plsql"
	}
	return "sql"
}

// Files are the embedded SQL templates.
//
//go:embed *.tpl
var Files embed.FS
