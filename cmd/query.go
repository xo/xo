package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/gedex/inflector"
	"github.com/kenshaw/snaker"
	"github.com/xo/xo/loader"
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
	db, l, schema := DbLoaderSchema(ctx)
	// read query string from stdin if not provided via --query
	querystr := args.QueryParams.Query
	if querystr == "" {
		buf, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		querystr = string(bytes.TrimRight(buf, "\r\n"))
	}
	// build query
	qstr, params, err := ParseQueryParams(querystr, l.Mask(), true, args.QueryParams.Delimiter, args.QueryParams.Interpolate)
	if err != nil {
		return err
	}
	// build introspection query
	istr, _, err := ParseQueryParams(querystr, "NULL", false, args.QueryParams.Delimiter, args.QueryParams.Interpolate)
	if err != nil {
		return err
	}
	// split up query and inspect based on lines
	query, inspect := strings.Split(qstr, "\n"), strings.Split(istr, "\n")
	// query comment placeholder
	comments := make([]string, len(query)+1)
	// trim whitespace if applicable
	if args.QueryParams.Trim {
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
	// query strip
	if args.QueryParams.Strip && l.QueryStrip != nil {
		l.QueryStrip(query, comments)
	}
	// template for query type
	typ := &templates.Type{
		Name: args.QueryParams.Type,
		Kind: loader.KindTable.String(),
		Table: &models.Table{
			TableName: fmt.Sprintf("[custom %s]", strings.ToLower(snaker.CamelToSnake(args.QueryParams.Type))),
		},
		Comment: args.QueryParams.TypeComment,
	}
	if len(args.QueryParams.Fields) == 0 {
		// if no query fields specified, then load from database
		cols, err := l.QueryColumns(ctx, db, schema, inspect)
		if err != nil {
			return err
		}
		// process columns
		for _, col := range cols {
			// get type
			goType, zero, prec, err := l.GoType(ctx, col.DataType, args.QueryParams.AllowNulls && !col.NotNull)
			if err != nil {
				return err
			}
			// determine name
			name := snaker.SnakeToCamelIdentifier(col.ColumnName)
			if args.QueryParams.Flat {
				name = snaker.ForceLowerCamelIdentifier(col.ColumnName)
			}
			typ.Fields = append(typ.Fields, &templates.Field{
				Name: name,
				Type: goType,
				Zero: zero,
				Prec: prec,
				Col:  col,
			})
		}
	} else {
		// extract fields from query fields
		for _, field := range strings.Split(args.QueryParams.Fields, ",") {
			field = strings.TrimSpace(field)
			name, goType := field, "string"
			if i := strings.Index(field, " "); i != -1 {
				name, goType = field[:i], field[i+1:]
			}
			typ.Fields = append(typ.Fields, &templates.Field{
				Name: name,
				Type: goType,
				Col: &models.Column{
					ColumnName: snaker.CamelToSnake(name),
				},
			})
		}
	}
	if !args.QueryParams.Flat && !args.OutParams.Append {
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
	// emit custom query
	if err := templates.Emit(ctx, &templates.Template{
		Set:      "query",
		Template: "custom",
		Type:     args.QueryParams.Type,
		Data: &templates.Query{
			Name:        name,
			Query:       query,
			Comments:    comments,
			Params:      params,
			One:         args.QueryParams.Flat || args.QueryParams.One,
			Flat:        args.QueryParams.Flat,
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

// ParseQueryParams takes a SQL query and looks for strings in the form of
// "<delim><name> <type>[,<option>,...]<delim>", replacing them with the supplied mask.
//
// Mask can contain "%d" to indicate current position. The modified query is
// returned, along with any extracted parameters.
func ParseQueryParams(query, mask string, interpol bool, delim string, allowInterpol bool) (string, []*templates.QueryParam, error) {
	// create regexp for delimiter
	placeholderRE, err := regexp.Compile(delim + `[^` + delim[:1] + `]+` + delim)
	if err != nil {
		return "", nil, err
	}
	// grab matches from query string
	matches := placeholderRE.FindAllStringIndex(query, -1)
	// return vals and placeholders
	var params []*templates.QueryParam
	sqlstr, i, last := "", 1, 0
	// loop over matches, extracting each placeholder and splitting to name/type
	for _, m := range matches {
		// generate place holder value
		pstr := mask
		if strings.Contains(mask, "%d") {
			pstr = fmt.Sprintf(mask, i)
		}
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
				case "interpolate":
					if !allowInterpol {
						return "", nil, errors.New("query interpolate is not enabled")
					}
					param.Interpolate = true
				default:
					return "", nil, fmt.Errorf("unknown option encountered on query parameter %q", paramStr)
				}
			}
		}
		// add to string
		sqlstr = sqlstr + query[last:m[0]]
		if interpol && param.Interpolate {
			// handle interpolation case
			xstr := param.Name
			if param.Type != "string" {
				xstr = fmt.Sprintf(`fmt.Sprintf("%%v", %s)`, param.Name)
			}
			sqlstr += "` + " + xstr + " + `"
		} else {
			sqlstr += pstr
		}
		params, last = append(params, param), m[1]
		i++
	}
	// return built query and any remaining
	return sqlstr + query[last:], params, nil
}
