package loaders

import (
	"database/sql"
	"fmt"
	"net/url"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/gedex/inflector"
	"github.com/serenize/snaker"

	_ "github.com/mattn/go-sqlite3"

	"github.com/knq/xo/internal"
	"github.com/knq/xo/models"
)

func init() {
	internal.SchemaLoaders["sqlite3"] = internal.TypeLoader{
		Schemes:        []string{"sqlite3", "sqlite", "file"},
		ProcessDSN:     SqProcessDSN,
		QueryFunc:      SqParseQuery,
		LoadSchemaFunc: SqLoadSchemaTypes,
		ParamN:         func(i int) string { return "?" },
	}
}

// SqProcessDSN processes a sqlite DSN.
func SqProcessDSN(u *url.URL, protocol string) string {
	p := u.Opaque
	if u.Path != "" {
		p = u.Path
	}

	if u.Host != "" && u.Host != "localhost" {
		p = path.Join(u.Host, p)
	}

	return p + u.Query().Encode()
}

// SqLoadSchemaTypes loads the sqlite type definitions from a database.
func SqLoadSchemaTypes(args *internal.ArgType, db *sql.DB) error {
	var err error

	// load tables
	tableMap, err := SqLoadTables(args, db)
	if err != nil {
		return err
	}

	return nil

	// load foreign relationships
	_, err = SqLoadForeignKeys(args, db, tableMap)
	if err != nil {
		return err
	}

	// load idx
	_, err = SqLoadIndexes(args, db, tableMap)
	if err != nil {
		return err
	}

	return nil
}

// SqLoadTables loads the table definitions from the database, writing the
// resulting templates to args.TypeMap and returning the created TableTemplates.
func SqLoadTables(args *internal.ArgType, db *sql.DB) (map[string]*internal.TableTemplate, error) {
	var err error

	// load table info
	tableInfo, err := models.SqTableInfosByRelkind(db, "table")
	if err != nil {
		return nil, err
	}

	// loop over tables and get info for each
	tableMap := make(map[string]*internal.TableTemplate)
	for _, ti := range tableInfo {
		// create table template
		tableTpl := &internal.TableTemplate{
			Type:      inflector.Singularize(snaker.SnakeToCamel(ti.TableName)),
			TableName: ti.TableName,
			Fields:    []*models.Column{},
		}

		// load columns
		err = SqLoadColumns(args, db, tableTpl)
		if err != nil {
			return nil, err
		}
		tableMap[tableTpl.Type] = tableTpl
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

// SqLoadColumns loads the column definition from the database, writing the
// adding the fields to the supplied table template.
func SqLoadColumns(args *internal.ArgType, db *sql.DB, tableTpl *internal.TableTemplate) error {
	var err error

	// load columns
	cols, err := models.SqColumnsByTable(db, tableTpl.TableName)
	if err != nil {
		return err
	}

	// process columns
	for _, sqC := range cols {
		// copy all the sq column fields
		c := &models.Column{
			FieldOrdinal: sqC.FieldOrdinal,
			ColumnName:   sqC.ColumnName,
			DataType:     sqC.DataType,
			IsNullable:   sqC.IsNullable,
			DefaultValue: sqC.DefaultValue.String,
			HasDefault:   sqC.DefaultValue.Valid,
			IsPrimaryKey: sqC.IsPrimaryKey,
		}

		// set col info
		c.Field = snaker.SnakeToCamel(c.ColumnName)
		c.Len, c.NilType, c.Type = SqParseType(args, c.DataType, c.IsNullable)

		// set primary key
		if c.IsPrimaryKey {
			tableTpl.PrimaryKey = c.ColumnName
			tableTpl.PrimaryKeyField = c.Field
			tableTpl.PrimaryKeyType = c.Type
		}

		// append col to template fields
		tableTpl.Fields = append(tableTpl.Fields, c)
	}

	return nil
}

// SqLoadForeignKeys loads foreign key relationships from the database.
func SqLoadForeignKeys(args *internal.ArgType, db *sql.DB, tableMap map[string]*internal.TableTemplate) (map[string]*internal.ForeignKeyTemplate, error) {
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
					FieldType:      f.Type,
					RefType:        tableMap[fk.RefTableName].Type,
				}

				// find ref field
				for _, f := range tableMap[fk.RefTableName].Fields {
					if f.ColumnName == fk.RefColumnName {
						fkMap[fk.ForeignKeyName].RefField = f.Field
						fkMap[fk.ForeignKeyName].RefFieldType = f.Type
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

// SqLoadIndexes loads indexes from the database.
func SqLoadIndexes(args *internal.ArgType, db *sql.DB, tableMap map[string]*internal.TableTemplate) (map[string]*internal.IndexTemplate, error) {
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

// SqParseQuery parses a sqlite query and generates a type for it.
func SqParseQuery(args *internal.ArgType, db *sql.DB) error {
	var err error

	// parse supplied query
	queryStr, params := args.ParseQuery("?", true)
	inspectStr, _ := args.ParseQuery("NULL", false)

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

	// create template for query type
	typeTpl := &internal.TableTemplate{
		Type:        args.QueryType,
		TableSchema: args.Schema,
		Fields:      []*models.Column{},
		Comment:     args.QueryTypeComment,
	}

	// if no query fields specified, then inspect with a view
	if args.QueryFields == "" {
		// create temporary view xoid
		xoid := "_xo_" + internal.GenRandomID()
		viewq := `CREATE TEMPORARY VIEW ` + xoid + ` AS ` + strings.Join(inspect, "\n")
		models.XOLog(viewq)
		_, err = db.Exec(viewq)
		if err != nil {
			return err
		}

		// load column information
		cols, err := models.SqColumnsByTable(db, xoid)
		if err != nil {
			return err
		}

		// process columns
		for _, sqC := range cols {
			// copy all the sq column fields
			c := &models.Column{
				FieldOrdinal: sqC.FieldOrdinal,
				ColumnName:   sqC.ColumnName,
				DataType:     sqC.DataType,
				IsNullable:   sqC.IsNullable,
				DefaultValue: sqC.DefaultValue.String,
				HasDefault:   sqC.DefaultValue.Valid,
				IsPrimaryKey: sqC.IsPrimaryKey,
			}

			c.Field = snaker.SnakeToCamel(c.ColumnName)
			c.Len, c.NilType, c.Type = SqParseType(args, c.DataType, false)
			typeTpl.Fields = append(typeTpl.Fields, c)
		}
	} else {
		for _, f := range strings.Split(args.QueryFields, ",") {
			f = strings.TrimSpace(f)
			colName := f
			colType := "string"

			i := strings.Index(f, " ")
			if i != -1 {
				colName = f[:i]
				colType = f[i+1:]
			}

			typeTpl.Fields = append(typeTpl.Fields, &models.Column{
				Field:      colName,
				Type:       colType,
				ColumnName: snaker.CamelToSnake(colName),
			})
		}
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
		Interpolate:   args.QueryInterpolate,
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

// SqParseType parse a postgres type into a Go type based on the column
// definition.
func SqParseType(args *internal.ArgType, dt string, nullable bool) (int, string, string) {
	precision := 0
	nilType := "nil"

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

	case "text":
		nilType = `""`
		typ = "string"
		if nullable {
			nilType = "sql.NullString{}"
			typ = "sql.NullString"
		}

	case "integer":
		nilType = "0"
		typ = args.Int32Type
		if nullable {
			nilType = "sql.NullInt64{}"
			typ = "sql.NullInt64"
		}

	case "numeric", "real":
		nilType = "0.0"
		typ = "float64"
		if nullable {
			nilType = "sql.NullFloat64{}"
			typ = "sql.NullFloat64"
		}

	case "blob":
		typ = "[]byte"

	default:
		panic(fmt.Errorf("unknown type %s", dt))
	}

	return precision, nilType, typ
}
