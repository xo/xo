// Package gotpl provides a Go template for xo.
package gotpl

import (
	"context"
	"embed"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gedex/inflector"
	"github.com/kenshaw/snaker"
	"github.com/xo/xo/loader"
	"github.com/xo/xo/templates"
	xo "github.com/xo/xo/types"
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
		Flags: []xo.Flag{
			{
				ContextKey: NotFirstKey,
				Desc:       "disable package comment (ie, not first generated file)",
				Short:      '2',
				Default:    "false",
				Value:      false,
			},
			{
				ContextKey:  xo.Int32Key,
				Desc:        "int32 type (default: int)",
				PlaceHolder: "int",
				Default:     "int",
				Value:       "",
			},
			{
				ContextKey:  xo.Uint32Key,
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
				Default:     "`json:\"{{ .SQLName }}\"`",
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
		Process: func(ctx context.Context, doAppend bool, set *templates.TemplateSet, v *xo.XO) error {
			for _, q := range v.Queries {
				if err := emitQuery(ctx, doAppend, set, q); err != nil {
					return err
				}
			}
			for _, s := range v.Schemas {
				if err := emitSchema(ctx, set, s); err != nil {
					return err
				}
			}
			return nil
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
		Order: []string{"enum", "typedef", "custom", "index", "foreignkey", "proc"},
	})
}

func emitQuery(ctx context.Context, doAppend bool, set *templates.TemplateSet, query xo.Query) error {
	var table Table
	// build type if needed
	if !query.Exec {
		var err error
		table, err = buildQueryType(ctx, query)
		if err != nil {
			return err
		}
	}
	// emit type definition
	if !query.Exec && !query.Flat && !doAppend {
		if err := set.Emit(ctx, &templates.Template{
			Set:      "query",
			Template: "typedef",
			Type:     query.Type,
			Data:     table,
		}); err != nil {
			return err
		}
	}
	// build query params
	var params []QueryParam
	for _, z := range query.Params {
		params = append(params, QueryParam{
			Name:        z.Name,
			Type:        z.Datatype.Type,
			Interpolate: z.Interpolate,
			Join:        z.Join,
		})
	}
	// emit query
	return set.Emit(ctx, &templates.Template{
		Set:      "query",
		Template: "custom",
		Type:     query.Type,
		Data: Query{
			Name:        buildQueryName(query),
			Query:       query.Query,
			Comments:    query.Comments,
			Params:      params,
			One:         query.Exec || query.Flat || query.One,
			Flat:        query.Flat,
			Exec:        query.Exec,
			Interpolate: query.Interpolate,
			Type:        table,
			Comment:     query.Comment,
		},
	})
}

func buildQueryType(ctx context.Context, query xo.Query) (Table, error) {
	var fields []Field
	for _, z := range query.Fields {
		f, err := convertField(ctx, true, z)
		if err != nil {
			return Table{}, err
		}
		// Provided by user already
		if query.ManualFields {
			f = Field{
				GoName:  z.Name,
				SQLName: snaker.CamelToSnake(z.Name),
				Type:    z.Datatype.Type,
			}
		}
		if query.Flat {
			f.GoName = snaker.ForceLowerCamelIdentifier(f.GoName)
		}
		fields = append(fields, f)
	}
	sqlName := strings.ToLower(snaker.CamelToSnake(query.Type))
	return Table{
		GoName:  query.Type,
		SQLName: sqlName,
		Fields:  fields,
		Comment: query.TypeComment,
	}, nil
}

func buildQueryName(query xo.Query) string {
	name := query.Name
	if name == "" {
		// no func name specified, so generate based on type
		if query.One {
			name = query.Type
		} else {
			name = inflector.Pluralize(query.Type)
		}
		// affix any params
		if len(query.Params) == 0 {
			name = "Get" + name
		} else {
			name += "By"
			for _, p := range query.Params {
				name += snaker.ForceCamelIdentifier(p.Name)
			}
		}
	}
	return name
}

func emitSchema(ctx context.Context, set *templates.TemplateSet, s xo.Schema) error {
	// emit enums
	for _, e := range s.Enums {
		enum := convertEnum(e)
		if err := set.Emit(ctx, &templates.Template{
			Set:      "schema",
			Template: "enum",
			Type:     enum.GoName,
			Data:     enum,
		}); err != nil {
			return err
		}
	}
	// emit procs
	for _, p := range s.Procs {
		proc, err := convertProc(ctx, p)
		if err != nil {
			return err
		}
		prefix := "sp_"
		if proc.Kind == "function" {
			prefix = "sf_"
		}
		if err := set.Emit(ctx, &templates.Template{
			Set:      "schema",
			Template: "proc",
			Type:     prefix + proc.GoName,
			Data:     proc,
		}); err != nil {
			return err
		}
	}
	// emit tables
	for _, t := range append(s.Tables, s.Views...) {
		table, err := convertTable(ctx, t)
		if err != nil {
			return err
		}
		if err := set.Emit(ctx, &templates.Template{
			Set:      "schema",
			Template: "typedef",
			Type:     table.GoName,
			Data:     table,
		}); err != nil {
			return err
		}
		// emit indexes
		for _, i := range t.Indexes {
			index, err := convertIndex(ctx, table, i)
			if err != nil {
				return err
			}
			if err := set.Emit(ctx, &templates.Template{
				Set:      "schema",
				Template: "index",
				Type:     table.GoName,
				Name:     index.SQLName,
				Data:     index,
			}); err != nil {
				return err
			}
		}
		// emit fkeys
		for _, fk := range t.ForeignKeys {
			fkey, err := convertFKey(ctx, table, fk)
			if err != nil {
				return err
			}
			if err := set.Emit(ctx, &templates.Template{
				Set:      "schema",
				Template: "foreignkey",
				Type:     table.GoName,
				Name:     fkey.SQLName,
				Data:     fkey,
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

func convertEnum(e xo.Enum) Enum {
	var vals []EnumValue
	for _, v := range e.Values {
		name := snaker.ForceCamelIdentifier(strings.ToLower(v.Name))
		if strings.HasSuffix(name, strings.ToLower(e.Name)) {
			n := name[:len(name)-len(e.Name)]
			if len(n) > 0 {
				name = n
			}
		}
		vals = append(vals, EnumValue{
			GoName:     name,
			SQLName:    v.Name,
			ConstValue: *v.ConstValue,
		})
	}
	return Enum{
		GoName:  snaker.ForceCamelIdentifier((e.Name)),
		SQLName: e.Name,
		Values:  vals,
	}
}

func convertProc(ctx context.Context, p xo.Proc) (Proc, error) {
	var paramList, retList []string
	var params []Field
	var returns []Field
	// proc params
	for _, z := range p.Params {
		f, err := convertField(ctx, false, z)
		if err != nil {
			return Proc{}, err
		}
		paramList = append(paramList, z.Datatype.Type)
		params = append(params, f)
	}
	// proc return
	for _, z := range p.Returns {
		f, err := convertField(ctx, false, z)
		if err != nil {
			return Proc{}, err
		}
		retList = append(retList, z.Datatype.Type)
		returns = append(returns, f)
	}
	// signature used for comment
	var format string
	switch {
	case p.Void:
		format = "%s.%s(%s)%s"
	case len(p.Returns) == 1:
		format = "%s.%s(%s) %s"
	default:
		format = "%s.%s(%s) (%s)"
	}
	signature := fmt.Sprintf(format,
		templates.Schema(ctx), p.Name,
		strings.Join(paramList, ", "),
		strings.Join(retList, ", "))
	return Proc{
		GoName:    snaker.SnakeToCamelIdentifier(p.Name),
		SQLName:   p.Name,
		Kind:      p.Kind,
		Signature: signature,
		Params:    params,
		Returns:   returns,
		Void:      p.Void,
		Comment:   "",
	}, nil
}

func convertTable(ctx context.Context, t xo.Table) (Table, error) {
	var cols, pkCols []Field
	for _, z := range t.Columns {
		f, err := convertField(ctx, true, z)
		if err != nil {
			return Table{}, err
		}
		cols = append(cols, f)
		if z.IsPrimary {
			pkCols = append(pkCols, f)
		}
	}
	name := snaker.ForceCamelIdentifier(singularize(t.Name))
	return Table{
		GoName:      name,
		SQLName:     t.Name,
		Kind:        t.Type,
		Fields:      cols,
		PrimaryKeys: pkCols,
		Manual:      t.Manual,
	}, nil
}

func convertIndex(ctx context.Context, t Table, i xo.Index) (Index, error) {
	var fields []Field
	for _, z := range i.Fields {
		f, err := convertField(ctx, true, z)
		if err != nil {
			return Index{}, err
		}
		fields = append(fields, f)
	}
	funcName := snaker.ForceCamelIdentifier(i.FuncName)
	return Index{
		SQLName:   i.Name,
		FuncName:  funcName,
		Table:     t,
		Fields:    fields,
		IsUnique:  i.IsUnique,
		IsPrimary: i.IsPrimary,
	}, nil
}

func convertFKey(ctx context.Context, t Table, fk xo.ForeignKey) (ForeignKey, error) {
	name := snaker.SnakeToCamelIdentifier(fk.ResolvedName)
	refName := snaker.ForceCamelIdentifier(singularize(fk.RefTable))
	refFunc := snaker.ForceCamelIdentifier(fk.RefFuncName)
	var fields, refFields []Field
	// convert fields
	for _, f := range fk.Fields {
		field, err := convertField(ctx, true, f)
		if err != nil {
			return ForeignKey{}, err
		}
		fields = append(fields, field)
	}
	// convert ref fields
	for _, f := range fk.RefFields {
		refField, err := convertField(ctx, true, f)
		if err != nil {
			return ForeignKey{}, err
		}
		refFields = append(refFields, refField)
	}
	return ForeignKey{
		GoName:      name,
		SQLName:     fk.Name,
		Table:       t,
		Fields:      fields,
		RefTable:    refName,
		RefFields:   refFields,
		RefFuncName: refFunc,
	}, nil
}

func convertField(ctx context.Context, identifier bool, f xo.Field) (Field, error) {
	l := Loader(ctx)
	name := snaker.ForceCamelIdentifier(f.Name)
	if !identifier {
		name = snaker.ForceLowerCamelIdentifier(f.Name)
	}
	typ, zero, err := l.GoType(ctx, f.Datatype)
	if err != nil {
		return Field{}, err
	}
	return Field{
		GoName:     name,
		SQLName:    f.Name,
		Type:       typ,
		Zero:       zero,
		IsPrimary:  f.IsPrimary,
		IsSequence: f.IsSequence,
	}, nil
}

// Context keys.
const (
	NotFirstKey   xo.ContextKey = "not-first"
	PkgKey        xo.ContextKey = "pkg"
	TagKey        xo.ContextKey = "tag"
	ImportKey     xo.ContextKey = "import"
	UUIDKey       xo.ContextKey = "uuid"
	CustomKey     xo.ContextKey = "custom"
	ConflictKey   xo.ContextKey = "conflict"
	EscKey        xo.ContextKey = "esc"
	FieldTagKey   xo.ContextKey = "field-tag"
	ContextKey    xo.ContextKey = "context"
	InjectKey     xo.ContextKey = "inject"
	InjectFileKey xo.ContextKey = "inject-file"
)

// Loader returns the loader from the context.
func Loader(ctx context.Context) *loader.Loader {
	l, _ := ctx.Value(xo.LoaderKey).(*loader.Loader)
	return l
}

// NotFirst returns not-first from the context.
func NotFirst(ctx context.Context) bool {
	b, _ := ctx.Value(NotFirstKey).(bool)
	return b
}

// Pkg returns pkg from the context.
func Pkg(ctx context.Context) string {
	s, _ := ctx.Value(PkgKey).(string)
	if s == "" {
		s = filepath.Base(templates.Out(ctx))
	}
	return s
}

// Tags returns tags from the context.
func Tags(ctx context.Context) []string {
	v, _ := ctx.Value(TagKey).([]string)
	// build tags
	var tags []string
	for _, tag := range v {
		if tag != "" {
			tags = append(tags, tag)
		}
	}
	return tags
}

// Imports returns package imports from the context.
func Imports(ctx context.Context) []string {
	v, _ := ctx.Value(ImportKey).([]string)
	// build imports
	var imports []string
	for _, s := range v {
		if s != "" {
			imports = append(imports, s)
		}
	}
	// add uuid import
	if s, _ := ctx.Value(UUIDKey).(string); s != "" {
		imports = append(imports, s)
	}
	return imports
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

// Esc indicates if esc should be escaped based from the context.
func Esc(ctx context.Context, esc string) bool {
	v, _ := ctx.Value(EscKey).([]string)
	return !contains(v, "none") && (contains(v, "all") || contains(v, esc))
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

// singuralize will singularize a identifier, returning in CamelCase.
func singularize(s string) string {
	if i := lastIndex(s, '_'); i != -1 {
		return s[:i] + "_" + inflector.Singularize(s[i+1:])
	}
	return inflector.Singularize(s)
}

// lastIndex finds the last rune r in s, returning -1 if not present.
func lastIndex(s string, c rune) int {
	r := []rune(s)
	for i := len(r) - 1; i >= 0; i-- {
		if r[i] == c {
			return i
		}
	}
	return -1
}

// Files are the embedded Go templates.
//
//go:embed *.tpl
//go:embed */*.tpl
var Files embed.FS
