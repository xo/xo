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
		"inc":            a.inc,
		"reniltype":      a.reniltype,
		"retype":         a.retype,
		"shortname":      a.shortname,
		"nulltypeext":    a.nulltypeext,
		"schema":         a.schemafn,
	}
}

// retype checks the type against the known types, adding the custom type
// package (if any).
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

// reniltype checks the nil type against the known types (similar to retype),
// adding the custom type package (if applicable).
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

// shortname checks the passed type against the ShortNameTypeMap and returns
// the value for it. If the type is not found, then the value is calculated and
// stored in the ShortNameTypeMap for use in the future.
func (a *ArgType) shortname(typ string) string {
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

		// store back in short name map
		a.ShortNameTypeMap[typ] = v
	}

	return v
}

// inc increements i by 1.
func (a *ArgType) inc(i int) int {
	return i + 1
}

// colnames creates a list of the column names found in columns, excluding any
// FieldName in ignoreFields.
//
// Used to present a comma separated list of column names, ie in a sql select,
// or update. (ie, field_1, field_2, field_3, ...)
func (a *ArgType) colnames(columns []*models.Column, ignoreFields ...string) string {
	ignore := map[string]bool{}
	for _, n := range ignoreFields {
		ignore[n] = true
	}

	str := ""
	i := 0
	for _, col := range columns {
		if ignore[col.Field] {
			continue
		}

		if i != 0 {
			str = str + ", "
		}
		str = str + col.ColumnName
		i++
	}

	return str
}

// colnamesquery creates a list of the column names found in as a query and
// joined by sep.
//
// Used to create a sql query list of column names in a where clause (ie,
// field_1 = $1 AND field_2 = $2 AND ... ) or in an update clause (ie, field =
// $1, field = $2, ...)
func (a *ArgType) colnamesquery(columns []*models.Column, sep string, ignoreFields ...string) string {
	ignore := map[string]bool{}
	for _, n := range ignoreFields {
		ignore[n] = true
	}

	str := ""
	i := 0
	for _, col := range columns {
		if ignore[col.Field] {
			continue
		}

		if i != 0 {
			str = str + sep
		}
		str = str + col.ColumnName + " = " + a.Loader.NthParam(i)
		i++
	}

	return str
}

// colprefixnames creates a list of the column names found in columns with the
// supplied prefix, excluding any FieldName in ignoreFields.
//
// Used to present a comma separated list of column names, with a prefix, ie in
// a sql select, or update. (ie, t.field_1, t.field_2, t.field_3, ...)
func (a *ArgType) colprefixnames(columns []*models.Column, prefix string, ignoreFields ...string) string {
	ignore := map[string]bool{}
	for _, n := range ignoreFields {
		ignore[n] = true
	}

	str := ""
	i := 0
	for _, col := range columns {
		if ignore[col.Field] {
			continue
		}

		if i != 0 {
			str = str + ", "
		}
		str = str + prefix + "." + col.ColumnName
		i++
	}

	return str
}

// colvals creates a list of value place holders for the columns found in
// columns, excluding any FieldName in ignoreFields.
//
// Used to present a comma separated list of column names, ie in a sql select,
// or update. (ie, $1, $2, $3 ...)
func (a *ArgType) colvals(columns []*models.Column, ignoreFields ...string) string {
	ignore := map[string]bool{}
	for _, n := range ignoreFields {
		ignore[n] = true
	}

	str := ""
	i := 0
	for _, col := range columns {
		if ignore[col.Field] {
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

// fieldnames creates a list of field names from the field names of the
// provided columns, adding the prefix provided, and excluding any field name
// in ignoreFields.
//
// Used to present a comma separated list of field names, ie in a Go statement
// (ie, t.Field1, t.Field2, t.Field3 ...)
func (a *ArgType) fieldnames(columns []*models.Column, prefix string, ignoreFields ...string) string {
	ignore := map[string]bool{}
	for _, n := range ignoreFields {
		ignore[n] = true
	}

	str := ""
	i := 0
	for _, col := range columns {
		if ignore[col.Field] {
			continue
		}

		if i != 0 {
			str = str + ", "
		}
		str = str + prefix + "." + col.Field
		i++
	}

	return str
}

// count returns the 1-based count of columns, excluding any field name in
// ignoreFields.
//
// Used to get the count of columns, and useful for specifying the last sql
// parameter.
func (a *ArgType) colcount(columns []*models.Column, ignoreFields ...string) int {
	ignore := map[string]bool{}
	for _, n := range ignoreFields {
		ignore[n] = true
	}

	i := 1
	for _, col := range columns {
		if ignore[col.Field] {
			continue
		}

		i++
	}
	return i
}

// goparamlist converts a list of fields into their named Go parameters,
// skipping any field names in ignoreFields.
func (a *ArgType) goparamlist(columns []*models.Column, addType bool, ignoreFields ...string) string {
	ignore := map[string]bool{}
	for _, n := range ignoreFields {
		ignore[n] = true
	}

	str := ""
	i := 0
	for _, col := range columns {
		if ignore[col.Field] {
			continue
		}

		s := "v" + strconv.Itoa(i)
		if len(col.Field) > 0 {
			s = strings.ToLower(col.Field[:1]) + col.Field[1:]
		}

		str = str + ", " + s
		if addType {
			str = str + " " + a.retype(col.Type)
		}

		i++
	}

	return str
}

// nulltypeext determines if the type begins with Null (ie, NullString,
// NullInt64, etc) and returns the remainder with the prefix.
func (a *ArgType) nulltypeext(typ string) string {
	if strings.HasPrefix(typ, "sql.Null") {
		return "." + typ[8:]
	}

	return ""
}

// schemafn takes a series of names and joins them with the schema name.
func (a *ArgType) schemafn(s string, names ...string) string {
	n := strings.Join(names, ".")

	if s == "" && n == "" {
		return ""
	}

	if s != "" && n != "" {
		s = s + "."
	}

	return s + n
}
