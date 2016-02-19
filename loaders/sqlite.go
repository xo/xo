package loaders

import (
	"database/sql"
	"errors"
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

	// load table info
	tableInfo, err := models.SqTableinfosByRelkind(db, "table")
	if err != nil {
		return err
	}

	// load tables
	tableMap, err := SqLoadTables(args, db, tableInfo)
	if err != nil {
		return err
	}

	// load foreign relationships
	_, err = SqLoadForeignKeys(args, db, tableInfo, tableMap)
	if err != nil {
		return err
	}

	// load idx
	_, err = SqLoadIndexes(args, db, tableInfo, tableMap)
	if err != nil {
		return err
	}

	return nil
}

// SqLoadTables loads the table definitions from the database, writing the
// resulting templates to args.TypeMap and returning the created TableTemplates.
func SqLoadTables(args *internal.ArgType, db *sql.DB, tableInfo []*models.SqTableinfo) (map[string]*internal.TableTemplate, error) {
	var err error

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
		err = SqLoadTableColumns(args, db, tableTpl)
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

// SqLoadTableColumns loads the column definition from the database, writing the
// adding the fields to the supplied table template.
func SqLoadTableColumns(args *internal.ArgType, db *sql.DB, tableTpl *internal.TableTemplate) error {
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
			IsNullable:   !sqC.NotNull,
			DefaultValue: sqC.DefaultValue.String,
			HasDefault:   sqC.DefaultValue.Valid,
			IsPrimaryKey: sqC.IsPrimaryKey,
		}

		// set col info
		c.Field = snaker.SnakeToCamel(c.ColumnName)
		fmt.Printf(">>> %s -- not null: %t // %t\n", sqC.ColumnName, sqC.NotNull, c.IsNullable)
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
func SqLoadForeignKeys(args *internal.ArgType, db *sql.DB, tableInfo []*models.SqTableinfo, tableMap map[string]*internal.TableTemplate) (map[string]*internal.ForeignKeyTemplate, error) {
	var err error

	fkMap := map[string]*internal.ForeignKeyTemplate{}
	for _, ti := range tableInfo {
		// find table
		var tableTpl *internal.TableTemplate
		for _, t := range tableMap {
			if ti.TableName == t.TableName {
				tableTpl = t
				break
			}
		}

		// sanity check
		if tableTpl == nil {
			return nil, errors.New("could not find tableTpl")
		}

		// load keys per table
		err = SqLoadTableForeignKeys(args, db, tableMap, tableTpl, fkMap)
		if err != nil {
			return nil, err
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

// SqLoadTableForeignKeys loads the foreign keys for the specified table.
func SqLoadTableForeignKeys(args *internal.ArgType, db *sql.DB, tableMap map[string]*internal.TableTemplate, tableTpl *internal.TableTemplate, fkMap map[string]*internal.ForeignKeyTemplate) error {
	// load foreign keys per table
	fkeys, err := models.SqForeignKeysByTable(db, tableTpl.TableName)
	if err != nil {
		return err
	}

	// loop over foreign keys for table
	for _, fk := range fkeys {
		var refTpl *internal.TableTemplate
		var col, refCol *models.Column

	colLoop:
		// find column
		for _, c := range tableTpl.Fields {
			if c.ColumnName == fk.ColumnName {
				col = c
				break colLoop
			}
		}

	refTplLoop:
		// find ref table
		for _, t := range tableMap {
			if t.TableName == fk.RefTableName {
				refTpl = t
				break refTplLoop
			}
		}

	refColLoop:
		// find ref column
		for _, c := range refTpl.Fields {
			if c.ColumnName == fk.RefColumnName {
				refCol = c
				break refColLoop
			}
		}

		// sanity check
		if col == nil || refTpl == nil || refCol == nil {
			return errors.New("could not find col, refTpl, or refCol")
		}

		// foreign key name
		fkName := tableTpl.TableName + "_" + col.ColumnName + "_fkey"

		// set back in column
		col.IsForeignKey = true
		col.ForeignIndexName = fkName

		// create foreign key template
		fkMap[fkName] = &internal.ForeignKeyTemplate{
			Type:           tableTpl.Type,
			ForeignKeyName: fkName,
			ColumnName:     fk.ColumnName,
			Field:          col.Field,
			FieldType:      col.Type,
			RefType:        refTpl.Type,
			RefField:       refCol.Field,
			RefFieldType:   refCol.Type,
		}
	}

	return nil
}

// SqLoadIndexes loads indexes from the database.
func SqLoadIndexes(args *internal.ArgType, db *sql.DB, tableInfo []*models.SqTableinfo, tableMap map[string]*internal.TableTemplate) (map[string]*internal.IndexTemplate, error) {
	var err error

	idxMap := map[string]*internal.IndexTemplate{}
	for _, ti := range tableInfo {
		// find table
		var tableTpl *internal.TableTemplate
		for _, t := range tableMap {
			if ti.TableName == t.TableName {
				tableTpl = t
				break
			}
		}

		// sanity check
		if tableTpl == nil {
			return nil, errors.New("could not find tableTpl")
		}

		// load keys per table
		err = SqLoadTableIndexes(args, db, tableTpl, idxMap)
		if err != nil {
			return nil, err
		}
	}

	// sort idx keys
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

// SqLoadTableIndexes loads indexes for the specified table.
func SqLoadTableIndexes(args *internal.ArgType, db *sql.DB, tableTpl *internal.TableTemplate, idxMap map[string]*internal.IndexTemplate) error {
	var err error

	// load primary key
	//SqLoadTablePrimaryKeyIndex(args, db, tableTpl, idxMap)
	if tableTpl.PrimaryKey != "" {
		// find primary field
		var f *models.Column
		for _, c := range tableTpl.Fields {
			if c.ColumnName == tableTpl.PrimaryKey {
				f = c
				break
			}
		}

		pkName := tableTpl.TableName + "_" + tableTpl.PrimaryKey + "_pkey"
		idxMap[pkName] = &internal.IndexTemplate{
			Type:        tableTpl.Type,
			Name:        tableTpl.PrimaryKeyField,
			TableSchema: tableTpl.TableSchema,
			TableName:   tableTpl.TableName,
			IndexName:   pkName,
			IsUnique:    true,
			Fields:      []*models.Column{f},
			Table:       tableTpl,
		}
	}

	// load indexes
	indexes, err := models.SqIndicesByTable(db, tableTpl.TableName)
	if err != nil {
		return err
	}

	// loop each index
	for _, ix := range indexes {
		idxName := ix.IndexName

		// chop off tablename_
		if strings.HasPrefix(idxName, tableTpl.TableName+"_") {
			idxName = idxName[len(tableTpl.TableName)+1:]
		}

		// chop off _idx or _index
		switch {
		case strings.HasSuffix(idxName, "_idx"):
			idxName = idxName[:len(idxName)-len("_idx")]
		case strings.HasSuffix(idxName, "_index"):
			idxName = idxName[:len(idxName)-len("_index")]
		}

		// create index template
		idxTpl := &internal.IndexTemplate{
			Type:        tableTpl.Type,
			Name:        snaker.SnakeToCamel(idxName),
			Plural:      inflector.Pluralize(tableTpl.Type),
			TableSchema: tableTpl.TableSchema,
			TableName:   tableTpl.TableName,
			IndexName:   ix.IndexName,
			IsUnique:    ix.IsUnique,
			Table:       tableTpl,
			Fields:      []*models.Column{},
		}

		// load index info
		err = SqLoadIndexInfo(args, db, idxTpl)
		if err != nil {
			return err
		}

		idxMap[ix.IndexName] = idxTpl
	}

	return nil
}

// SqLoadIndexInfo loads index info for the specified table and index.
func SqLoadIndexInfo(args *internal.ArgType, db *sql.DB, idxTpl *internal.IndexTemplate) error {
	var err error

	// load index info
	idxInfo, err := models.SqIndexinfosByIndex(db, idxTpl.IndexName)
	if err != nil {
		return err
	}

	// sort by seqno
	infoList := SqIndexinfoList(idxInfo)
	sort.Sort(infoList)

	// build fields for index
	for _, ii := range infoList {
		// find field
		for _, c := range idxTpl.Table.Fields {
			if ii.ColumnName == c.ColumnName {
				idxTpl.Fields = append(idxTpl.Fields, c)
				break
			}
		}
	}

	return nil
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
				IsNullable:   !sqC.NotNull,
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
	}

	return precision, nilType, typ
}

// sort interface for SqIndexinfo
type SqIndexinfoList []*models.SqIndexinfo

func (s SqIndexinfoList) Len() int {
	return len(s)
}

func (s SqIndexinfoList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s SqIndexinfoList) Less(i, j int) bool {
	return s[i].SeqNo < s[j].SeqNo
}
