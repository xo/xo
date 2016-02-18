package loaders

import (
	"database/sql"
	"fmt"
	"net/url"
	"path"
	"sort"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"

	"github.com/gedex/inflector"
	"github.com/serenize/snaker"

	"github.com/knq/xo/internal"
	"github.com/knq/xo/models"
)

func init() {
	internal.SchemaLoaders["mysql"] = internal.TypeLoader{
		Schemes:        []string{"mysql", "my", "mariadb", "aurora"},
		ProcessDSN:     MyProcessDSN,
		QueryFunc:      MyParseQuery,
		LoadSchemaFunc: MyLoadSchemaTypes,
		ParamN:         func(i int) string { return "?" },
	}
}

// MyProcessDSN processes a mysql DSN.
func MyProcessDSN(u *url.URL, protocol string) string {
	// build host or domain socket
	host := u.Host
	dbname := u.Path[1:]
	if protocol == "unix" {
		host = path.Join(u.Host, path.Dir(u.Path))
		dbname = path.Base(u.Path)
	} else if !strings.Contains(host, ":") {
		// append default port
		host = host + ":3306"
	}

	// build user/pass
	userinfo := ""
	if un := u.User.Username(); len(un) > 0 {
		userinfo = un
		if up, ok := u.User.Password(); ok {
			userinfo = userinfo + ":" + up
		}
	}

	// build params
	params := u.Query().Encode()
	if len(params) > 0 {
		params = "?" + params
	}

	// format
	return fmt.Sprintf(
		"%s@%s(%s)/%s%s",
		userinfo,
		protocol,
		host,
		dbname,
		params,
	)
}

// MyLoadSchemaTypes loads the sqlite type definitions from a database.
func MyLoadSchemaTypes(args *internal.ArgType, db *sql.DB) error {
	var err error

	// schema query
	const sqlstr = `SELECT SCHEMA()`
	models.XOLog(sqlstr)
	err = db.QueryRow(sqlstr).Scan(&args.Schema)
	if err != nil {
		return err
	}

	// load enums
	_, err = MyLoadEnums(args, db)
	if err != nil {
		return err
	}

	// load tables
	tableMap, err := MyLoadTables(args, db)
	if err != nil {
		return err
	}

	// load foreign relationships
	_, err = MyLoadForeignKeys(args, db, tableMap)
	if err != nil {
		return err
	}

	// load idx
	_, err = MyLoadIndexes(args, db, tableMap)
	if err != nil {
		return err
	}

	// load procs
	_, err = MyLoadProcs(args, db)
	if err != nil {
		return err
	}

	return nil
}

// MyLoadEnums reads the enums from the database, writing the values to the
// args.TypeMap and returning the created EnumTemplates.
func MyLoadEnums(args *internal.ArgType, db *sql.DB) (map[string]*internal.EnumTemplate, error) {
	var err error

	// load enums
	enums, err := models.MyEnumsBySchema(db, args.Schema)
	if err != nil {
		return nil, err
	}

	// process enums
	enumMap := map[string]*internal.EnumTemplate{}
	for _, e := range enums {
		// calc go type and add to known types
		typ := inflector.Singularize(snaker.SnakeToCamel(e.EnumType))
		args.KnownTypeMap[typ] = true

		// create enum template
		enumTpl := &internal.EnumTemplate{
			Type:     typ,
			EnumType: e.EnumType,
			Values:   []*models.Enum{},
		}

		// loop over values
		for i, eVal := range strings.Split(e.EnumValues[1:len(e.EnumValues)-1], "','") {
			// calculate val, chopping off redundant type name if applicable
			val := snaker.SnakeToCamel(strings.ToLower(eVal))
			if strings.HasSuffix(strings.ToLower(val), strings.ToLower(typ)) {
				v := val[:len(val)-len(typ)]
				if len(v) > 0 {
					val = v
				}
			}

			// append enum to template vals
			enumTpl.Values = append(enumTpl.Values, &models.Enum{
				EnumType:   e.EnumType,
				EnumValue:  eVal,
				Type:       typ,
				Value:      val,
				ConstValue: i + 1,
			})
		}

		enumMap[typ] = enumTpl
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

// MyLoadProcs reads the procs from the database, writing the values to the
// args.TypeMap and returning the created ProcTemplates.
func MyLoadProcs(args *internal.ArgType, db *sql.DB) (map[string]*internal.ProcTemplate, error) {
	var err error

	// load procs
	procs, err := models.MyProcsBySchema(db, args.Schema)
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
		_, procTpl.NilReturnType, procTpl.ReturnType = MyParseType(args, p.ReturnType, false)

		// split parameter types
		if len(p.ParameterTypes) > 0 {
			for i, paramType := range strings.Split(p.ParameterTypes, ", ") {
				// determine the go equivalent parameter types and add to parameters
				_, _, pt := MyParseType(args, paramType, false)
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

// MyLoadTables loads the table definitions from the database, writing the
// resulting templates to args.TypeMap and returning the created TableTemplates.
func MyLoadTables(args *internal.ArgType, db *sql.DB) (map[string]*internal.TableTemplate, error) {
	var err error

	// load columns
	cols, err := models.MyColumnsByRelkindSchema(db, "BASE TABLE", args.Schema)
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

		// fix issues with mysql data
		if c.IsPrimaryKey { // mysql will return 'PRIMARY' for a primary key index name
			c.IndexName = c.ColumnName
		}
		if c.IndexName != "" {
			c.IsIndex = true
		}
		if c.ForeignIndexName != "" {
			c.IsForeignKey = true
		}

		// set col info
		c.Field = snaker.SnakeToCamel(c.ColumnName)
		c.Len, c.NilType, c.Type = MyParseType(args, c.DataType, c.IsNullable)

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

// MyLoadForeignKeys loads foreign key relationships from the database.
func MyLoadForeignKeys(args *internal.ArgType, db *sql.DB, tableMap map[string]*internal.TableTemplate) (map[string]*internal.ForeignKeyTemplate, error) {
	var err error

	// load foreign keys
	fkeys, err := models.MyForeignKeysBySchema(db, args.Schema)
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

// MyLoadIndexes loads indexes from the database.
func MyLoadIndexes(args *internal.ArgType, db *sql.DB, tableMap map[string]*internal.TableTemplate) (map[string]*internal.IndexTemplate, error) {
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

// MyParseQuery parses the query and generates a type for it.
func MyParseQuery(args *internal.ArgType, db *sql.DB) error {
	var err error

	// schema query
	const sqlstr = `SELECT SCHEMA()`
	models.XOLog(sqlstr)
	err = db.QueryRow(sqlstr).Scan(&args.Schema)
	if err != nil {
		return err
	}

	// parse supplied query
	queryStr, params := args.ParseQuery("?")
	inspectStr, _ := args.ParseQuery("NULL")

	// split up query and inspect based on lines
	query := strings.Split(queryStr, "\n")
	inspect := strings.Split(inspectStr, "\n")

	// query comment placeholder
	queryComments := make([]string, len(query)+1)

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
	viewq := `CREATE VIEW ` + xoid + ` AS (` + strings.Join(inspect, "\n") + `)`
	models.XOLog(viewq)
	_, err = db.Exec(viewq)
	if err != nil {
		return err
	}

	// load column information
	cols, err := models.MyColumnsByRelkindSchema(db, "VIEW", args.Schema)
	if err != nil {
		return err
	}

	// drop inspect view
	dropq := `DROP VIEW ` + xoid
	models.XOLog(dropq)
	_, err = db.Exec(dropq)
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
		c.Len, c.NilType, c.Type = MyParseType(args, c.DataType, false)
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
	funcTpl := &internal.QueryTemplate{
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
	err = args.ExecuteTemplate(internal.Query, args.QueryType, funcTpl)
	if err != nil {
		return err
	}

	return nil
}

// MyParseType parse a postgres type into a Go type based on the column
// definition.
func MyParseType(args *internal.ArgType, dt string, nullable bool) (int, string, string) {
	precision := 0
	nilType := "nil"
	asSlice := false
	unsigned := false

	// extract unsigned
	if strings.HasSuffix(dt, " unsigned") {
		unsigned = true
		dt = dt[:len(dt)-len(" unsigned")]
	}

	// extract length
	if loc := internal.LenRE.FindStringIndex(dt); len(loc) != 0 {
		precision, _ = strconv.Atoi(dt[loc[0]+1 : loc[1]-1])
		dt = dt[:loc[0]]
	}

	var typ string
	switch dt {
	case "bool", "boolean":
		nilType = "false"
		typ = "bool"
		if nullable {
			nilType = "sql.NullBool{}"
			typ = "sql.NullBool"
		}

	case "char", "varchar", "tinytext", "text", "mediumtext", "longtext":
		nilType = `""`
		typ = "string"
		if nullable {
			nilType = "sql.NullString{}"
			typ = "sql.NullString"
		}

	case "tinyint", "smallint", "mediumint":
		nilType = "0"
		typ = "int16"
		if nullable {
			nilType = "sql.NullInt64{}"
			typ = "sql.NullInt64"
		}
	case "int", "integer":
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

	case "decimal", "float":
		nilType = "0.0"
		typ = "float32"
		if nullable {
			nilType = "sql.NullFloat64{}"
			typ = "sql.NullFloat64"
		}
	case "double":
		nilType = "0.0"
		typ = "float64"
		if nullable {
			nilType = "sql.NullFloat64{}"
			typ = "sql.NullFloat64"
		}

	case "binary", "varbinary", "tinyblob", "blob", "mediumblob", "longblob":
		asSlice = true
		typ = "byte"

	case "timestamp", "datetime":
		typ = "*time.Time"
		if nullable {
			nilType = "pq.NullTime{}"
			typ = "pq.NullTime"
		}

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

	// add 'u' as prefix to type if its unsigned
	// FIXME: this needs to be tested properly...
	if unsigned && internal.IntRE.MatchString(typ) {
		typ = "u" + typ
	}

	return precision, nilType, typ
}
