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

	"github.com/gedex/inflector"
	"github.com/kenshaw/snaker"
	"github.com/xo/xo/models"
	"github.com/xo/xo/templates"
)

// QueryGenerator generates code for custom SQL queries.
//
// Provides a generator that parses a query and writes templates for generated
// type(s).
type QueryGenerator struct{}

// NewQueryGenerator creates a new query generator.
func NewQueryGenerator() *QueryGenerator {
	return &QueryGenerator{}
}

// Satisfies the Generator interface.
func (*QueryGenerator) Exec(ctx context.Context, args *Args) error {
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
	query, inspect, comments, params, err := ParseQuery(
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
	var typ *templates.Type
	if !args.QueryParams.Exec {
		// build query type
		typ, err = BuildQueryType(
			ctx,
			args.QueryParams.Type,
			args.QueryParams.TypeComment,
			args.QueryParams.AllowNulls,
			args.QueryParams.Flat,
			inspect,
			args.QueryParams.Fields,
		)
		if err != nil {
			return err
		}
	}
	// build func name
	name := args.QueryParams.Func
	if name == "" {
		// no func name specified, so generate based on type
		if args.QueryParams.One {
			name = args.QueryParams.Type
		} else {
			name = inflector.Pluralize(args.QueryParams.Type)
		}
		// affix any params
		if len(params) == 0 {
			name = "Get" + name
		} else {
			name += "By"
			for _, p := range params {
				name += strings.ToUpper(p.Name[:1]) + p.Name[1:]
			}
		}
	}
	// emit type template
	if !args.QueryParams.Exec && !args.QueryParams.Flat && !args.OutParams.Append {
		// emit typedef
		if err := templates.Emit(ctx, &templates.Template{
			Set:      "query",
			Template: "typedef",
			Type:     args.QueryParams.Type,
			Data:     typ,
		}); err != nil {
			return err
		}
	}
	// emit query template
	if err := templates.Emit(ctx, &templates.Template{
		Set:      "query",
		Template: "custom",
		Type:     args.QueryParams.Type,
		Data: &templates.Query{
			Name:        name,
			Query:       query,
			Comments:    comments,
			Params:      params,
			One:         args.QueryParams.Exec || args.QueryParams.Flat || args.QueryParams.One,
			Flat:        args.QueryParams.Flat,
			Exec:        args.QueryParams.Exec,
			Interpolate: args.QueryParams.Interpolate,
			Type:        typ,
			Comment:     args.QueryParams.FuncComment,
		},
	}); err != nil {
		return err
	}
	return nil
}

// Process satisfies the Generator interface.
func (*QueryGenerator) Process(ctx context.Context, args *Args) error {
	return templates.Process(
		ctx,
		args.OutParams.Append,
		args.OutParams.Single,
		"typedef", "custom",
	)
}

// ParseQuery parses a query returning the processed query, a query for
// introspection, related comments, and extracted params.
func ParseQuery(ctx context.Context, sqlstr, delimiter string, interpolate, trim, strip bool) ([]string, []string, []string, []*templates.QueryParam, error) {
	_, l, _ := DbLoaderSchema(ctx)
	// build query
	qstr, params, err := ParseQueryParams(
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
	istr, _, err := ParseQueryParams(
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
	return query, inspect, comments, params, nil
}

// ParseQueryParams takes a SQL query and looks for strings in the form of
// "<delim><name> <type>[,<option>,...]<delim>", replacing them with the nth
// param value.
//
// The modified query is returned, along with any extracted parameters.
func ParseQueryParams(query, delim string, interpolate, paramInterpolate bool, nth func(int) string) (string, []*templates.QueryParam, error) {
	// create regexp for delimiter
	placeholderRE, err := regexp.Compile(delim + `[^` + delim[:1] + `]+` + delim)
	if err != nil {
		return "", nil, err
	}
	// grab matches from query string
	matches := placeholderRE.FindAllStringIndex(query, -1)
	// return vals and placeholders
	var params []*templates.QueryParam
	sqlstr, i, last := "", 0, 0
	// loop over matches, extracting each placeholder and splitting to name/type
	for _, m := range matches {
		// extract parameter info
		paramStr := query[m[0]+len(delim) : m[1]-len(delim)]
		p := strings.SplitN(paramStr, " ", 2)
		param := &templates.QueryParam{
			Name: p[0],
			Type: p[1],
		}
		// parse parameter options if present
		if strings.Contains(param.Type, ",") {
			opts := strings.Split(param.Type, ",")
			param.Type = opts[0]
			for _, o := range opts[1:] {
				switch o {
				case "interpolate": // enable interpolation of the variable
					if !interpolate {
						return "", nil, errors.New("query interpolate is not enabled")
					}
					param.Interpolate = true
				case "join": // enable string join of the variable
					param.Join = true
				default:
					return "", nil, fmt.Errorf("unknown option encountered on query parameter %q", paramStr)
				}
			}
		}
		// add to string
		sqlstr = sqlstr + query[last:m[0]]
		if paramInterpolate && param.Interpolate {
			// handle interpolation case
			s := param.Name
			switch {
			case param.Join:
				s = `strings.Join(` + param.Name + `, "\n")`
			case param.Type != "string":
				s = fmt.Sprintf(`fmt.Sprintf("%%v", %s)`, param.Name)
			}
			sqlstr += "` + " + s + " + `"
		} else {
			sqlstr += nth(i)
			i++
		}
		params, last = append(params, param), m[1]
	}
	// return built query and any remaining
	return sqlstr + query[last:], params, nil
}

// BuildQueryType creates the type for the query.
func BuildQueryType(ctx context.Context, name, comment string, allowNulls, flat bool, query []string, fieldstr string) (*templates.Type, error) {
	// template for query type
	typ := &templates.Type{
		Name: name,
		Kind: "table",
		Table: &models.Table{
			TableName: fmt.Sprintf("[custom %s]", strings.ToLower(snaker.CamelToSnake(name))),
		},
		Comment: comment,
	}
	// introspect or use defined user fields
	f := Introspect
	if fieldstr != "" {
		// wrap ...
		f = func(context.Context, []string, bool, bool) ([]*templates.Field, error) {
			return SplitFields(fieldstr)
		}
	}
	// introspect
	var err error
	if typ.Fields, err = f(ctx, query, allowNulls, flat); err != nil {
		return nil, err
	}
	return typ, nil
}

// Introspect creates a view of a query, introspecting the query's columns and
// returning as fields.
//
// Creates a temporary view/table, retrieves its column definitions and
// dropping the temporary view/table.
func Introspect(ctx context.Context, query []string, allowNulls, flat bool) ([]*templates.Field, error) {
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
	var fields []*templates.Field
	for _, col := range cols {
		// get type
		goType, zero, prec, err := l.GoType(ctx, col.DataType, allowNulls && !col.NotNull)
		if err != nil {
			return nil, err
		}
		// determine name
		name := snaker.SnakeToCamelIdentifier(col.ColumnName)
		if flat {
			name = snaker.ForceLowerCamelIdentifier(col.ColumnName)
		}
		fields = append(fields, &templates.Field{
			Name: name,
			Type: goType,
			Zero: zero,
			Prec: prec,
			Col:  col,
		})
	}
	return fields, nil
}

// letters are used for random IDs.
const letters = "abcdefghijklmnopqrstuvwxyz0123456789"

// SplitFields splits s (common separated) into fields.
func SplitFields(s string) ([]*templates.Field, error) {
	var fields []*templates.Field
	for _, field := range strings.Split(s, ",") {
		// process fields
		field = strings.TrimSpace(field)
		name, goType := field, "string"
		if i := strings.Index(field, " "); i != -1 {
			name, goType = field[:i], field[i+1:]
		}
		fields = append(fields, &templates.Field{
			Name: name,
			Type: goType,
			Col: &models.Column{
				ColumnName: snaker.CamelToSnake(name),
			},
		})
	}
	return fields, nil
}
