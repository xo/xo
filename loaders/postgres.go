package loaders

import (
	"database/sql"
	"regexp"
	"sort"
	"strconv"
	"strings"

	_ "github.com/lib/pq"

	"github.com/gedex/inflector"
	"github.com/serenize/snaker"

	"github.com/knq/xo/internal"
	"github.com/knq/xo/models"
)

func init() {
	internal.SchemaLoaders["postgres"] = internal.TypeLoader{
		Schemes:        []string{"postgres", "postgresql", "pgsql", "pg"},
		QueryFunc:      PgParseQuery,
		LoadSchemaFunc: PgLoadSchemaTypes,
	}
}

// PgLoadSchemaTypes loads the postgres type definitions from a database.
func PgLoadSchemaTypes(args *internal.ArgType, db *sql.DB) error {
	var err error

	// set schema
	if args.Schema == "" {
		args.Schema = "public"
	}

	// load enums
	_, err = PgLoadEnums(args, db)
	if err != nil {
		return err
	}

	// load tables
	tableMap, err := PgLoadTables(args, db)
	if err != nil {
		return err
	}

	// load foreign relationships
	_, err = PgLoadForeignKeys(args, db, tableMap)
	if err != nil {
		return err
	}

	// load idx
	_, err = PgLoadIndexes(args, db, tableMap)
	if err != nil {
		return err
	}

	// load procs
	_, err = PgLoadProcs(args, db)
	if err != nil {
		return err
	}

	return nil
}

// PgLoadEnums reads the enums from the database, writing the values to the
// args.TypeMap and returning the created EnumTemplates.
func PgLoadEnums(args *internal.ArgType, db *sql.DB) (map[string]*internal.EnumTemplate, error) {
	var err error

	// load enums
	enums, err := models.PgEnumsBySchema(db, args.Schema)
	if err != nil {
		return nil, err
	}

	// process enums
	enumMap := map[string]*internal.EnumTemplate{}
	for _, e := range enums {
		// calc go type and add to known types
		goType := inflector.Singularize(snaker.SnakeToCamel(e.EnumType))
		args.KnownTypeMap[goType] = true

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
			enumMap[typ] = &internal.EnumTemplate{
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
		err = args.ExecuteTemplate(internal.Enum, e.Type, e)
		if err != nil {
			return nil, err
		}
	}

	return enumMap, nil
}

// PgLoadProcs reads the procs from the database, writing the values to the
// args.TypeMap and returning the created ProcTemplates.
func PgLoadProcs(args *internal.ArgType, db *sql.DB) (map[string]*internal.ProcTemplate, error) {
	var err error

	// load procs
	procs, err := models.PgProcsBySchema(db, args.Schema)
	if err != nil {
		return nil, err
	}

	// process procs
	procMap := map[string]*internal.ProcTemplate{}
	for _, p := range procs {
		// fix the name if it starts with underscore
		name := p.ProcName
		if name[0:1] == "_" {
			name = name[1:]
		}

		// create template
		procTpl := internal.ProcTemplate{
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
		err = args.ExecuteTemplate(internal.Proc, "sp_"+p.Name, p)
		if err != nil {
			return nil, err
		}
	}

	return procMap, nil
}

// PgLoadTables loads the table definitions from the database, writing the
// resulting templates to args.TypeMap and returning the created TableTemplates.
func PgLoadTables(args *internal.ArgType, db *sql.DB) (map[string]*internal.TableTemplate, error) {
	var err error

	// load columns
	cols, err := models.PgColumnsByRelkindSchema(db, "r", args.Schema)
	if err != nil {
		return nil, err
	}

	// process columns
	fieldMap := map[string]map[string]bool{}
	tableMap := map[string]*internal.TableTemplate{}
	for _, c := range cols {
		if _, ok := fieldMap[c.TableName]; !ok {
			fieldMap[c.TableName] = map[string]bool{}
		}

		// set col info
		c.Field = snaker.SnakeToCamel(c.ColumnName)
		c.Len, c.NilType, c.Type = PgParseType(args, c.DataType, c.IsNullable)

		// set value in table map if not present
		if _, ok := tableMap[c.TableName]; !ok {
			tableMap[c.TableName] = &internal.TableTemplate{
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
		err = args.ExecuteTemplate(internal.Model, t.Type, t)
		if err != nil {
			return nil, err
		}
	}

	return tableMap, nil
}

// PgLoadForeignKeys loads foreign key relationships from the database.
func PgLoadForeignKeys(args *internal.ArgType, db *sql.DB, tableMap map[string]*internal.TableTemplate) (map[string]*internal.ForeignKeyTemplate, error) {
	var err error

	// load foreign keys
	fkeys, err := models.PgForeignKeysBySchema(db, args.Schema)
	if err != nil {
		return nil, err
	}

	// process foreign keys
	fkMap := map[string]*internal.ForeignKeyTemplate{}
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

				// create foreign key template
				fkMap[fk.ForeignKeyName] = &internal.ForeignKeyTemplate{
					Type:           t.Type,
					ForeignKeyName: fk.ForeignKeyName,
					ColumnName:     fk.ColumnName,
					Field:          f.Field,
					RefType:        tableMap[fk.RefTableName].Type,
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
		err = args.ExecuteTemplate(internal.ForeignKey, fk.Type, fk)
		if err != nil {
			return nil, err
		}
	}

	return fkMap, nil
}

// PgLoadIndexes loads indexes from the database.
func PgLoadIndexes(args *internal.ArgType, db *sql.DB, tableMap map[string]*internal.TableTemplate) (map[string]*internal.IndexTemplate, error) {
	var err error

	// load idx's
	idxMap := map[string]*internal.IndexTemplate{}
	for _, t := range tableMap {
		// find relevant columns
		fields := []*models.Column{}
		for _, f := range t.Fields {
			if f.IsIndex && !f.IsForeignKey {
				if _, ok := idxMap[f.IndexName]; !ok {
					i := &internal.IndexTemplate{
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
		err = args.ExecuteTemplate(internal.Index, idxMap[k].Type, idxMap[k])
		if err != nil {
			return nil, err
		}
	}

	return idxMap, nil
}

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
	if loc := internal.LenRE.FindStringIndex(dt); len(loc) != 0 {
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
		typ = args.Int32Type
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
		typ = args.Uint32Type
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
		if nullable {
			nilType = "pq.NullTime{}"
			typ = "pq.NullTime"
		}

	case "time with time zone", "time without time zone", "timestamp without time zone":
		nilType = "0"
		typ = "int64"
		if nullable {
			nilType = "sql.NullInt64{}"
			typ = "sql.NullInt64"
		}

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

	// special case for []slice
	if typ == "string" && asSlice {
		return precision, "StringSlice{}", "StringSlice"
	}

	// correct type if slice
	if asSlice {
		typ = "[]" + typ
		nilType = "nil"
	}

	return precision, nilType, typ
}

// pgQueryStripRE is the regexp to match the '::type AS name' portion in a query,
// which is a quirk/requirement of generating queries as is done in this
// package.
var pgQueryStripRE = regexp.MustCompile(`(?i)::[a-z][a-z0-9_\.]+\s+AS\s+[a-z][a-z0-9_\.]+`)

// PgParseQuery parses the query and generates a type for it.
func PgParseQuery(args *internal.ArgType, db *sql.DB) error {
	var err error

	// set schema
	if args.Schema == "" {
		args.Schema = "public"
	}

	// parse supplied query
	queryStr, params := args.ParseQuery("$%d")
	inspectStr, _ := args.ParseQuery("NULL")

	// strip out
	if args.QueryStrip {
		queryStr = pgQueryStripRE.ReplaceAllString(queryStr, "")
	}

	// split up query and inspect based on lines
	query := strings.Split(queryStr, "\n")
	inspect := strings.Split(inspectStr, "\n")

	// build query comments with stripped values
	// FIXME: this is off by one, because of golang template syntax limitations
	queryComments := make([]string, len(query)+1)
	if args.QueryStrip {
		for i, l := range inspect {
			pos := pgQueryStripRE.FindStringIndex(l)
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
	xoid := "_xo_" + internal.GenRandomID()
	viewq := `CREATE TEMPORARY VIEW ` + xoid + ` AS (` + strings.Join(inspect, "\n") + `)`
	models.XOLog(viewq)
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
	models.XOLog(nspq, xoid)
	err = db.QueryRow(nspq, xoid).Scan(&schema)
	if err != nil {
		return err
	}

	// load column information ("v" == view)
	cols, err := models.PgColumnsByRelkindSchema(db, "v", schema)
	if err != nil {
		return err
	}

	// create template for query type
	typeTpl := &internal.TableTemplate{
		Type:        args.QueryType,
		TableSchema: args.Schema,
		Fields:      []*models.Column{},
		Comment:     args.QueryTypeComment,
	}

	// process columns
	for _, c := range cols {
		if c.TableName != xoid {
			continue
		}

		c.Field = snaker.SnakeToCamel(c.ColumnName)
		c.Len, c.NilType, c.Type = PgParseType(args, c.DataType, false)
		typeTpl.Fields = append(typeTpl.Fields, c)
	}

	// generate query type template
	err = args.ExecuteTemplate(internal.QueryModel, args.QueryType, typeTpl)
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
				funcName = funcName + strings.ToUpper(p.Name[:1]) + p.Name[1:]
			}
		}
	}

	// create func template
	queryTpl := &internal.QueryTemplate{
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
	err = args.ExecuteTemplate(internal.Query, args.QueryType, queryTpl)
	if err != nil {
		return err
	}

	return nil
}
