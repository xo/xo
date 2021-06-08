package gotpl

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"text/template"

	"github.com/kenshaw/snaker"
	"github.com/xo/xo/models"
	"github.com/xo/xo/templates"
)

// Funcs is a set of template funcs.
type Funcs struct {
	driver     string
	schema     string
	first      *bool
	escSchema  bool
	escTables  bool
	escColumns bool
	nth        func(int) string
	pkg        string
	conflict   string
	custom     string
	// knownTypes is the collection of known Go types.
	knownTypes map[string]bool
	// shortNames is the collection of Go style short names for types, mainly
	// used for use with declaring a func receiver on a type.
	shortNames map[string]string
}

// NewFuncs returns a set of template funcs.
func NewFuncs(ctx context.Context, knownTypes map[string]bool, shortNames map[string]string, first *bool) *Funcs {
	// force not first
	if NotFirst(ctx) {
		b := false
		first = &b
	}
	return &Funcs{
		driver:     templates.Driver(ctx),
		schema:     templates.Schema(ctx),
		first:      first,
		escSchema:  templates.EscSchema(ctx),
		escTables:  templates.EscTables(ctx),
		escColumns: templates.EscColumns(ctx),
		nth:        templates.NthParam(ctx),
		pkg:        Pkg(ctx),
		conflict:   Conflict(ctx),
		custom:     Custom(ctx),
		knownTypes: knownTypes,
		shortNames: shortNames,
	}
}

// AddKnownType adds a known type.
func (f *Funcs) AddKnownType(name string) {
	f.knownTypes[name] = true
}

// FuncMap returns the func map.
func (f *Funcs) FuncMap() template.FuncMap {
	return template.FuncMap{
		// context
		"driver":   f.driverfn,
		"schema":   f.schemafn,
		"first":    f.firstfn,
		"pkg":      f.pkgfn,
		"nthparam": f.nthparam,

		// general
		"colcount":           f.colcount,
		"colname":            f.colname,
		"colnames":           f.colnames,
		"colnamesmulti":      f.colnamesmulti,
		"colnamesquery":      f.colnamesquery,
		"colnamesquerymulti": f.colnamesquerymulti,
		"colprefixnames":     f.colprefixnames,
		"colvals":            f.colvals,
		"colvalsmulti":       f.colvalsmulti,
		"convext":            f.convext,
		"fieldnames":         f.fieldnames,
		"fieldnamesmulti":    f.fieldnamesmulti,
		"fieldtag":           f.fieldtag,
		"startcount":         f.startcount,
		"hascolumn":          f.hascolumn,
		"hasfield":           f.hasfield,
		"paramlist":          f.paramlist,
		"reniltype":          f.reniltype,
		"retype":             f.retype,
		"shortname":          f.shortname,
	}
}

// driverfn returns true if the driver is any of the passed drivers.
func (f *Funcs) driverfn(drivers ...string) bool {
	for _, driver := range drivers {
		if f.driver == driver {
			return true
		}
	}
	return false
}

// schemafn takes a series of names and joins them with the schema name.
func (f *Funcs) schemafn(names ...string) string {
	s := f.schema
	// escape table names
	if f.escTables {
		for i, name := range names {
			names[i] = f.esc(name, "table")
		}
	}
	n := strings.Join(names, ".")
	if s == "" && n == "" {
		return ""
	}
	if s != "" && n != "" {
		if f.escSchema {
			s = f.esc(s, "schema")
		}
		s += "."
	}
	return s + n
}

// first returns true if it is the template was the first template generated.
func (f *Funcs) firstfn() bool {
	b := *f.first
	*f.first = false
	return b
}

// pkgfn returns the package name.
func (f *Funcs) pkgfn() string {
	return f.pkg
}

// nthparam returns the nth param placeholder
func (f *Funcs) nthparam(i int) string {
	return f.nth(i)
}

// retype checks typ against known types, and prefixing custom (if applicable).
func (f *Funcs) retype(typ string) string {
	if strings.Contains(typ, ".") {
		return typ
	}
	prefix := ""
	for strings.HasPrefix(typ, "[]") {
		typ = typ[2:]
		prefix = prefix + "[]"
	}
	if _, ok := f.knownTypes[typ]; !ok {
		pkg := f.custom
		if pkg != "" {
			pkg = pkg + "."
		}
		return prefix + pkg + typ
	}
	return prefix + typ
}

// reniltype checks typ against known nil types (similar to retype), prefixing
// custom (if applicable).
func (f *Funcs) reniltype(typ string) string {
	if strings.Contains(typ, ".") {
		return typ
	}
	if strings.HasSuffix(typ, "{}") {
		if _, ok := f.knownTypes[typ[:len(typ)-2]]; ok {
			return typ
		}
		pkg := f.custom
		if pkg != "" {
			pkg = pkg + "."
		}
		return pkg + typ
	}
	return typ
}

// shortname generates a safe Go identifier for typ. typ is first checked
// against shortNames, and if not found, then the value is calculated and
// stored in the shortNames for future use.
//
// A shortname is the concatentation of the lowercase of the first character in
// the words comprising the name. For example, "MyCustomName" will have have
// the shortname of "mcn".
//
// If a generated shortname conflicts with a Go reserved name, then the
// corresponding value in goReservedNames map will be used.
//
// Generated shortnames that have conflicts with any scopeConflicts member will
// have nameConflictSuffix appended.
//
// Note: recognized types for scopeConflicts are string, []*templates.Field,
// []*templates.QueryParam.
func (f *Funcs) shortname(typ string, scopeConflicts ...interface{}) string {
	// check short name map
	v, ok := f.shortNames[typ]
	if !ok {
		// calc the short name
		var u []string
		for _, s := range strings.Split(strings.ToLower(snaker.CamelToSnake(typ)), "_") {
			if len(s) > 0 && s != "id" {
				u = append(u, s[:1])
			}
		}
		v = strings.Join(u, "")
		// check go reserved names
		if n, ok := goReservedNames[v]; ok {
			v = n
		}
		// store back to short name map
		f.shortNames[typ] = v
	}
	// initial conflicts are the default imported packages
	conflicts := map[string]bool{
		"sql":     true,
		"driver":  true,
		"csv":     true,
		"errors":  true,
		"fmt":     true,
		"regexp":  true,
		"strings": true,
		"time":    true,
	}
	// add scopeConflicts to conflicts
	for _, c := range scopeConflicts {
		switch z := c.(type) {
		case string:
			conflicts[z] = true
		case []*templates.Field:
			for _, field := range z {
				conflicts[field.Name] = true
			}
		case []*templates.QueryParam:
			for _, p := range z {
				conflicts[p.Name] = true
			}
		default:
			panic(fmt.Sprintf("unsupported type %T", c))
		}
	}
	// append suffix if conflict exists
	if _, ok := conflicts[v]; ok {
		v = v + f.conflict
	}
	return v
}

// colnames creates a list of the column names found in fields, excluding any
// Field with Name contained in ignore.
//
// Used to present a comma separated list of column names, that can be used in
// a SELECT, or UPDATE, or other SQL clause requiring an list of identifiers
// (ie, "field_1, field_2, field_3, ...").
func (f *Funcs) colnames(fields []*templates.Field, ignore ...string) string {
	m := map[string]bool{}
	for _, n := range ignore {
		m[n] = true
	}
	s, i := "", 0
	for _, field := range fields {
		if m[field.Name] {
			continue
		}
		if i != 0 {
			s = s + ", "
		}
		s += f.colname(field.Col)
		i++
	}
	return s
}

// colnamesmulti creates a list of the column names found in fields, excluding any
// Field with Name contained in ignore.
//
// Used to present a comma separated list of column names, that can be used in
// a SELECT, or UPDATE, or other SQL clause requiring an list of identifiers
// (ie, "field_1, field_2, field_3, ...").
func (f *Funcs) colnamesmulti(fields []*templates.Field, ignore []*templates.Field) string {
	m := map[string]bool{}
	for _, field := range ignore {
		m[field.Name] = true
	}
	s, i := "", 0
	for _, field := range fields {
		if m[field.Name] {
			continue
		}
		if i != 0 {
			s += ", "
		}
		s += f.colname(field.Col)
		i++
	}
	return s
}

// colnamesquery creates a list of the column names in fields as a query and
// joined by sep, excluding any Field with Name contained in ignore.
//
// Used to create a list of column names in a WHERE clause (ie, "field_1 = $1
// AND field_2 = $2 AND ...") or in an UPDATE clause (ie, "field = $1, field =
// $2, ...").
func (f *Funcs) colnamesquery(fields []*templates.Field, sep string, ignore ...string) string {
	m := map[string]bool{}
	for _, n := range ignore {
		m[n] = true
	}
	s, i := "", 0
	for _, field := range fields {
		if m[field.Name] {
			continue
		}
		if i != 0 {
			s += sep
		}
		s += f.colname(field.Col) + " = " + f.nth(i)
		i++
	}
	return s
}

// colnamesquerymulti creates a list of the column names in fields as a query and
// joined by sep, excluding any Field with Name contained in the slice of fields in ignore.
//
// Used to create a list of column names in a WHERE clause (ie, "field_1 = $1
// AND field_2 = $2 AND ...") or in an UPDATE clause (ie, "field = $1, field =
// $2, ...").
func (f *Funcs) colnamesquerymulti(fields []*templates.Field, sep string, startCount int, ignore []*templates.Field) string {
	m := map[string]bool{}
	for _, field := range ignore {
		m[field.Name] = true
	}
	s, i := "", startCount
	for _, field := range fields {
		if m[field.Name] {
			continue
		}
		if i > startCount {
			s += sep
		}
		s += f.colname(field.Col) + " = " + f.nth(i)
		i++
	}
	return s
}

// colprefixnames creates a list of the column names found in fields with the
// supplied prefix, excluding any Field with Name contained in ignore.
//
// Used to present a comma separated list of column names with a prefix. Used in
// a SELECT, or UPDATE (ie, "t.field_1, t.field_2, t.field_3, ...").
func (f *Funcs) colprefixnames(fields []*templates.Field, prefix string, ignore ...string) string {
	m := map[string]bool{}
	for _, n := range ignore {
		m[n] = true
	}
	s, i := "", 0
	for _, field := range fields {
		if m[field.Name] {
			continue
		}
		if i != 0 {
			s += ", "
		}
		s += prefix + "." + f.colname(field.Col)
		i++
	}
	return s
}

// colvals creates a list of value place holders for fields excluding any Field
// with Name contained in ignore.
//
// Used to present a comma separated list of column place holders, used in a
// SELECT or UPDATE statement (ie, "$1, $2, $3 ...").
func (f *Funcs) colvals(fields []*templates.Field, ignore ...string) string {
	m := map[string]bool{}
	for _, n := range ignore {
		m[n] = true
	}
	s, i := "", 0
	for _, field := range fields {
		if m[field.Name] {
			continue
		}
		if i != 0 {
			s += ", "
		}
		s += f.nth(i)
		i++
	}
	return s
}

// colvalsmulti creates a list of value place holders for fields excluding any Field
// with Name contained in ignore.
//
// Used to present a comma separated list of column place holders, used in a
// SELECT or UPDATE statement (ie, "$1, $2, $3 ...").
func (f *Funcs) colvalsmulti(fields []*templates.Field, ignore []*templates.Field) string {
	m := map[string]bool{}
	for _, field := range ignore {
		m[field.Name] = true
	}
	s, i := "", 0
	for _, field := range fields {
		if m[field.Name] {
			continue
		}
		if i != 0 {
			s += ", "
		}
		s += f.nth(i)
		i++
	}
	return s
}

// fieldnames creates a list of field names from fields of the adding the
// provided prefix, and excluding any Field with Name contained in ignore.
//
// Used to present a comma separated list of field names, ie in a Go statement
// (ie, "t.Field1, t.Field2, t.Field3 ...")
func (f *Funcs) fieldnames(fields []*templates.Field, prefix string, ignore ...string) string {
	m := map[string]bool{}
	for _, n := range ignore {
		m[n] = true
	}
	s, i := "", 0
	for _, field := range fields {
		if m[field.Name] {
			continue
		}
		if i != 0 {
			s += ", "
		}
		if prefix == "" {
			s += "&"
		} else {
			s += prefix + "."
		}
		s += field.Name
		i++
	}
	return s
}

// fieldnamesmulti creates a list of field names from fields of the adding the
// provided prefix, and excluding any Field with the slice contained in ignore.
//
// Used to present a comma separated list of field names, ie in a Go statement
// (ie, "t.Field1, t.Field2, t.Field3 ...")
func (f *Funcs) fieldnamesmulti(fields []*templates.Field, prefix string, ignore []*templates.Field) string {
	m := map[string]bool{}
	for _, field := range ignore {
		m[field.Name] = true
	}
	s, i := "", 0
	for _, field := range fields {
		if m[field.Name] {
			continue
		}
		if i != 0 {
			s += ", "
		}
		if prefix == "" {
			s += "&"
		} else {
			s += prefix + "."
		}
		s += field.Name
		i++
	}
	return s
}

// colcount returns the 1-based count of fields, excluding any Field with Name
// contained in ignore.
//
// Used to get the count of fields, and useful for specifying the last SQL
// parameter.
func (f *Funcs) colcount(fields []*templates.Field, ignore ...string) int {
	m := map[string]bool{}
	for _, n := range ignore {
		m[n] = true
	}
	i := 1
	for _, field := range fields {
		if m[field.Name] {
			continue
		}
		i++
	}
	return i
}

// paramlist converts a list of fields into their named Go parameters, skipping
// any Field with Name contained in ignore. addType will cause the go Type to
// be added after each variable name. addPrefix will cause the returned string
// to be prefixed with ", " if the generated string is not empty.
//
// Any field name encountered will be checked against goReservedNames, and will
// have its name substituted by its corresponding looked up value.
//
// Used to present a comma separated list of Go variable names for use with as
// either a Go func parameter list, or in a call to another Go func.
// (ie, ", a, b, c, ..." or ", a T1, b T2, c T3, ...").
func (f *Funcs) paramlist(fields []*templates.Field, addPrefix, addType bool, ignore ...string) string {
	m := map[string]bool{}
	for _, n := range ignore {
		m[n] = true
	}
	var vals []string
	i := 0
	for _, field := range fields {
		if m[field.Name] {
			continue
		}
		s := "v" + strconv.Itoa(i)
		if len(field.Name) > 0 {
			n := strings.Split(snaker.CamelToSnake(field.Name), "_")
			s = strings.ToLower(n[0]) + field.Name[len(n[0]):]
		}
		// check go reserved names
		if r, ok := goReservedNames[strings.ToLower(s)]; ok {
			s = r
		}
		// add the go type
		if addType {
			s += " " + f.retype(field.Type)
		}
		// add to vals
		vals = append(vals, s)
		i++
	}
	// concat generated values
	s := strings.Join(vals, ", ")
	if addPrefix && s != "" {
		return ", " + s
	}
	return s
}

// convext generates the Go conversion for a to be assignable to b.
func (*Funcs) convext(prefix string, a, b *templates.Field) string {
	expr := prefix + "." + a.Name
	if a.Type == b.Type {
		return expr
	}
	typ := a.Type
	if strings.HasPrefix(typ, "sql.Null") {
		expr = expr + "." + a.Type[8:]
		typ = strings.ToLower(a.Type[8:])
	}
	if b.Type != typ {
		expr = b.Type + "(" + expr + ")"
	}
	return expr
}

// colname returns the ColumnName of col, optionally escaping it if
// escapeColumnNames is toggled.
func (f *Funcs) colname(col *models.Column) string {
	if f.escColumns {
		return f.esc(col.ColumnName, "column")
	}
	return col.ColumnName
}

// hascolumn takes a list of fields and determines if field with the specified
// column name is in the list.
func (f *Funcs) hascolumn(fields []*templates.Field, name string) bool {
	for _, field := range fields {
		if field.Col.ColumnName == name {
			return true
		}
	}
	return false
}

// hasfield takes a list of fields and determines if field with the specified
// field name is in the list.
func (f *Funcs) hasfield(fields []*templates.Field, name string) bool {
	for _, field := range fields {
		if field.Name == name {
			return true
		}
	}
	return false
}

// startcount returns a starting count for numbering columns in queries.
func (f *Funcs) startcount(fields []*templates.Field, pkFields []*templates.Field) int {
	return len(fields) - len(pkFields)
}

// esc escapes s.
func (f *Funcs) esc(s string, typ string) string {
	return `"` + s + `"`
}

// fieldtag returns the field tag for the field.
func (f *Funcs) fieldtag(field *templates.Field) string {
	return fmt.Sprintf("`json:\"%s\"`", field.Col.ColumnName)
}

// goReservedNames is a map of of go reserved names to "safe" names.
var goReservedNames = map[string]string{
	"break":       "brk",
	"case":        "cs",
	"chan":        "chn",
	"const":       "cnst",
	"continue":    "cnt",
	"default":     "def",
	"defer":       "dfr",
	"else":        "els",
	"fallthrough": "flthrough",
	"for":         "fr",
	"func":        "fn",
	"go":          "goVal",
	"goto":        "gt",
	"if":          "ifVal",
	"import":      "imp",
	"interface":   "iface",
	"map":         "mp",
	"package":     "pkg",
	"range":       "rnge",
	"return":      "ret",
	"select":      "slct",
	"struct":      "strct",
	"switch":      "swtch",
	"type":        "typ",
	"var":         "vr",
	// go types
	"error":      "e",
	"bool":       "b",
	"string":     "str",
	"byte":       "byt",
	"rune":       "r",
	"uintptr":    "uptr",
	"int":        "i",
	"int8":       "i8",
	"int16":      "i16",
	"int32":      "i32",
	"int64":      "i64",
	"uint":       "u",
	"uint8":      "u8",
	"uint16":     "u16",
	"uint32":     "u32",
	"uint64":     "u64",
	"float32":    "z",
	"float64":    "f",
	"complex64":  "c",
	"complex128": "c128",
}
