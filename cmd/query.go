package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"

	xo "github.com/xo/xo/types"
)

// QueryBuilder builds a custom query.
type QueryBuilder struct{}

// NewQueryBuilder creates a new query loader.
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{}
}

// Build satisfies the Builder interface.
func (*QueryBuilder) Build(ctx context.Context, args *Args, dest *xo.XO) error {
	// driver info
	_, l, _ := DbLoaderSchema(ctx)
	// read query string from stdin if not provided via --query
	sqlstr := args.QueryParams.Query
	if sqlstr == "" {
		buf, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		sqlstr = string(bytes.TrimRight(buf, "\r\n"))
	}
	// introspect query if not exec mode
	query, inspect, comments, fields, err := ParseQuery(
		ctx,
		sqlstr,
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
		typeFields, err = LoadTypeFields(
			ctx,
			args.QueryParams.AllowNulls,
			args.QueryParams.Flat,
			inspect,
			args.QueryParams.Fields,
		)
		if err != nil {
			return err
		}
	}
	dest.Emit(xo.Query{
		Driver:       l.Driver,
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
	_, l, _ := DbLoaderSchema(ctx)
	// build query
	qstr, fields, err := ParseQueryFields(
		sqlstr,
		delimiter,
		interpolate,
		true,
		l.NthParam,
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
	var comments []string
	switch {
	case strip && l.Driver == "postgres":
		query, comments = l.ViewStrip(query)
	case strip && l.Driver == "sqlserver":
		inspect, _ = l.ViewStrip(inspect)
		comments = make([]string, len(query))
	default:
		comments = make([]string, len(query))
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
			Datatype: xo.Datatype{
				Type: typ,
			},
		}
		// parse parameter options if present
		if opts := strings.Split(typ, ","); len(opts) > 1 {
			field.Datatype.Type = opts[0]
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
		if paramInterpolate && field.Interpolate {
			// handle interpolation case
			switch {
			case field.Join:
				name = `strings.Join(` + field.Name + `, "\n")`
			case typ != "string":
				name = fmt.Sprintf(`fmt.Sprintf("%%v", %s)`, field.Name)
			}
			sqlstr += "` + " + name + " + `"
		} else {
			sqlstr += nth(i)
			i++
		}
		fields, last = append(fields, field), m[1]
	}
	// return built query and any remaining
	return sqlstr + query[last:], fields, nil
}

// LoadTypeFields loads the query type fields.
func LoadTypeFields(ctx context.Context, allowNulls, flat bool, query []string, fields string) ([]xo.Field, error) {
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
	// create random id
	id := func(r *rand.Rand) string {
		buf := make([]byte, 8)
		for i := range buf {
			buf[i] = letters[r.Intn(len(letters))]
		}
		return string(buf)
	}(rand.New(rand.NewSource(time.Now().UTC().UnixNano())))
	// determine prefix
	db, l, schema := DbLoaderSchema(ctx)
	prefix := "_xo_"
	if l.Driver == "oracle" {
		prefix = "XO$"
	}
	id = prefix + id
	// create introspection view
	if _, err := l.ViewCreate(ctx, db, schema, id, query); err != nil {
		return nil, err
	}
	// determine schema the view was created in (if applicable)
	if l.ViewSchema != nil {
		var err error
		if schema, err = l.ViewSchema(ctx, db, id); err != nil {
			return nil, err
		}
	}
	// retrieve column info
	cols, err := l.TableColumns(ctx, db, schema, id)
	if err != nil {
		return nil, err
	}
	// truncate view
	if l.ViewTruncate != nil {
		if _, err := l.ViewTruncate(ctx, db, schema, id); err != nil {
			return nil, err
		}
	}
	// drop view
	if _, err := l.ViewDrop(ctx, db, schema, id); err != nil {
		return nil, err
	}
	// process columns
	var fields []xo.Field
	for _, col := range cols {
		// get type
		name, prec, scale, array, err := parseType(col.DataType)
		if err != nil {
			return nil, err
		}
		fields = append(fields, xo.Field{
			Name: col.ColumnName,
			Datatype: xo.Datatype{
				Type:  name,
				Prec:  prec,
				Scale: scale,
				Array: array,
			},
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
			Datatype: xo.Datatype{
				Type: typ,
			},
		})
	}
	return fields, nil
}
