package gotpl

import (
	"context"
	"strconv"
	"strings"
	"text/template"

	"github.com/kenshaw/snaker"
	xo "github.com/xo/xo/types"
)

// addLegacyFuncs adds the legacy template funcs.
func addLegacyFuncs(ctx context.Context, funcs template.FuncMap) {
	_, _, nthParam := xo.DriverSchemaNthParam(ctx)
	// colnames creates a list of the column names found in fields, excluding any
	// Field with Name contained in ignoreNames.
	//
	// Used to present a comma separated list of column names, that can be used in
	// a SELECT, or UPDATE, or other SQL clause requiring an list of identifiers
	// (ie, "field_1, field_2, field_3, ...").
	funcs["colnames"] = func(fields []*Field, ignoreNames ...string) string {
		ignore := map[string]bool{}
		for _, n := range ignoreNames {
			ignore[n] = true
		}
		str := ""
		i := 0
		for _, f := range fields {
			if ignore[f.SQLName] {
				continue
			}
			if i != 0 {
				str = str + ", "
			}
			str = str + f.SQLName
			i++
		}
		return str
	}
	// colnamesmulti creates a list of the column names found in fields, excluding any
	// Field with Name contained in ignoreNames.
	//
	// Used to present a comma separated list of column names, that can be used in
	// a SELECT, or UPDATE, or other SQL clause requiring an list of identifiers
	// (ie, "field_1, field_2, field_3, ...").
	funcs["colnamesmulti"] = func(fields []*Field, ignoreNames []*Field) string {
		ignore := map[string]bool{}
		for _, f := range ignoreNames {
			ignore[f.SQLName] = true
		}
		str := ""
		i := 0
		for _, f := range fields {
			if ignore[f.SQLName] {
				continue
			}
			if i != 0 {
				str = str + ", "
			}
			str = str + f.SQLName
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
	funcs["colnamesquery"] = func(fields []*Field, sep string, ignoreNames ...string) string {
		ignore := map[string]bool{}
		for _, n := range ignoreNames {
			ignore[n] = true
		}
		str := ""
		i := 0
		for _, f := range fields {
			if ignore[f.SQLName] {
				continue
			}
			if i != 0 {
				str = str + sep
			}
			str = str + f.SQLName + " = " + nthParam(i)
			i++
		}
		return str
	}
	// colnamesquerymulti creates a list of the column names in fields as a query and
	// joined by sep, excluding any Field with Name contained in the slice of fields in ignoreNames.
	//
	// Used to create a list of column names in a WHERE clause (ie, "field_1 = $1
	// AND field_2 = $2 AND ...") or in an UPDATE clause (ie, "field = $1, field =
	// $2, ...").
	funcs["colnamesquerymulti"] = func(fields []*Field, sep string, startCount int, ignoreNames []*Field) string {
		ignore := map[string]bool{}
		for _, f := range ignoreNames {
			ignore[f.SQLName] = true
		}
		str := ""
		i := startCount
		for _, f := range fields {
			if ignore[f.SQLName] {
				continue
			}
			if i > startCount {
				str = str + sep
			}
			str = str + f.SQLName + " = " + nthParam(i)
			i++
		}
		return str
	}
	// colprefixnames creates a list of the column names found in fields with the
	// supplied prefix, excluding any Field with Name contained in ignoreNames.
	//
	// Used to present a comma separated list of column names with a prefix. Used in
	// a SELECT, or UPDATE (ie, "t.field_1, t.field_2, t.field_3, ...").
	funcs["colprefixnames"] = func(fields []*Field, prefix string, ignoreNames ...string) string {
		ignore := map[string]bool{}
		for _, n := range ignoreNames {
			ignore[n] = true
		}
		str := ""
		i := 0
		for _, f := range fields {
			if ignore[f.SQLName] {
				continue
			}
			if i != 0 {
				str = str + ", "
			}
			str = str + prefix + "." + f.SQLName
			i++
		}
		return str
	}
	// colvals creates a list of value place holders for fields excluding any Field
	// with Name contained in ignoreNames.
	//
	// Used to present a comma separated list of column place holders, used in a
	// SELECT or UPDATE statement (ie, "$1, $2, $3 ...").
	funcs["colvals"] = func(fields []*Field, ignoreNames ...string) string {
		ignore := map[string]bool{}
		for _, n := range ignoreNames {
			ignore[n] = true
		}
		str := ""
		i := 0
		for _, f := range fields {
			if ignore[f.SQLName] {
				continue
			}
			if i != 0 {
				str = str + ", "
			}
			str = str + nthParam(i)
			i++
		}
		return str
	}
	// colvalsmulti creates a list of value place holders for fields excluding any Field
	// with Name contained in ignoreNames.
	//
	// Used to present a comma separated list of column place holders, used in a
	// SELECT or UPDATE statement (ie, "$1, $2, $3 ...").
	funcs["colvalsmulti"] = func(fields []*Field, ignoreNames []*Field) string {
		ignore := map[string]bool{}
		for _, f := range ignoreNames {
			ignore[f.SQLName] = true
		}
		str := ""
		i := 0
		for _, f := range fields {
			if ignore[f.SQLName] {
				continue
			}
			if i != 0 {
				str = str + ", "
			}
			str = str + nthParam(i)
			i++
		}
		return str
	}
	// fieldnames creates a list of field names from fields of the adding the
	// provided prefix, and excluding any Field with Name contained in ignoreNames.
	//
	// Used to present a comma separated list of field names, ie in a Go statement
	// (ie, "t.Field1, t.Field2, t.Field3 ...")
	funcs["fieldnames"] = func(fields []*Field, prefix string, ignoreNames ...string) string {
		ignore := map[string]bool{}
		for _, n := range ignoreNames {
			ignore[n] = true
		}
		str := ""
		i := 0
		for _, f := range fields {
			if ignore[f.SQLName] {
				continue
			}
			if i != 0 {
				str = str + ", "
			}
			str = str + prefix + "." + f.SQLName
			i++
		}
		return str
	}
	// fieldnamesmulti creates a list of field names from fields of the adding the
	// provided prefix, and excluding any Field with the slice contained in ignoreNames.
	//
	// Used to present a comma separated list of field names, ie in a Go statement
	// (ie, "t.Field1, t.Field2, t.Field3 ...")
	funcs["fieldnamesmulti"] = func(fields []*Field, prefix string, ignoreNames []*Field) string {
		ignore := map[string]bool{}
		for _, f := range ignoreNames {
			ignore[f.SQLName] = true
		}
		str := ""
		i := 0
		for _, f := range fields {
			if ignore[f.SQLName] {
				continue
			}
			if i != 0 {
				str = str + ", "
			}
			str = str + prefix + "." + f.SQLName
			i++
		}
		return str
	}
	// colcount returns the 1-based count of fields, excluding any Field with Name
	// contained in ignoreNames.
	//
	// Used to get the count of fields, and useful for specifying the last SQL
	// parameter.
	funcs["colcount"] = func(fields []*Field, ignoreNames ...string) int {
		ignore := map[string]bool{}
		for _, n := range ignoreNames {
			ignore[n] = true
		}
		i := 1
		for _, f := range fields {
			if ignore[f.SQLName] {
				continue
			}
			i++
		}
		return i
	}
	// goparamlist converts a list of fields into their named Go parameters,
	// skipping any Field with Name contained in ignoreNames. addType will cause
	// the go Type to be added after each variable name. addPrefix will cause the
	// returned string to be prefixed with ", " if the generated string is not
	// empty.
	//
	// Any field name encountered will be checked against goReservedNames, and will
	// have its name substituted by its corresponding looked up value.
	//
	// Used to present a comma separated list of Go variable names for use with as
	// either a Go func parameter list, or in a call to another Go func.
	// (ie, ", a, b, c, ..." or ", a T1, b T2, c T3, ...").
	funcs["goparamlist"] = func(fields []*Field, addPrefix bool, addType bool, ignoreNames ...string) string {
		ignore := map[string]bool{}
		for _, n := range ignoreNames {
			ignore[n] = true
		}
		i := 0
		vals := []string{}
		for _, f := range fields {
			if ignore[f.SQLName] {
				continue
			}
			s := "v" + strconv.Itoa(i)
			if len(f.SQLName) > 0 {
				n := strings.Split(snaker.CamelToSnake(f.SQLName), "_")
				s = strings.ToLower(n[0]) + f.SQLName[len(n[0]):]
			}
			// add the go type
			if addType {
				s += " " + f.Type
			}
			// add to vals
			vals = append(vals, s)
			i++
		}
		// concat generated values
		str := strings.Join(vals, ", ")
		if addPrefix && str != "" {
			return ", " + str
		}
		return str
	}
	// convext generates the Go conversion for f in order for it to be assignable
	// to t.
	//
	// FIXME: this should be a better name, like "goconversion" or some such.
	funcs["convext"] = func(prefix string, f *Field, t *Field) string {
		expr := prefix + "." + f.SQLName
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
	// getstartcount returns a starting count for numbering columsn in queries
	funcs["getstartcount"] = func(fields []*Field, pkFields []*Field) int {
		return len(fields) - len(pkFields)
	}
}
