package internal

import (
	"strconv"
	"strings"
	"text/template"

	"github.com/serenize/snaker"

	"github.com/knq/xo/models"
)

// NewTemplateFuncs returns a set of template funcs bound to the supplied args.
func (a *ArgType) NewTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"colcount":       a.colcount,
		"colnames":       a.colnames,
		"colnamesquery":  a.colnamesquery,
		"colprefixnames": a.colprefixnames,
		"colvals":        a.colvals,
		"fieldnames":     a.fieldnames,
		"goparamlist":    a.goparamlist,
		"reniltype":      a.reniltype,
		"retype":         a.retype,
		"shortname":      a.shortname,
		"convext":        a.convext,
		"schema":         a.schemafn,
		"colname":        a.colname,
	}
}

// retype checks typ against known types, and prefixing
// ArgType.CustomTypePackage (if applicable).
func (a *ArgType) retype(typ string) string {
	if strings.Contains(typ, ".") {
		return typ
	}

	prefix := ""
	for strings.HasPrefix(typ, "[]") {
		typ = typ[2:]
		prefix = prefix + "[]"
	}

	if _, ok := a.KnownTypeMap[typ]; !ok {
		pkg := a.CustomTypePackage
		if pkg != "" {
			pkg = pkg + "."
		}

		return prefix + pkg + typ
	}

	return prefix + typ
}

// reniltype checks typ against known nil types (similar to retype), prefixing
// ArgType.CustomTypePackage (if applicable).
func (a *ArgType) reniltype(typ string) string {
	if strings.Contains(typ, ".") {
		return typ
	}

	if strings.HasSuffix(typ, "{}") {
		if _, ok := a.KnownTypeMap[typ[:len(typ)-2]]; ok {
			return typ
		}

		pkg := a.CustomTypePackage
		if pkg != "" {
			pkg = pkg + "."
		}

		return pkg + typ
	}

	return typ
}

// shortname generates a safe Go identifier for typ. typ is first checked
// against ArgType.ShortNameTypeMap, and if not found, then the value is
// calculated and stored in the ShortNameTypeMap for future use.
//
// A shortname is the concatentation of the lowercase of the first character in
// the words comprising the name. For example, "MyCustomName" will have have
// the shortname of "mcn".
//
// If a generated shortname conflicts with a Go reserved name, then the
// corresponding value in goReservedNames map will be used.
//
// Generated shortnames that have conflicts with any scopeConflicts member will
// have ArgType.NameConflictSuffix appended.
//
// Note: recognized types for scopeConflicts are string, []*Field,
// []*QueryParam.
func (a *ArgType) shortname(typ string, scopeConflicts ...interface{}) string {
	var v string
	var ok bool

	// check short name map
	if v, ok = a.ShortNameTypeMap[typ]; !ok {
		// calc the short name
		u := []string{}
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
		a.ShortNameTypeMap[typ] = v
	}

	// initial conflicts are the default imported packages from
	// xo_package.go.tpl
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
		switch k := c.(type) {
		case string:
			conflicts[k] = true

		case []*Field:
			for _, f := range k {
				conflicts[f.Name] = true
			}
		case []*QueryParam:
			for _, f := range k {
				conflicts[f.Name] = true
			}

		default:
			panic("not implemented")
		}
	}

	// append suffix if conflict exists
	if _, ok := conflicts[v]; ok {
		v = v + a.NameConflictSuffix
	}

	return v
}

// colnames creates a list of the column names found in fields, excluding any
// Field with Name contained in ignoreNames.
//
// Used to present a comma separated list of column names, that can be used in
// a SELECT, or UPDATE, or other SQL clause requiring an list of identifiers
// (ie, "field_1, field_2, field_3, ...").
func (a *ArgType) colnames(fields []*Field, ignoreNames ...string) string {
	ignore := map[string]bool{}
	for _, n := range ignoreNames {
		ignore[n] = true
	}

	str := ""
	i := 0
	for _, f := range fields {
		if ignore[f.Name] {
			continue
		}

		if i != 0 {
			str = str + ", "
		}
		str = str + a.colname(f.Col)
		i++
	}

	return str
}

// colnamesquery creates a list of the column names in fields as a query and
// joined by sep, excluding any Field with Name contained in ignoreNames.
//
// Used to create a list of column names in a WHERE clause (ie, "field_1 = $1
// AND field_2 = $2 AND ...") or in an UPDATE clause (ie, "field = $1, field =
// $2, ...").
func (a *ArgType) colnamesquery(fields []*Field, sep string, ignoreNames ...string) string {
	ignore := map[string]bool{}
	for _, n := range ignoreNames {
		ignore[n] = true
	}

	str := ""
	i := 0
	for _, f := range fields {
		if ignore[f.Name] {
			continue
		}

		if i != 0 {
			str = str + sep
		}
		str = str + a.colname(f.Col) + " = " + a.Loader.NthParam(i)
		i++
	}

	return str
}

// colprefixnames creates a list of the column names found in fields with the
// supplied prefix, excluding any Field with Name contained in ignoreNames.
//
// Used to present a comma separated list of column names with a prefix. Used in
// a SELECT, or UPDATE (ie, "t.field_1, t.field_2, t.field_3, ...").
func (a *ArgType) colprefixnames(fields []*Field, prefix string, ignoreNames ...string) string {
	ignore := map[string]bool{}
	for _, n := range ignoreNames {
		ignore[n] = true
	}

	str := ""
	i := 0
	for _, f := range fields {
		if ignore[f.Name] {
			continue
		}

		if i != 0 {
			str = str + ", "
		}
		str = str + prefix + "." + a.colname(f.Col)
		i++
	}

	return str
}

// colvals creates a list of value place holders for fields excluding any Field
// with Name contained in ignoreNames.
//
// Used to present a comma separated list of column place holders, used in a
// SELECT or UPDATE statement (ie, "$1, $2, $3 ...").
func (a *ArgType) colvals(fields []*Field, ignoreNames ...string) string {
	ignore := map[string]bool{}
	for _, n := range ignoreNames {
		ignore[n] = true
	}

	str := ""
	i := 0
	for _, f := range fields {
		if ignore[f.Name] {
			continue
		}

		if i != 0 {
			str = str + ", "
		}
		str = str + a.Loader.NthParam(i)
		i++
	}

	return str
}

// fieldnames creates a list of field names from fields of the adding the
// provided prefix, and excluding any Field with Name contained in ignoreNames.
//
// Used to present a comma separated list of field names, ie in a Go statement
// (ie, "t.Field1, t.Field2, t.Field3 ...")
func (a *ArgType) fieldnames(fields []*Field, prefix string, ignoreNames ...string) string {
	ignore := map[string]bool{}
	for _, n := range ignoreNames {
		ignore[n] = true
	}

	str := ""
	i := 0
	for _, f := range fields {
		if ignore[f.Name] {
			continue
		}

		if i != 0 {
			str = str + ", "
		}
		str = str + prefix + "." + f.Name
		i++
	}

	return str
}

// colcount returns the 1-based count of fields, excluding any Field with Name
// contained in ignoreNames.
//
// Used to get the count of fields, and useful for specifying the last SQL
// parameter.
func (a *ArgType) colcount(fields []*Field, ignoreNames ...string) int {
	ignore := map[string]bool{}
	for _, n := range ignoreNames {
		ignore[n] = true
	}

	i := 1
	for _, f := range fields {
		if ignore[f.Name] {
			continue
		}

		i++
	}
	return i
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
	"uintptr":    "up",
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

// goparamlist converts a list of fields into their named Go parameters,
// skipping any Field with Name contained in ignoreNames.
func (a *ArgType) goparamlist(fields []*Field, addType bool, ignoreNames ...string) string {
	ignore := map[string]bool{}
	for _, n := range ignoreNames {
		ignore[n] = true
	}

	str := ""
	i := 0
	for _, f := range fields {
		if ignore[f.Name] {
			continue
		}

		s := "v" + strconv.Itoa(i)
		if len(f.Name) > 0 {
			n := strings.Split(snaker.CamelToSnake(f.Name), "_")
			s = strings.ToLower(n[0]) + f.Name[len(n[0]):]
		}

		// check go reserved names
		if r, ok := goReservedNames[strings.ToLower(s)]; ok {
			s = r
		}

		str = str + ", " + s
		if addType {
			str = str + " " + a.retype(f.Type)
		}

		i++
	}

	return str
}

// convext generates the Go conversion for f in order for it to be assignable
// to t.
//
// FIXME: this should be a better name, like "goconversion" or some such.
func (a *ArgType) convext(prefix string, f *Field, t *Field) string {
	expr := prefix + "." + f.Name
	if f.Type == t.Type {
		return expr
	}

	ft := f.Type
	if strings.HasPrefix(ft, "sql.Null") {
		expr = expr + "." + f.Type[8:]
		ft = strings.ToLower(f.Type[8:])
	}

	if t.Type != ft {
		expr = t.Type + "(" + expr + ")"
	}

	return expr
}

// schemafn takes a series of names and joins them with the schema name.
func (a *ArgType) schemafn(s string, names ...string) string {
	// escape table names
	if a.EscapeTableNames {
		for i, t := range names {
			names[i] = a.Loader.Escape(TableEsc, t)
		}
	}

	n := strings.Join(names, ".")

	if s == "" && n == "" {
		return ""
	}

	if s != "" && n != "" {
		if a.EscapeSchemaName {
			s = a.Loader.Escape(SchemaEsc, s)
		}
		s = s + "."
	}

	return s + n
}

// colname returns the ColumnName of col, optionally escaping it if
// ArgType.EscapeColumnNames is toggled.
func (a *ArgType) colname(col *models.Column) string {
	if a.EscapeColumnNames {
		return a.Loader.Escape(ColumnEsc, col.ColumnName)
	}

	return col.ColumnName
}
