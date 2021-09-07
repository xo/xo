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
	shorts := map[string]string{
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
				Type:       "bool",
				Desc:       "disable package comment (ie, not first generated file)",
				Short:      '2',
				Default:    "false",
			},
			{
				ContextKey:  xo.Int32Key,
				Type:        "string",
				Desc:        "int32 type (default: int)",
				PlaceHolder: "int",
				Default:     "int",
			},
			{
				ContextKey:  xo.Uint32Key,
				Type:        "string",
				Desc:        "uint32 type (default: uint)",
				PlaceHolder: "uint",
				Default:     "uint",
			},
			{
				ContextKey:  PkgKey,
				Type:        "string",
				Desc:        "package name",
				PlaceHolder: "<name>",
			},
			{
				ContextKey:  TagKey,
				Type:        "[]string",
				Desc:        "build tags",
				PlaceHolder: `""`,
			},
			{
				ContextKey:  ImportKey,
				Type:        "[]string",
				Desc:        "package imports",
				PlaceHolder: `""`,
			},
			{
				ContextKey:  UUIDKey,
				Type:        "string",
				Desc:        "uuid type package",
				PlaceHolder: "<pkg>",
				Default:     "github.com/google/uuid",
			},
			{
				ContextKey:  CustomKey,
				Type:        "string",
				Desc:        "package name for custom types",
				PlaceHolder: "<name>",
			},
			{
				ContextKey:  ConflictKey,
				Type:        "string",
				Desc:        "name conflict suffix (default: Val)",
				PlaceHolder: "Val",
				Default:     "Val",
			},
			{
				ContextKey:  InitialismKey,
				Type:        "[]string",
				Desc:        "add initialism (i.e ID, API, URI)",
				PlaceHolder: "<val>",
			},
			{
				ContextKey:  EscKey,
				Type:        "[]string",
				Desc:        "escape fields (none, schema, table, column, all; default: none)",
				PlaceHolder: "none",
				Default:     "none",
				Enums:       []string{"none", "schema", "table", "column", "all"},
			},
			{
				ContextKey:  FieldTagKey,
				Type:        "string",
				Desc:        "field tag",
				PlaceHolder: `<tag>`,
				Short:       'g',
				Default:     "`json:\"{{ .SQLName }}\"`",
			},
			{
				ContextKey:  ContextKey,
				Type:        "string",
				Desc:        "context mode (disable, both, only; default: only)",
				PlaceHolder: "only",
				Default:     "only",
				Enums:       []string{"disable", "both", "only"},
			},
			{
				ContextKey:  InjectKey,
				Type:        "string",
				Desc:        "insert code into generated file headers",
				PlaceHolder: `""`,
				Default:     "",
			},
			{
				ContextKey:  InjectFileKey,
				Type:        "string",
				Desc:        "insert code into generated file headers from a file",
				PlaceHolder: `<file>`,
				Default:     "",
			},
			{
				ContextKey: LegacyKey,
				Type:       "bool",
				Desc:       "enables legacy v1 template funcs",
				Default:    "false",
			},
		},
		Funcs: func(ctx context.Context) template.FuncMap {
			funcs := templates.BaseFuncs()
			if Legacy(ctx) {
				addLegacyFuncs(ctx, funcs)
			}
			return funcs
		},
		BuildContext: func(ctx context.Context) context.Context {
			ctx = context.WithValue(ctx, FirstKey, &first)
			ctx = context.WithValue(ctx, KnownTypesKey, knownTypes)
			ctx = context.WithValue(ctx, ShortsKey, shorts)
			return ctx
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
			if err := addInitialisms(ctx); err != nil {
				return err
			}
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
	tf := camelExport
	if query.Flat {
		tf = camel
	}
	var fields []Field
	for _, z := range query.Fields {
		f, err := convertField(ctx, tf, z)
		if err != nil {
			return Table{}, err
		}
		// dont use convertField; the types are already provided by the user
		if query.ManualFields {
			f = Field{
				GoName:  z.Name,
				SQLName: snake(z.Name),
				Type:    z.Datatype.Type,
			}
		}
		fields = append(fields, f)
	}
	sqlName := snake(query.Type)
	return Table{
		GoName:  query.Type,
		SQLName: sqlName,
		Fields:  fields,
		Comment: query.TypeComment,
	}, nil
}

// buildQueryName builds a name for the query.
func buildQueryName(query xo.Query) string {
	if query.Name != "" {
		return query.Name
	}
	// generate name if not specified
	name := query.Type
	if !query.One {
		name = inflector.Pluralize(name)
	}
	// add params
	if len(query.Params) == 0 {
		name = "Get" + name
	} else {
		name += "By"
		for _, p := range query.Params {
			name += camelExport(p.Name)
		}
	}
	return name
}

// emitSchema emits the xo schema for the template set.
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
	// build procs
	overloadMap := make(map[string][]Proc)
	// procOrder ensures procs are always emitted in alphabetic order for
	// consistency in single mode
	var procOrder []string
	for _, p := range s.Procs {
		var err error
		if procOrder, err = convertProc(ctx, overloadMap, procOrder, p); err != nil {
			return err
		}
	}
	// emit procs
	for _, name := range procOrder {
		procs := overloadMap[name]
		prefix := "sp_"
		if procs[0].Type == "function" {
			prefix = "sf_"
		}
		// change GoName to their overloaded versions if needed
		for i := range procs {
			procs[i].Overloaded = len(procs) > 1
		}
		if err := set.Emit(ctx, &templates.Template{
			Set:      "schema",
			Template: "proc",
			Type:     prefix + name,
			Data:     procs,
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
	goName := camelExport(e.Name)
	for _, v := range e.Values {
		name := camelExport(strings.ToLower(v.Name))
		if strings.HasSuffix(name, goName) && goName != name {
			name = strings.TrimSuffix(name, goName)
		}
		vals = append(vals, EnumValue{
			GoName:     name,
			SQLName:    v.Name,
			ConstValue: *v.ConstValue,
		})
	}
	return Enum{
		GoName:  goName,
		SQLName: e.Name,
		Values:  vals,
	}
}

// convertProc converts a xo.Proc.
func convertProc(ctx context.Context, overloadMap map[string][]Proc, order []string, p xo.Proc) ([]string, error) {
	_, schema, _ := xo.DriverSchemaNthParam(ctx)
	proc := Proc{
		Type:      p.Type,
		GoName:    camelExport(p.Name),
		SQLName:   p.Name,
		Signature: fmt.Sprintf("%s.%s", schema, p.Name),
		Void:      p.Void,
	}
	// proc params
	var types []string
	for _, z := range p.Params {
		f, err := convertField(ctx, camel, z)
		if err != nil {
			return nil, err
		}
		proc.Params = append(proc.Params, f)
		types = append(types, z.Datatype.Type)
	}
	// add to signature, generate name
	proc.Signature += "(" + strings.Join(types, ", ") + ")"
	proc.OverloadedName = overloadedName(types, proc)
	types = nil
	// proc return
	for _, z := range p.Returns {
		f, err := convertField(ctx, camel, z)
		if err != nil {
			return nil, err
		}
		proc.Returns = append(proc.Returns, f)
		types = append(types, z.Datatype.Type)
	}
	// append signature
	if !p.Void {
		format := " (%s)"
		if len(p.Returns) == 1 {
			format = " %s"
		}
		proc.Signature += fmt.Sprintf(format, strings.Join(types, ", "))
	}
	// add proc
	procs, ok := overloadMap[proc.GoName]
	if !ok {
		order = append(order, proc.GoName)
	}
	overloadMap[proc.GoName] = append(procs, proc)
	return order, nil
}

// convertTable converts a xo.Table to a Table.
func convertTable(ctx context.Context, t xo.Table) (Table, error) {
	var cols, pkCols []Field
	for _, z := range t.Columns {
		f, err := convertField(ctx, camelExport, z)
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
		Type:        t.Type,
		GoName:      name,
		SQLName:     t.Name,
		Fields:      cols,
		PrimaryKeys: pkCols,
		Manual:      t.Manual,
	}, nil
}

func convertIndex(ctx context.Context, t Table, i xo.Index) (Index, error) {
	var fields []Field
	for _, z := range i.Fields {
		f, err := convertField(ctx, camelExport, z)
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
	var fields, refFields []Field
	// convert fields
	for _, f := range fk.Fields {
		field, err := convertField(ctx, camelExport, f)
		if err != nil {
			return ForeignKey{}, err
		}
		fields = append(fields, field)
	}
	// convert ref fields
	for _, f := range fk.RefFields {
		refField, err := convertField(ctx, camelExport, f)
		if err != nil {
			return ForeignKey{}, err
		}
		refFields = append(refFields, refField)
	}
	return ForeignKey{
		GoName:      camelExport(fk.FuncName),
		SQLName:     fk.Name,
		Table:       t,
		Fields:      fields,
		RefTable:    camelExport(singularize(fk.RefTable)),
		RefFields:   refFields,
		RefFuncName: camelExport(fk.RefFuncName),
	}, nil
}

func overloadedName(sqlTypes []string, proc Proc) string {
	if len(proc.Params) == 0 {
		return proc.GoName
	}
	var names []string
	// build parameters for proc.
	// if the proc's parameter has no name, use the types of the proc instead
	for i, f := range proc.Params {
		if f.SQLName == fmt.Sprintf("p%d", i) {
			names = append(names, camelExport(strings.Split(sqlTypes[i], " ")...))
			continue
		}
		names = append(names, camelExport(f.GoName))
	}
	if len(names) == 1 {
		return fmt.Sprintf("%sBy%s", proc.GoName, names[0])
	}
	front, last := strings.Join(names[:len(names)-1], ""), names[len(names)-1]
	return fmt.Sprintf("%sBy%sAnd%s", proc.GoName, front, last)
}

func convertField(ctx context.Context, tf transformFunc, f xo.Field) (Field, error) {
	l := Loader(ctx)
	typ, zero, err := l.GoType(ctx, f.Datatype)
	if err != nil {
		return Field{}, err
	}
	return Field{
		Type:       typ,
		GoName:     tf(f.Name),
		SQLName:    f.Name,
		Zero:       zero,
		IsPrimary:  f.IsPrimary,
		IsSequence: f.IsSequence,
	}, nil
}

type transformFunc func(...string) string

func snake(names ...string) string {
	return snaker.CamelToSnake(strings.Join(names, "_"))
}

func camel(names ...string) string {
	return snaker.ForceLowerCamelIdentifier(strings.Join(names, "_"))
}

func camelExport(names ...string) string {
	return snaker.ForceCamelIdentifier(strings.Join(names, "_"))
}

// Context keys.
const (
	FirstKey      xo.ContextKey = "first"
	KnownTypesKey xo.ContextKey = "known-types"
	ShortsKey     xo.ContextKey = "shorts"
	NotFirstKey   xo.ContextKey = "not-first"
	PkgKey        xo.ContextKey = "pkg"
	TagKey        xo.ContextKey = "tag"
	ImportKey     xo.ContextKey = "import"
	UUIDKey       xo.ContextKey = "uuid"
	CustomKey     xo.ContextKey = "custom"
	ConflictKey   xo.ContextKey = "conflict"
	InitialismKey xo.ContextKey = "initialism"
	EscKey        xo.ContextKey = "esc"
	FieldTagKey   xo.ContextKey = "field-tag"
	ContextKey    xo.ContextKey = "context"
	InjectKey     xo.ContextKey = "inject"
	InjectFileKey xo.ContextKey = "inject-file"
	LegacyKey     xo.ContextKey = "legacy"
)

// Loader returns the loader from the context.
func Loader(ctx context.Context) *loader.Loader {
	l, _ := ctx.Value(xo.LoaderKey).(*loader.Loader)
	return l
}

// First returns first from the context.
func First(ctx context.Context) *bool {
	b, _ := ctx.Value(FirstKey).(*bool)
	return b
}

// KnownTYpes returns known-types from the context.
func KnownTypes(ctx context.Context) map[string]bool {
	m, _ := ctx.Value(KnownTypesKey).(map[string]bool)
	return m
}

// Shorts retruns shorts from the context.
func Shorts(ctx context.Context) map[string]string {
	m, _ := ctx.Value(ShortsKey).(map[string]string)
	return m
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

// Legacy returns legacy from the context.
func Legacy(ctx context.Context) bool {
	b, _ := ctx.Value(LegacyKey).(bool)
	return b
}

// addInitialisms adds snaker initialisms from the context.
func addInitialisms(ctx context.Context) error {
	var v []string
	for _, s := range ctx.Value(InitialismKey).([]string) {
		if s != "" {
			v = append(v, s)
		}
	}
	return snaker.DefaultInitialisms.Add(v...)
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
	if i := strings.LastIndex(s, "_"); i != -1 {
		return s[:i+1] + inflector.Singularize(s[i+1:])
	}
	return inflector.Singularize(s)
}

// Files are the embedded Go templates.
//
//go:embed *.tpl
//go:embed */*.tpl
var Files embed.FS
