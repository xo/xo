package cmd

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/xo/xo/loader"
	xo "github.com/xo/xo/types"
)

// LoadQuery loads a query.
func LoadQuery(ctx context.Context, set *xo.Set, args *Args) error {
	driver, _, _ := xo.DriverDbSchema(ctx)
	// introspect query if not exec mode
	query, inspect, comments, fields, err := ParseQuery(
		ctx,
		args.QueryParams.Query,
		args.QueryParams.Delimiter,
		args.QueryParams.Interpolate,
		args.QueryParams.Trim,
		args.QueryParams.Strip,
	)
	if err != nil {
		return err
	}
	var typeFields []xo.Field
	if !args.QueryParams.Exec {
		// build query type
		typeFields, err = LoadQueryFields(
			ctx,
			inspect,
			args.QueryParams.Fields,
			args.QueryParams.AllowNulls,
			args.QueryParams.Flat,
		)
		if err != nil {
			return err
		}
	}
	set.Queries = append(set.Queries, xo.Query{
		Driver:       driver,
		Name:         args.QueryParams.Func,
		Comment:      args.QueryParams.FuncComment,
		Exec:         args.QueryParams.Exec,
		Flat:         args.QueryParams.Flat,
		One:          args.QueryParams.One,
		Interpolate:  args.QueryParams.Interpolate,
		Type:         args.QueryParams.Type,
		TypeComment:  args.QueryParams.TypeComment,
		Fields:       typeFields,
		ManualFields: args.QueryParams.Fields != "",
		Params:       fields,
		Query:        query,
		Comments:     comments,
	})
	return nil
}

// ParseQuery parses a query returning the processed query, a query for
// introspection, related comments, and extracted params.
func ParseQuery(ctx context.Context, sqlstr, delimiter string, interpolate, trim, strip bool) ([]string, []string, []string, []xo.Field, error) {
	// nth func
	nth, err := loader.NthParam(ctx)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	// build query
	qstr, fields, err := ParseQueryFields(
		sqlstr,
		delimiter,
		interpolate,
		true,
		nth,
	)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	// build introspection query
	istr, _, err := ParseQueryFields(
		sqlstr,
		delimiter,
		interpolate,
		false,
		func(int) string { return "NULL" },
	)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	// split up query and inspect based on lines
	query, inspect := strings.Split(qstr, "\n"), strings.Split(istr, "\n")
	// trim whitespace if applicable
	if trim {
		for i, line := range query {
			query[i] = strings.TrimSpace(line)
			if i < len(query)-1 {
				query[i] = query[i] + " "
			}
		}
		for i, line := range inspect {
			inspect[i] = strings.TrimSpace(line)
			if i < len(inspect)-1 {
				inspect[i] = inspect[i] + " "
			}
		}
	}
	// build comments
	comments := make([]string, len(query))
	if strip {
		// strip view
		if query, inspect, comments, err = loader.ViewStrip(ctx, query, inspect); err != nil {
			return nil, nil, nil, nil, err
		}
	}
	return query, inspect, comments, fields, nil
}

// ParseQueryFields takes a SQL query and looks for strings in the form of
// "<delim><name> <type>[,<option>,...]<delim>", replacing them with the nth
// param value.
//
// The modified query is returned, along with any extracted parameters.
func ParseQueryFields(query, delim string, interpolate, paramInterpolate bool, nth func(int) string) (string, []xo.Field, error) {
	// create regexp for delimiter
	placeholderRE, err := regexp.Compile(delim + `[^` + delim[:1] + `]+` + delim)
	if err != nil {
		return "", nil, err
	}
	// grab matches from query string
	matches := placeholderRE.FindAllStringIndex(query, -1)
	// return vals and placeholders
	var fields []xo.Field
	sqlstr, i, last := "", 0, 0
	// loop over matches, extracting each placeholder and splitting to name/type
	for _, m := range matches {
		// extract parameter info
		paramStr := query[m[0]+len(delim) : m[1]-len(delim)]
		p := strings.SplitN(paramStr, " ", 2)
		name, typ := p[0], p[1]
		field := xo.Field{
			Name: name,
			Type: xo.Type{
				Type: typ,
			},
		}
		// parse parameter options if present
		if opts := strings.Split(typ, ","); len(opts) > 1 {
			field.Type.Type = opts[0]
			for _, o := range opts[1:] {
				switch o {
				case "interpolate": // enable interpolation of the variable
					if !interpolate {
						return "", nil, errors.New("query interpolate is not enabled")
					}
					field.Interpolate = true
				case "join": // enable string join of the variable
					field.Join = true
				default:
					return "", nil, fmt.Errorf("unknown option encountered on query parameter %q", paramStr)
				}
			}
		}
		// add to string
		sqlstr = sqlstr + query[last:m[0]]
		// determine if parameter previously defined or not
		prevIndex := index(fields, name)
		if paramInterpolate && field.Interpolate {
			// handle interpolation case
			switch {
			case field.Join:
				name = `strings.Join(` + field.Name + `, "\n")`
			case typ != "string":
				name = field.Name
			}
			sqlstr += "` + " + name + " + `"
		} else {
			n := i
			if prevIndex != -1 {
				n = prevIndex
			} else {
				i++
			}
			sqlstr += nth(n)
		}
		// accumulate if not previously encountered
		if prevIndex == -1 {
			fields = append(fields, field)
		}
		last = m[1]
	}
	// return built query and any remaining
	return sqlstr + query[last:], fields, nil
}

// LoadQueryFields loads the query type fields.
func LoadQueryFields(ctx context.Context, query []string, fields string, allowNulls, flat bool) ([]xo.Field, error) {
	// introspect or use defined user fields
	f := Introspect
	if fields != "" {
		// wrap ...
		f = func(context.Context, []string, bool, bool) ([]xo.Field, error) {
			return SplitFields(fields)
		}
	}
	return f(ctx, query, allowNulls, flat)
}

// Introspect creates a view of a query, introspecting the query's columns and
// returning as fields.
//
// Creates a temporary view/table, retrieves its column definitions and
// dropping the temporary view/table.
func Introspect(ctx context.Context, query []string, allowNulls, flat bool) ([]xo.Field, error) {
	// determine prefix
	driver, _, _ := xo.DriverDbSchema(ctx)
	prefix := "_xo_"
	if driver == "oracle" {
		prefix = "XO$"
	}
	// create random id
	id := func(r *rand.Rand) string {
		buf := make([]byte, 8)
		for i := range buf {
			buf[i] = letters[r.Intn(len(letters))]
		}
		return prefix + string(buf)
	}(rand.New(rand.NewSource(time.Now().UTC().UnixNano())))
	// create introspection view
	if _, err := loader.ViewCreate(ctx, id, query); err != nil {
		return nil, err
	}
	// determine schema the view was created in (if applicable)
	schema, err := loader.ViewSchema(ctx, id)
	switch {
	case err != nil:
		return nil, err
	case schema != "":
		ctx = context.WithValue(ctx, xo.SchemaKey, schema)
	}
	// retrieve column info
	cols, err := loader.TableColumns(ctx, id)
	if err != nil {
		return nil, err
	}
	// truncate view
	if _, err := loader.ViewTruncate(ctx, id); err != nil {
		return nil, err
	}
	// drop view
	if _, err := loader.ViewDrop(ctx, id); err != nil {
		return nil, err
	}
	// process columns
	var fields []xo.Field
	for _, col := range cols {
		// get type
		d, err := xo.ParseType(col.DataType, driver)
		if err != nil {
			return nil, err
		}
		if allowNulls {
			d.Nullable = !col.NotNull
		}
		fields = append(fields, xo.Field{
			Name: col.ColumnName,
			Type: d,
		})
	}
	return fields, nil
}

// letters are used for random IDs.
const letters = "abcdefghijklmnopqrstuvwxyz0123456789"

// SplitFields splits s (comma separated) into fields.
func SplitFields(s string) ([]xo.Field, error) {
	var fields []xo.Field
	for _, field := range strings.Split(s, ",") {
		// process fields
		field = strings.TrimSpace(field)
		name, typ := field, "string"
		if i := strings.Index(field, " "); i != -1 {
			name, typ = field[:i], field[i+1:]
		}
		fields = append(fields, xo.Field{
			Name: name,
			Type: xo.Type{
				Type: typ,
			},
		})
	}
	return fields, nil
}

// index finds index of name in v.
func index(v []xo.Field, name string) int {
	for i := 0; i < len(v); i++ {
		if v[i].Name == name {
			return i
		}
	}
	return -1
}
