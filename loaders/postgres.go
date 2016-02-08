package loaders

import (
	"bytes"
	"database/sql"
	"fmt"
	"math/rand"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gedex/inflector"
	"github.com/knq/xo/internal"
	"github.com/knq/xo/models"
	"github.com/knq/xo/templates"
	"github.com/serenize/snaker"
)

// PgLoadSchemaTypes loads the postgres type definitions from a database.
func PgLoadSchemaTypes(args *internal.ArgType, db *sql.DB, typeMap map[string]*bytes.Buffer) error {
	var err error

	// load enums
	_, err = PgLoadEnums(args, db, typeMap)
	if err != nil {
		return err
	}

	// load tables
	tableMap, err := PgLoadTables(args, db, typeMap)
	if err != nil {
		return err
	}

	// load foreign relationships
	_, err = PgLoadForeignKeys(args, db, typeMap, tableMap)
	if err != nil {
		return err
	}

	// load idx
	_, err = PgLoadIdx(args, db, typeMap, tableMap)
	if err != nil {
		return err
	}

	// load procs
	_, err = PgLoadProcs(args, db, typeMap)
	if err != nil {
		return err
	}

	return nil
}

// PgLoadEnums reads the enums from the database, writing the values to the
// typeMap and returning the created EnumTemplates.
func PgLoadEnums(args *internal.ArgType, db *sql.DB, typeMap map[string]*bytes.Buffer) (map[string]*templates.EnumTemplate, error) {
	var err error

	// load enums
	enums, err := models.EnumsBySchema(db, args.Schema)
	if err != nil {
		return nil, err
	}

	// process enums
	enumMap := map[string]*templates.EnumTemplate{}
	for _, e := range enums {
		// calc go type and add to known types
		goType := inflector.Singularize(snaker.SnakeToCamel(e.EnumType))
		templates.KnownTypeMap[goType] = true

		// calculate val, chopping off redundant type name if applicable
		val := snaker.SnakeToCamel(strings.ToLower(e.EnumValue))
		if strings.HasSuffix(strings.ToLower(val), strings.ToLower(goType)) {
			v := val[:len(val)-len(goType)]
			if len(v) > 0 {
				val = v
			}
		}

		// copy values back into model
		e.Value = val
		e.Type = goType

		// set value in enum map if not present
		typ := strings.ToLower(goType)
		if _, ok := enumMap[typ]; !ok {
			enumMap[typ] = &templates.EnumTemplate{
				Type:     goType,
				EnumType: e.EnumType,
				Values:   []*models.Enum{},
			}
		}

		// append enum to template vals
		enumMap[typ].Values = append(enumMap[typ].Values, e)
	}

	// generate enum templates
	for _, e := range enumMap {
		buf := GetBuf(typeMap, strings.ToLower(e.Type))
		err = templates.Tpls["postgres.enum.go.tpl"].Execute(buf, e)
		if err != nil {
			return nil, err
		}
	}

	return enumMap, nil
}

// PgLoadProcs reads the procs from the database, writing the values to the
// typeMap and returning the created ProcTemplates.
func PgLoadProcs(args *internal.ArgType, db *sql.DB, typeMap map[string]*bytes.Buffer) (map[string]*templates.ProcTemplate, error) {
	var err error

	// load procs
	procs, err := models.ProcsBySchema(db, args.Schema)
	if err != nil {
		return nil, err
	}

	// process procs
	procMap := map[string]*templates.ProcTemplate{}
	for _, p := range procs {
		// fix the name if it starts with underscore
		name := p.ProcName
		if name[0:1] == "_" {
			name = name[1:]
		}

		// create template
		procTpl := templates.ProcTemplate{
			Name:               snaker.SnakeToCamel(name),
			TableSchema:        args.Schema,
			ProcName:           p.ProcName,
			ProcParameterTypes: p.ParameterTypes,
			ProcReturnType:     p.ReturnType,
			Parameters:         []*models.Column{},
		}

		// parse return type into template
		_, procTpl.NilReturnType, procTpl.ReturnType = PgParseType(args, p.ReturnType, false)

		// split parameter types
		if len(p.ParameterTypes) > 0 {
			for i, paramType := range strings.Split(p.ParameterTypes, ", ") {
				// determine the go equivalent parameter types and add to parameters
				_, _, pt := PgParseType(args, paramType, false)
				procTpl.Parameters = append(procTpl.Parameters, &models.Column{
					Field: "v" + strconv.Itoa(i),
					Type:  pt,
				})
			}
		}

		procMap[p.ProcName] = &procTpl
	}

	// generate proc templates
	for _, p := range procMap {
		buf := GetBuf(typeMap, strings.ToLower("sp_"+p.Name))
		err = templates.Tpls["postgres.proc.go.tpl"].Execute(buf, p)
		if err != nil {
			return nil, err
		}
	}

	return procMap, nil
}

// PgLoadTables loads the table definitions from the database, writing the
// resulting templates to typeMap and returning the created templates.TableTemplates.
func PgLoadTables(args *internal.ArgType, db *sql.DB, typeMap map[string]*bytes.Buffer) (map[string]*templates.TableTemplate, error) {
	var err error

	// load columns
	cols, err := models.ColumnsByRelkindSchema(db, "t", args.Schema)
	if err != nil {
		return nil, err
	}

	// process columns
	fieldMap := map[string]map[string]bool{}
	tableMap := map[string]*templates.TableTemplate{}
	for _, c := range cols {
		if _, ok := fieldMap[c.TableName]; !ok {
			fieldMap[c.TableName] = map[string]bool{}
		}

		// set col info
		c.Field = snaker.SnakeToCamel(c.ColumnName)
		c.Len, c.NilType, c.Type = PgParseType(args, c.DataType, c.IsNullable)

		// set value in table map if not present
		if _, ok := tableMap[c.TableName]; !ok {
			tableMap[c.TableName] = &templates.TableTemplate{
				Type:        inflector.Singularize(snaker.SnakeToCamel(c.TableName)),
				TableSchema: args.Schema,
				TableName:   c.TableName,
				Fields:      []*models.Column{},
			}
		}

		// set primary key
		if c.IsPrimaryKey {
			tableMap[c.TableName].PrimaryKey = c.ColumnName
			tableMap[c.TableName].PrimaryKeyField = c.Field
			tableMap[c.TableName].PrimaryKeyType = c.Type
		}

		// append col to template fields
		if _, ok := fieldMap[c.TableName][c.ColumnName]; !ok {
			tableMap[c.TableName].Fields = append(tableMap[c.TableName].Fields, c)
		}

		fieldMap[c.TableName][c.ColumnName] = true
	}

	// generate table templates
	for _, t := range tableMap {
		buf := GetBuf(typeMap, strings.ToLower(t.Type))
		err = templates.Tpls["postgres.model.go.tpl"].Execute(buf, t)
		if err != nil {
			return nil, err
		}
	}

	return tableMap, nil
}

// PgLoadForeignKeys loads foreign key relationships from the database.
func PgLoadForeignKeys(args *internal.ArgType, db *sql.DB, typeMap map[string]*bytes.Buffer, tableMap map[string]*templates.TableTemplate) (map[string]*templates.FkTemplate, error) {
	var err error

	// load foreign keys
	fkeys, err := models.ForeignKeysBySchema(db, args.Schema)
	if err != nil {
		return nil, err
	}

	// process foreign keys
	fkMap := map[string]*templates.FkTemplate{}
	for _, t := range tableMap {
		for _, f := range t.Fields {
			if f.IsForeignKey {
				var fk *models.ForeignKey
				// find foreign key
				for _, v := range fkeys {
					if f.ForeignIndexName == v.ForeignKeyName {
						fk = v
						break
					}
				}

				fkMap[fk.ForeignKeyName] = &templates.FkTemplate{
					Type:       t.Type,
					ColumnName: fk.ColumnName,
					Field:      f.Field,
					RefType:    tableMap[fk.RefTableName].Type,
				}

				// find field
				for _, f := range tableMap[fk.RefTableName].Fields {
					if f.ColumnName == fk.RefColumnName {
						fkMap[fk.ForeignKeyName].RefField = f.Field
						break
					}
				}
			}
		}
	}

	// generate templates
	for _, fk := range fkMap {
		buf := GetBuf(typeMap, strings.ToLower(fk.Type))
		err = templates.Tpls["postgres.fkey.go.tpl"].Execute(buf, fk)
		if err != nil {
			return nil, err
		}
	}

	return fkMap, nil
}

// PgLoadIdx loads indexes from the database.
func PgLoadIdx(args *internal.ArgType, db *sql.DB, typeMap map[string]*bytes.Buffer, tableMap map[string]*templates.TableTemplate) (map[string]*templates.IdxTemplate, error) {
	var err error

	// load idx's
	idxMap := map[string]*templates.IdxTemplate{}
	for _, t := range tableMap {
		// find relevant columns
		fields := []*models.Column{}
		for _, f := range t.Fields {
			if f.IsIndex && !f.IsForeignKey {
				if _, ok := idxMap[f.IndexName]; !ok {
					i := &templates.IdxTemplate{
						Type:        t.Type,
						Name:        snaker.SnakeToCamel(f.IndexName),
						TableSchema: t.TableSchema,
						TableName:   f.TableName,
						IndexName:   f.IndexName,
						IsUnique:    f.IsUnique,
						Fields:      fields,
						Table:       t,
					}

					// non unique lookup
					if !f.IsUnique {
						idxName := i.IndexName

						// chop off tablename_
						if strings.HasPrefix(idxName, f.TableName+"_") {
							idxName = idxName[len(f.TableName)+1:]
						}

						// chop off _idx or _index
						switch {
						case strings.HasSuffix(idxName, "_idx"):
							idxName = idxName[:len(idxName)-len("_idx")]
						case strings.HasSuffix(idxName, "_index"):
							idxName = idxName[:len(idxName)-len("_index")]
						}

						i.Name = snaker.SnakeToCamel(idxName)
						i.Plural = inflector.Pluralize(t.Type)
					}

					idxMap[f.IndexName] = i
				}

				idxMap[f.IndexName].Fields = append(idxMap[f.IndexName].Fields, f)
			}
		}
	}

	// idx keys
	idxKeys := []string{}
	for k := range idxMap {
		idxKeys = append(idxKeys, k)
	}
	sort.Strings(idxKeys)

	// generate templates
	for _, k := range idxKeys {
		buf := GetBuf(typeMap, strings.ToLower(idxMap[k].Type))
		err = templates.Tpls["postgres.idx.go.tpl"].Execute(buf, idxMap[k])
		if err != nil {
			return nil, err
		}
	}

	return idxMap, nil
}

// pgLenRE is a regular expression that matches precision (length) definitions in
// a pg database.
var pgLenRE = regexp.MustCompile(`\([0-9]+\)$`)

// PgParseType parse a postgres type into a Go type based on the column
// definition.
func PgParseType(args *internal.ArgType, dt string, nullable bool) (int, string, string) {
	precision := 0
	nilType := "nil"
	asSlice := false

	// handle SETOF
	if strings.HasPrefix(dt, "SETOF ") {
		_, _, t := PgParseType(args, dt[len("SETOF "):], false)
		return 0, "nil", "[]" + t
	}

	// determine if it's a slice
	if strings.HasSuffix(dt, "[]") {
		dt = dt[:len(dt)-2]
		asSlice = true
	}

	// extract length
	if loc := pgLenRE.FindStringIndex(dt); len(loc) != 0 {
		precision, _ = strconv.Atoi(dt[loc[0]+1 : loc[1]-1])
		dt = dt[:loc[0]]
	}

	var typ string
	switch dt {
	case "boolean":
		nilType = "false"
		typ = "bool"
		if nullable {
			nilType = "sql.NullBool{}"
			typ = "sql.NullBool"
		}

	case "character", "character varying", "text":
		nilType = `""`
		typ = "string"
		if nullable {
			nilType = "sql.NullString{}"
			typ = "sql.NullString"
		}

	case "smallint":
		nilType = "0"
		typ = "int16"
		if nullable {
			nilType = "sql.NullInt64{}"
			typ = "sql.NullInt64"
		}
	case "integer":
		nilType = "0"
		typ = "int32"
		if nullable {
			nilType = "sql.NullInt64{}"
			typ = "sql.NullInt64"
		}
	case "bigint":
		nilType = "0"
		typ = "int64"
		if nullable {
			nilType = "sql.NullInt64{}"
			typ = "sql.NullInt64"
		}

	case "smallserial":
		nilType = "0"
		typ = "uint16"
		if nullable {
			nilType = "sql.NullInt64{}"
			typ = "sql.NullInt64"
		}
	case "serial":
		nilType = "0"
		typ = "uint32"
		if nullable {
			nilType = "sql.NullInt64{}"
			typ = "sql.NullInt64"
		}
	case "bigserial":
		nilType = "0"
		typ = "uint64"
		if nullable {
			nilType = "sql.NullInt64{}"
			typ = "sql.NullInt64"
		}

	case "real":
		nilType = "0.0"
		typ = "float32"
		if nullable {
			nilType = "sql.NullFloat64{}"
			typ = "sql.NullFloat64"
		}
	case "double precision":
		nilType = "0.0"
		typ = "float64"
		if nullable {
			nilType = "sql.NullFloat64{}"
			typ = "sql.NullFloat64"
		}

	case "bytea":
		asSlice = true
		typ = "byte"

	case "timestamp with time zone":
		typ = "*time.Time"

	case "time with time zone", "time without time zone", "timestamp without time zone":
		nilType = "0"
		typ = "int64"

	case "interval":
		typ = "*time.Duration"

	case `"char"`, "bit":
		// FIXME: this needs to actually be tested ...
		// i think this should be 'rune' but I don't think database/sql
		// supports 'rune' as a type?
		//
		// this is mainly here because postgres's pg_catalog.* meta tables have
		// this as a type.
		//typ = "rune"
		nilType = `uint8(0)`
		typ = "uint8"

	case `"any"`, "bit varying":
		asSlice = true
		typ = "byte"

	default:
		if strings.HasPrefix(dt, args.Schema+".") {
			// in the same schema, so chop off
			typ = snaker.SnakeToCamel(dt[len(args.Schema)+1:])
			nilType = typ + "(0)"
		} else {
			typ = snaker.SnakeToCamel(dt)
			nilType = typ + "{}"
		}
	}

	// correct type if slice
	if asSlice {
		typ = "[]" + typ
		nilType = "nil"
	}

	return precision, nilType, typ
}

// parseQuery takes the query in args and looks for strings in the form of
// "%%<name> <type>%%", replacing them with the supplied mask. mask can contain
// "%d" to indicate current position. The modified query is returned, and the
// extracted text.
func parseQuery(args *internal.ArgType, query string, mask string) (string, [][]string) {
	// create the regexp for the delimiter
	placeholderRE := regexp.MustCompile(
		args.QueryParamDelimiter + `[^` + args.QueryParamDelimiter[:1] + `]+` + args.QueryParamDelimiter,
	)

	// grab matches from query string
	matches := placeholderRE.FindAllStringIndex(query, -1)

	// no matches, so return query unmodified
	/*if matches == nil || len(matches) == 0 {
		return query, [][]string{}
	}*/

	// return vals and placeholders
	str := ""
	params := [][]string{}
	i := 1
	last := 0

	// loop over matches, extracting each placeholder and splitting to name/type
	for _, m := range matches {
		// generate place holder value
		pstr := mask
		if strings.Contains(mask, "%d") {
			pstr = fmt.Sprintf(mask, i)
		}

		// build string
		str = str + args.Query[last:m[0]] + pstr
		params = append(params, strings.SplitN(
			query[m[0]+len(args.QueryParamDelimiter):m[1]-len(args.QueryParamDelimiter)],
			" ",
			2,
		))
		last = m[1]
		i++
	}

	// add part of query remains
	str = str + args.Query[last:]

	return str, params
}

// letters for genRandomID
var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

// genRandomID generates a 8 character random string.
func genRandomID() string {
	rand.Seed(time.Now().UTC().UnixNano())
	b := make([]rune, 8)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// queryStripRE is the regexp to match the '::type AS name' portion in a query,
// which is a quirk/requirement of generating queries as is done in this
// package.
var queryStripRE = regexp.MustCompile(`(?i)::[a-z][a-z0-9_\.]+\s+AS\s+[a-z][a-z0-9_\.]+`)

// PgParseQuery parses the query and generates a type for it.
func PgParseQuery(args *internal.ArgType, db *sql.DB, typeMap map[string]*bytes.Buffer) error {
	var err error

	// parse supplied query
	queryStr, params := parseQuery(args, args.Query, "$%d")
	inspectStr, _ := parseQuery(args, args.Query, "NULL")

	// strip out
	if args.QueryStrip {
		queryStr = queryStripRE.ReplaceAllString(queryStr, "")
	}

	// split up query and inspect based on lines
	query := strings.Split(queryStr, "\n")
	inspect := strings.Split(inspectStr, "\n")

	// build query comments with stripped values
	// FIXME: this is off by one, because of golang template syntax limitations
	queryComments := make([]string, len(query)+1)
	if args.QueryStrip {
		for i, l := range inspect {
			pos := queryStripRE.FindStringIndex(l)
			if pos != nil {
				queryComments[i+1] = l[pos[0]:pos[1]]
			} else {
				queryComments[i+1] = ""
			}
		}
	}

	// trim whitespace if applicable
	if args.QueryTrim {
		for n, l := range query {
			query[n] = strings.TrimSpace(l)
			if n < len(query)-1 {
				query[n] = query[n] + " "
			}
		}

		for n, l := range inspect {
			inspect[n] = strings.TrimSpace(l)
			if n < len(inspect)-1 {
				inspect[n] = inspect[n] + " "
			}
		}
	}

	// create temporary view xoid
	xoid := "_xo_" + genRandomID()
	viewq := `CREATE TEMPORARY VIEW ` + xoid + ` AS (` + strings.Join(inspect, "\n") + `)`
	_, err = db.Exec(viewq)
	if err != nil {
		return err
	}

	// determine schema name temporary view was created on
	// sql query
	var nspq = `SELECT n.nspname ` +
		`FROM pg_class c ` +
		`JOIN pg_namespace n ON n.oid = c.relnamespace ` +
		`WHERE n.nspname LIKE 'pg_temp%' AND c.relname = $1`

	// run schema query
	var schema string
	err = db.QueryRow(nspq, xoid).Scan(&schema)
	if err != nil {
		return err
	}

	// load column information ("v" == view)
	cols, err := models.ColumnsByRelkindSchema(db, "v", schema)
	if err != nil {
		return err
	}

	// create template for query type
	typeTpl := &templates.TableTemplate{
		Type:        args.QueryType,
		TableSchema: args.Schema,
		Fields:      []*models.Column{},
		Comment:     args.QueryTypeComment,
	}

	// process columns
	for _, c := range cols {
		c.Field = snaker.SnakeToCamel(c.ColumnName)
		c.Len, c.NilType, c.Type = PgParseType(args, c.DataType, false)
		typeTpl.Fields = append(typeTpl.Fields, c)
	}

	// generate query type template
	buf := GetBuf(typeMap, strings.ToLower(args.QueryType))
	err = templates.Tpls["postgres.model.go.tpl"].Execute(buf, typeTpl)
	if err != nil {
		return err
	}

	// build func name
	funcName := args.QueryFunc
	if funcName == "" {
		// no func name specified, so generate based on type
		if args.QueryOnlyOne {
			funcName = args.QueryType
		} else {
			funcName = inflector.Pluralize(args.QueryType)
		}

		// affix any params
		if len(params) == 0 {
			funcName = "Get" + funcName
		} else {
			funcName = funcName + "By"
			for _, p := range params {
				funcName = funcName + strings.ToUpper(p[0][:1]) + p[0][1:]
			}
		}
	}

	// create func template
	funcTpl := &templates.FuncTemplate{
		Name:          funcName,
		Type:          args.QueryType,
		Query:         query,
		QueryComments: queryComments,
		Parameters:    params,
		OnlyOne:       args.QueryOnlyOne,
		Comment:       args.QueryFuncComment,
		Table:         typeTpl,
	}

	// generate template
	err = templates.Tpls["postgres.func.go.tpl"].Execute(buf, funcTpl)
	if err != nil {
		return err
	}

	return nil
}
