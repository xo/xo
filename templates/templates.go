// Package templates contains the various Go code templates used by xo.
package templates

import (
	"strconv"
	"strings"
	"text/template"

	"github.com/knq/xo/internal"
	"github.com/knq/xo/models"
	"github.com/serenize/snaker"
)

// Tpls is the collection of template assets.
var Tpls = map[string]*template.Template{}

// KnownTypeMap is the collection of known Go types.
var KnownTypeMap = map[string]bool{
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

// ShortNameTypeMap is the collection of Go style short names for types, mainly
// used for use with declaring a func receiver on a type.
var ShortNameTypeMap = map[string]string{
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

// retype checks the type against the known types, adding the custom type
// package (if any).
func retype(typ string) string {
	if strings.Contains(typ, ".") {
		return typ
	}

	prefix := ""
	for strings.HasPrefix(typ, "[]") {
		typ = typ[2:]
		prefix = prefix + "[]"
	}

	if _, ok := KnownTypeMap[typ]; !ok {
		pkg := internal.CustomTypePackage
		if pkg != "" {
			pkg = pkg + "."
		}

		return prefix + pkg + typ
	}

	return prefix + typ
}

// reniltype checks the nil type against the known types (similar to
// retype), adding the custom type package (if applicable).
func reniltype(typ string) string {
	if strings.Contains(typ, ".") {
		return typ
	}

	if strings.HasSuffix(typ, "{}") {
		if _, ok := KnownTypeMap[typ[:len(typ)-2]]; ok {
			return typ
		}

		pkg := internal.CustomTypePackage
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
func shortname(typ string) string {
	var v string
	var ok bool

	// check short name map
	if v, ok = ShortNameTypeMap[typ]; !ok {
		// calc the short name
		u := []string{}
		for _, s := range strings.Split(strings.ToLower(snaker.CamelToSnake(typ)), "_") {
			if len(s) > 0 && s != "id" {
				u = append(u, s[:1])
			}
		}
		v = strings.Join(u, "")

		// store back in short name map
		ShortNameTypeMap[typ] = v
	}

	return v
}

// tplFuncMap is the func map provided to each template asset.
var tplFuncMap = template.FuncMap{
	// inc increements i by 1.
	"inc": func(i int) int {
		return i + 1
	},

	// colnames creates a list of the column names found in columns, excluding
	// the column with FieldName pkField.
	//
	// Used to present a comma separated list of column names, ie in a sql
	// select, or update. (ie, field_1, field_2, field_3, ...)
	"colnames": func(columns []*models.Column, pkField string) string {
		str := ""
		i := 0
		for _, col := range columns {
			if col.Field == pkField {
				continue
			}

			if i != 0 {
				str = str + ", "
			}
			str = str + col.ColumnName
			i++
		}

		return str
	},

	// colvals creates a list of value place holders for the columns found in
	// columns, excluding the column with FieldName pkField.
	//
	// Used to present a comma separated list of column names, ie in a sql
	// select, or update. (ie, $1, $2, $3 ...)
	"colvals": func(columns []*models.Column, pkField string) string {
		str := ""
		i := 1
		for _, col := range columns {
			if col.Field == pkField {
				continue
			}

			if i != 1 {
				str = str + ", "
			}
			str = str + "$" + strconv.Itoa(i)
			i++
		}

		return str
	},

	// fieldnames creates a list of field names from the field names of the
	// provided columns, excluding the column with field name pkField, and
	// using the prefix provided.
	//
	// Used to present a comma separated list of field names, ie in a Go
	// statement (ie, t.Field1, t.Field2, t.Field3 ...)
	"fieldnames": func(columns []*models.Column, pkField string, prefix string) string {
		str := ""
		i := 0
		for _, col := range columns {
			if col.Field == pkField {
				continue
			}

			if i != 0 {
				str = str + ", "
			}
			str = str + prefix + "." + col.Field
			i++
		}

		return str
	},

	// count returns the 1-based count of columns, excluding any column with
	// field name pkField.
	//
	// Used to get the count of columns, and useful for specifying the last sql
	// parameter.
	"colcount": func(columns []*models.Column, pkField string) int {
		i := 1
		for _, col := range columns {
			if col.Field == pkField {
				continue
			}

			i++
		}
		return i
	},

	// goparamlist converts a list of fields into the named go parameters.
	"goparamlist": func(columns []*models.Column, addType bool) string {
		str := ""
		for i, col := range columns {
			s := "v" + strconv.Itoa(i)
			if len(col.Field) > 0 {
				s = strings.ToLower(col.Field[:1]) + col.Field[1:]
			}

			str = str + ", " + s
			if addType {
				str = str + " " + retype(col.Type)
			}
		}

		return str
	},

	// retype checks the type against the known types, adding the custom type
	// package (if any).
	"retype": retype,

	// reniltype checks the nil type against the known types (similar to
	// retype), adding the custom type package (if applicable).
	"reniltype": reniltype,

	// shortname gets the type's short name, useful for Go receiver func's.
	"shortname": shortname,
}

// init loads the template assets from the stashed binary data.
func init() {
	for _, n := range AssetNames() {
		buf := MustAsset(n)
		Tpls[n] = template.Must(template.New(n).Funcs(tplFuncMap).Parse(string(buf)))
	}
}
