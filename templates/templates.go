// Package templates contains the various Go code templates used by xo.
package templates

import (
	"strconv"
	"strings"
	"text/template"

	"github.com/knq/xo/models"
)

// Tpls is the collection of template assets.
var Tpls = map[string]*template.Template{}

// KnownTypeMap is the collection of known Go types.
var KnownTypeMap = map[string]bool{
	"bool":    true,
	"string":  true,
	"byte":    true,
	"rune":    true,
	"int":     true,
	"int16":   true,
	"int32":   true,
	"int64":   true,
	"uint":    true,
	"uint8":   true,
	"uint16":  true,
	"uint32":  true,
	"uint64":  true,
	"float32": true,
	"float64": true,
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

	// retype checks the type against the known types, adding the custom type
	// package (if any).
	"retype": func(typ string) string {
		if strings.Contains(typ, ".") {
			return typ
		}

		prefix := ""
		for strings.HasPrefix(typ, "[]") {
			typ = typ[2:]
			prefix = prefix + "[]"
		}

		if _, ok := KnownTypeMap[typ]; !ok {
			pkg := "" //args.CustomTypePackage
			if pkg != "" {
				pkg = pkg + "."
			}

			return prefix + pkg + typ
		}

		return prefix + typ
	},

	// reniltype, similar to retype checks the nil type against the known
	// types, adding the custom type package (if applicable).
	"reniltype": func(typ string) string {
		if strings.Contains(typ, ".") {
			return typ
		}

		if strings.HasSuffix(typ, "{}") {
			pkg := "" //args.CustomTypePackage
			if pkg != "" {
				pkg = pkg + "."
			}

			return pkg + typ
		}

		return typ
	},
}

// init loads the template assets from the stashed binary data.
func init() {
	for _, n := range AssetNames() {
		buf := MustAsset(n)
		Tpls[n] = template.Must(template.New(n).Funcs(tplFuncMap).Parse(string(buf)))
	}
}
