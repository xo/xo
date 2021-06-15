// Package gotpl provides a Go template for xo.
package gotpl

import (
	"context"
	"embed"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/xo/xo/templates"
	"golang.org/x/tools/imports"
	"mvdan.cc/gofumpt/format"
)

func init() {
	first := true
	knownTypes := map[string]bool{
		"bool":        true,
		"string":      true,
		"byte":        true,
		"rune":        true,
		"int":         true,
		"int16":       true,
		"int32":       true,
		"int64":       true,
		"uint":        true,
		"uint8":       true,
		"uint16":      true,
		"uint32":      true,
		"uint64":      true,
		"float32":     true,
		"float64":     true,
		"Slice":       true,
		"StringSlice": true,
	}
	shortNames := map[string]string{
		"bool":        "b",
		"string":      "s",
		"byte":        "b",
		"rune":        "r",
		"int":         "i",
		"int16":       "i",
		"int32":       "i",
		"int64":       "i",
		"uint":        "u",
		"uint8":       "u",
		"uint16":      "u",
		"uint32":      "u",
		"uint64":      "u",
		"float32":     "f",
		"float64":     "f",
		"Slice":       "s",
		"StringSlice": "ss",
	}
	templates.Register("go", &templates.TemplateSet{
		Files:   Files,
		FileExt: ".xo.go",
		AddType: func(typ string) {
			knownTypes[typ] = true
		},
		Flags: []templates.Flag{
			{
				ContextKey: NotFirstKey,
				Desc:       "disable package comment (ie, not first generated file)",
				Short:      '2',
				Default:    "false",
				Value:      false,
			},
			{
				ContextKey:  Int32Key,
				Desc:        "int32 type (default: int)",
				PlaceHolder: "int",
				Default:     "int",
				Value:       "",
			},
			{
				ContextKey:  Uint32Key,
				Desc:        "uint32 type (default: uint)",
				PlaceHolder: "uint",
				Default:     "uint",
				Value:       "",
			},
			{
				ContextKey:  PkgKey,
				Desc:        "package name",
				PlaceHolder: "<name>",
				Value:       "",
			},
			{
				ContextKey:  TagKey,
				Desc:        "build tags",
				PlaceHolder: `""`,
				Value:       []string{},
			},
			{
				ContextKey:  ImportKey,
				Desc:        "package imports",
				PlaceHolder: `""`,
				Value:       []string{},
			},
			{
				ContextKey:  UUIDKey,
				Desc:        "uuid type package",
				PlaceHolder: "<pkg>",
				Default:     "github.com/google/uuid",
				Value:       "",
			},
			{
				ContextKey:  CustomKey,
				Desc:        "package name for custom types",
				PlaceHolder: "<name>",
				Value:       "",
			},
			{
				ContextKey:  ConflictKey,
				Desc:        "name conflict suffix (default: Val)",
				PlaceHolder: "Val",
				Default:     "Val",
				Value:       "",
			},
			{
				ContextKey:  EscKey,
				Desc:        "escape fields (none, schema, table, column, all; default: none)",
				PlaceHolder: "none",
				Default:     "none",
				Value:       []string{},
				Enums:       []string{"none", "schema", "table", "column", "all"},
			},
			{
				ContextKey:  FieldTagKey,
				Desc:        "field tag",
				PlaceHolder: `<tag>`,
				Short:       'g',
				Default:     "`json:\"{{ .Col.ColumnName }}\"`",
				Value:       "",
			},
			{
				ContextKey:  ContextKey,
				Desc:        "context mode (disable, both, only; default: only)",
				PlaceHolder: "only",
				Default:     "only",
				Value:       "",
				Enums:       []string{"disable", "both", "only"},
			},
			{
				ContextKey:  InjectKey,
				Desc:        "insert code into generated file headers",
				PlaceHolder: `""`,
				Default:     "",
				Value:       "",
			},
			{
				ContextKey:  InjectFileKey,
				Desc:        "insert code into generated file headers from a file",
				PlaceHolder: `<file>`,
				Default:     "",
				Value:       "",
			},
		},
		Funcs: func(ctx context.Context) (template.FuncMap, error) {
			f, err := NewFuncs(ctx, knownTypes, shortNames, &first)
			if err != nil {
				return nil, err
			}
			return f.FuncMap(), nil
		},
		HeaderTemplate: func(ctx context.Context) *templates.Template {
			return &templates.Template{
				Template: "hdr",
			}
		},
		PackageTemplates: func(ctx context.Context) []*templates.Template {
			if NotFirst(ctx) {
				return nil
			}
			return []*templates.Template{
				{
					Template: "db",
					Name:     "db",
				},
			}
		},
		FileName: func(ctx context.Context, tpl *templates.Template) string {
			if templates.GenType(ctx) == "schema" {
				switch tpl.Template {
				case "typedef", "enum", "index", "foreignkey", "proc":
					return strings.ToLower(tpl.Type)
				}
			}
			file := tpl.Name
			if file == "" {
				file = tpl.Type
			}
			return strings.ToLower(file)
		},
		Post: func(ctx context.Context, buf []byte) ([]byte, error) {
			// imports processing
			buf, err := imports.Process("", buf, nil)
			if err != nil {
				return nil, err
			}
			// format
			return format.Source(buf, format.Options{
				ExtraRules: true,
			})
		},
	})
}

// Context keys.
const (
	NotFirstKey   templates.ContextKey = "not-first"
	Int32Key      templates.ContextKey = "int32"
	Uint32Key     templates.ContextKey = "uint32"
	PkgKey        templates.ContextKey = "pkg"
	TagKey        templates.ContextKey = "tag"
	ImportKey     templates.ContextKey = "import"
	UUIDKey       templates.ContextKey = "uuid"
	CustomKey     templates.ContextKey = "custom"
	ConflictKey   templates.ContextKey = "conflict"
	EscKey        templates.ContextKey = "esc"
	FieldTagKey   templates.ContextKey = "field-tag"
	ContextKey    templates.ContextKey = "context"
	InjectKey     templates.ContextKey = "inject"
	InjectFileKey templates.ContextKey = "inject-file"
)

// NotFirst returns not-first from the context.
func NotFirst(ctx context.Context) bool {
	b, _ := ctx.Value(NotFirstKey).(bool)
	return b
}

// Int32 returns int32 from the context.
func Int32(ctx context.Context) string {
	s, _ := ctx.Value(Int32Key).(string)
	return s
}

// Uint32 returns uint32 from the context.
func Uint32(ctx context.Context) string {
	s, _ := ctx.Value(Uint32Key).(string)
	return s
}

// Pkg returns pkg from the context.
func Pkg(ctx context.Context) string {
	s, _ := ctx.Value(PkgKey).(string)
	if s == "" {
		s = filepath.Base(templates.Out(ctx))
	}
	return strings.ToLower(s)
}

// Tag returns tag from the context.
func Tag(ctx context.Context) []string {
	v, _ := ctx.Value(TagKey).([]string)
	return v
}

// Import returns import from the context.
func Import(ctx context.Context) []string {
	v, _ := ctx.Value(ImportKey).([]string)
	return v
}

// UUID returns uuid from the context.
func UUID(ctx context.Context) string {
	s, _ := ctx.Value(UUIDKey).(string)
	return s
}

// Custom returns custom-pkg from the context.
func Custom(ctx context.Context) string {
	s, _ := ctx.Value(CustomKey).(string)
	return s
}

// Conflict returns conflict from the context.
func Conflict(ctx context.Context) string {
	s, _ := ctx.Value(ConflictKey).(string)
	return s
}

// Esc returns esc from the context.
func Esc(ctx context.Context) []string {
	v, _ := ctx.Value(EscKey).([]string)
	return v
}

// FieldTag returns field-tag from the context.
func FieldTag(ctx context.Context) string {
	s, _ := ctx.Value(FieldTagKey).(string)
	return s
}

// Context returns context from the context.
func Context(ctx context.Context) string {
	s, _ := ctx.Value(ContextKey).(string)
	return s
}

// Inject returns inject from the context.
func Inject(ctx context.Context) string {
	s, _ := ctx.Value(InjectKey).(string)
	return s
}

// InjectFile returns inject-file from the context.
func InjectFile(ctx context.Context) string {
	s, _ := ctx.Value(InjectFileKey).(string)
	return s
}

// contains returns true when s is in v.
func contains(v []string, s string) bool {
	for _, z := range v {
		if z == s {
			return true
		}
	}
	return false
}

// Files are the embedded Go templates.
//
//go:embed *.tpl
//go:embed */*.tpl
var Files embed.FS
