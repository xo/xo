package internal

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/gedex/inflector"
	"github.com/knq/snaker"

	"github.com/xo/xo/models"
)

// Loader is the common interface for database drivers that can generate code
// from a database schema.
type Loader interface {
	// NthParam returns the 0-based Nth param for the Loader.
	NthParam(i int) string

	// Mask returns the mask.
	Mask() string

	// Escape escapes the passed identifier based on its EscType.
	Escape(EscType, string) string

	// Relkind returns the schema's relkind identifier (ie, TABLE, VIEW, BASE TABLE, etc).
	Relkind(RelType) string

	// SchemaName loads the active schema name from the database if not provided on the cli.
	SchemaName(*ArgType) (string, error)

	// ParseQuery parses the ArgType.Query and writes any necessary type(s) to
	// the ArgType from the opened database handle.
	ParseQuery(*ArgType) error

	// LoadSchema loads the ArgType.Schema from the opened database handle,
	// writing any generated types to ArgType.
	LoadSchema(*ArgType) error
}

// SchemaLoaders are the available schema loaders.
var SchemaLoaders = map[string]Loader{}

// schema_name -> { enum_name -> Enum }
type schemaEnumMap map[string]map[string]*Enum

// schema_name -> { proc_name -> Proc }
type schemaProcMap map[string]map[string]*Proc

// schema_name -> { type_name -> Type }
type schemaTypeMap map[string]map[string]*Type

// schema_name -> { foreign_key_name -> ForeignKey }
type schemaFkMap map[string]map[string]*ForeignKey

// schema_name -> { index_name -> Index }
type schemaIndexMap map[string]map[string]*Index

// TypeLoader provides a common Loader implementation used by the built in
// schema/query loaders.
type TypeLoader struct {
	ParamN          func(int) string
	MaskFunc        func() string
	Esc             map[EscType]func(string) string
	ProcessRelkind  func(RelType) string
	Schema          func(*ArgType) (string, error)
	ParseType       func(*ArgType, string, string, bool) (int, string, string)
	EnumList        func(models.XODB, string) ([]*models.Enum, error)
	EnumValueList   func(models.XODB, string, string) ([]*models.EnumValue, error)
	ProcList        func(models.XODB, string) ([]*models.Proc, error)
	ProcParamList   func(models.XODB, string, string) ([]*models.ProcParam, error)
	TableList       func(models.XODB, string, string) ([]*models.Table, error)
	ColumnList      func(models.XODB, string, string) ([]*models.Column, error)
	ForeignKeyList  func(models.XODB, string, string) ([]*models.ForeignKey, error)
	IndexList       func(models.XODB, string, string) ([]*models.Index, error)
	IndexColumnList func(models.XODB, string, string, string) ([]*models.IndexColumn, error)
	QueryStrip      func([]string, []string)
	QueryColumnList func(*ArgType, []string) ([]*models.Column, error)
}

// NthParam satisifies Loader's NthParam.
func (tl TypeLoader) NthParam(i int) string {
	if tl.ParamN != nil {
		return tl.ParamN(i)
	}

	return fmt.Sprintf("$%d", i+1)
}

// Mask returns the parameter mask.
func (tl TypeLoader) Mask() string {
	if tl.MaskFunc != nil {
		return tl.MaskFunc()
	}

	return "$%d"
}

// Escape escapes the provided identifier based on the EscType.
func (tl TypeLoader) Escape(typ EscType, s string) string {
	if e, ok := tl.Esc[typ]; ok && e != nil {
		return e(s)
	}

	return `"` + s + `"`
}

// Relkind satisfies Loader's Relkind.
func (tl TypeLoader) Relkind(rt RelType) string {
	if tl.ProcessRelkind != nil {
		return tl.ProcessRelkind(rt)
	}

	return rt.String()
}

// SchemaName returns the active schema name.
func (tl TypeLoader) SchemaName(args *ArgType) (string, error) {
	if tl.Schema != nil {
		return tl.Schema(args)
	}
	return "", nil
}

// ParseQuery satisfies Loader's ParseQuery.
func (tl TypeLoader) ParseQuery(args *ArgType) error {
	var err error

	// parse supplied query
	queryStr, params := args.ParseQuery(tl.Mask(), true)
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

	// query strip
	if args.QueryStrip && tl.QueryStrip != nil {
		tl.QueryStrip(query, queryComments)
	}

	// create template for query type
	typeTpl := &Type{
		Name:    args.QueryType,
		RelType: Table,
		Fields:  []*Field{},
		Table: &models.Table{
			TableName: "[custom " + strings.ToLower(snaker.CamelToSnake(args.QueryType)) + "]",
		},
		Comment: args.QueryTypeComment,
	}

	if args.QueryFields == "" {
		// if no query fields specified, then pass to inspector
		colList, err := tl.QueryColumnList(args, inspect)
		if err != nil {
			return err
		}

		// process columns
		for _, c := range colList {
			f := &Field{
				Name: snaker.SnakeToCamelIdentifier(c.ColumnName),
				Col:  c,
			}
			// FIXME
			schema := args.Schemas[0]
			f.Len, f.NilType, f.Type = tl.ParseType(args, schema.Name, c.DataType, args.QueryAllowNulls && !c.NotNull)
			typeTpl.Fields = append(typeTpl.Fields, f)
		}
	} else {
		// extract fields from query fields
		for _, qf := range strings.Split(args.QueryFields, ",") {
			qf = strings.TrimSpace(qf)
			colName := qf
			colType := "string"

			i := strings.Index(qf, " ")
			if i != -1 {
				colName = qf[:i]
				colType = qf[i+1:]
			}

			typeTpl.Fields = append(typeTpl.Fields, &Field{
				Name: colName,
				Type: colType,
				Col: &models.Column{
					ColumnName: snaker.CamelToSnake(colName),
				},
			})
		}
	}

	// generate query type template
	err = args.ExecuteTemplate(QueryTypeTemplate, args.Package, args.QueryType, "", typeTpl)
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
	queryTpl := &Query{
		Name:          funcName,
		Query:         query,
		QueryComments: queryComments,
		QueryParams:   params,
		OnlyOne:       args.QueryOnlyOne,
		Interpolate:   args.QueryInterpolate,
		Type:          typeTpl,
		Comment:       args.QueryFuncComment,
	}

	// generate template
	err = args.ExecuteTemplate(QueryTemplate, args.Package, args.QueryType, "", queryTpl)
	if err != nil {
		return err
	}

	return nil
}

// LoadSchema loads schema definitions.
func (tl TypeLoader) LoadSchema(args *ArgType) error {
	var err error

	// load enums
	_, err = tl.loadEnums(args)
	if err != nil {
		return err
	}

	// load procs
	_, err = tl.loadProcs(args)
	if err != nil {
		return err
	}

	// load tables
	allTableMap, err := tl.loadRelkind(args, Table)
	if err != nil {
		return err
	}

	// load views
	allViewMap, err := tl.loadRelkind(args, View)
	if err != nil {
		return err
	}

	// merge views with the tableMap
	for schema, viewMap := range allViewMap {
		tableMap, exists := allTableMap[schema]
		if !exists {
			allTableMap[schema] = viewMap
			continue
		}

		for k, v := range viewMap {
			tableMap[k] = v
		}
	}

	// load foreign keys
	_, err = tl.loadForeignKeys(args, allTableMap)
	if err != nil {
		return err
	}

	// load indexes
	_, err = tl.loadIndexes(args, allTableMap)
	if err != nil {
		return err
	}

	return nil
}

// LoadEnums loads schema enums.
func (tl TypeLoader) loadEnums(args *ArgType) (schemaEnumMap, error) {
	// not supplied, so bail
	if tl.EnumList == nil {
		return nil, nil
	}

	// map of schema_name -> {enum_name -> Enum}
	allEnumMap := make(map[string]map[string]*Enum)

	// load enums for each schema
	for _, dbSchema := range args.Schemas {
		schema := dbSchema.Name
		enumList, err := tl.EnumList(args.DB, schema)
		if err != nil {
			return nil, err
		}

		// process enums
		enumMap := map[string]*Enum{}
		for _, e := range enumList {
			enumTpl := &Enum{
				Name:              SingularizeIdentifier(e.EnumName),
				Schema:            schema,
				Values:            []*EnumValue{},
				Enum:              e,
				ReverseConstNames: args.UseReversedEnumConstNames,
			}

			err = tl.loadEnumValues(args, schema, enumTpl)
			if err != nil {
				return nil, err
			}

			enumMap[enumTpl.Name] = enumTpl
			args.KnownTypeMap[enumTpl.Name] = true

		}

		// generate enum templates
		for _, e := range enumMap {
			err = args.ExecuteTemplate(EnumTemplate, dbSchema.Name, e.Name, "", e)
			if err != nil {
				return nil, err
			}
		}
		allEnumMap[schema] = enumMap
	}

	return allEnumMap, nil
}

// loadEnumValues loads schema enum values.
func (tl TypeLoader) loadEnumValues(args *ArgType, schema string, enumTpl *Enum) error {
	var err error

	// load enum values
	enumValues, err := tl.EnumValueList(args.DB, schema, enumTpl.Enum.EnumName)
	if err != nil {
		return err
	}

	// process enum values
	for _, ev := range enumValues {
		// chop off redundant enum name if applicable
		name := snaker.SnakeToCamelIdentifier(ev.EnumValue)
		if strings.HasSuffix(strings.ToLower(name), strings.ToLower(enumTpl.Name)) {
			n := name[:len(name)-len(enumTpl.Name)]
			if len(n) > 0 {
				name = n
			}
		}

		enumTpl.Values = append(enumTpl.Values, &EnumValue{
			Name: name,
			Val:  ev,
		})
	}

	return nil
}

// LoadProcs loads schema stored procedures definitions.
func (tl TypeLoader) loadProcs(args *ArgType) (schemaProcMap, error) {
	// not supplied, so bail
	if tl.ProcList == nil {
		return nil, nil
	}

	allProcMap := make(map[string]map[string]*Proc)

	for _, dbSchema := range args.Schemas {
		schema := dbSchema.Name
		// load procs
		procList, err := tl.ProcList(args.DB, schema)
		if err != nil {
			return nil, err
		}

		// process procs
		procMap := map[string]*Proc{}
		for _, p := range procList {
			// fix the name if it starts with one or more underscores
			name := p.ProcName
			for strings.HasPrefix(name, "_") {
				name = name[1:]
			}

			// create template
			procTpl := &Proc{
				Name:   snaker.SnakeToCamelIdentifier(name),
				Schema: schema,
				Params: []*Field{},
				Return: &Field{},
				Proc:   p,
			}

			// parse return type into template
			// TODO: fix this so that nullable types can be returned
			_, procTpl.Return.NilType, procTpl.Return.Type = tl.ParseType(args, schema, p.ReturnType, false)

			// load proc parameters
			err = tl.loadProcParams(args, schema, procTpl)
			if err != nil {
				return nil, err
			}

			procMap[p.ProcName] = procTpl
		}

		// generate proc templates
		for _, p := range procMap {
			err = args.ExecuteTemplate(ProcTemplate, dbSchema.Name, "sp_"+p.Name, "", p)
			if err != nil {
				return nil, err
			}
		}

		allProcMap[schema] = procMap
	}

	return allProcMap, nil
}

// loadProcParams loads schema stored procedure parameters.
func (tl TypeLoader) loadProcParams(args *ArgType, schema string, procTpl *Proc) error {
	var err error

	// load proc params
	paramList, err := tl.ProcParamList(args.DB, schema, procTpl.Proc.ProcName)
	if err != nil {
		return err
	}

	// process params
	for i, p := range paramList {
		// TODO: some databases support named parameters in procs (MySQL)
		paramTpl := &Field{
			Name: fmt.Sprintf("v%d", i),
		}

		// TODO: fix this so that nullable types can be used as parameters
		_, _, paramTpl.Type = tl.ParseType(args, schema, strings.TrimSpace(p.ParamType), false)

		// add to proc params
		if procTpl.ProcParams != "" {
			procTpl.ProcParams = procTpl.ProcParams + ", "
		}
		procTpl.ProcParams = procTpl.ProcParams + p.ParamType

		procTpl.Params = append(procTpl.Params, paramTpl)
	}

	return nil
}

// loadRelkind loads a schema table/view definition.
func (tl TypeLoader) loadRelkind(args *ArgType, relType RelType) (schemaTypeMap, error) {
	allTableMap := make(map[string]map[string]*Type)

	for _, dbSchema := range args.Schemas {
		schema := dbSchema.Name
		// load tables
		tableList, err := tl.TableList(args.DB, schema, tl.Relkind(relType))
		if err != nil {
			return nil, err
		}

		// tables
		tableMap := make(map[string]*Type)
		for _, ti := range tableList {
			// create template
			typeTpl := &Type{
				Name:    SingularizeIdentifier(ti.TableName),
				Schema:  schema,
				RelType: relType,
				Fields:  []*Field{},
				Table:   ti,
			}

			// process columns
			if err = tl.loadColumns(args, schema, typeTpl); err != nil {
				return nil, err
			}

			tableMap[ti.TableName] = typeTpl
		}

		// generate table templates
		for _, t := range tableMap {
			err = args.ExecuteTemplate(TypeTemplate, dbSchema.Name, t.Name, "", t)
			if err != nil {
				return nil, err
			}
		}

		allTableMap[schema] = tableMap
	}

	return allTableMap, nil
}

// loadColumns loads schema table/view columns.
func (tl TypeLoader) loadColumns(args *ArgType, schema string, typeTpl *Type) error {
	var err error

	// load columns
	columnList, err := tl.ColumnList(args.DB, schema, typeTpl.Table.TableName)
	if err != nil {
		return err
	}

	// process columns
	for _, c := range columnList {
		ignore := false

		for _, ignoreField := range args.IgnoreFields {
			if ignoreField == c.ColumnName {
				// Skip adding this field if user has specified they are not
				// interested.
				//
				// This could be useful for fields which are managed by the
				// database (e.g. automatically updated timestamps) instead of
				// via Go code.
				ignore = true
			}
		}

		if ignore {
			continue
		}

		// set col info
		f := &Field{
			Name: snaker.SnakeToCamelIdentifier(c.ColumnName),
			Col:  c,
		}
		f.Len, f.NilType, f.Type = tl.ParseType(args, schema, c.DataType, !c.NotNull)

		// set primary key
		if c.IsPrimaryKey {
			typeTpl.PrimaryKeyFields = append(typeTpl.PrimaryKeyFields, f)
			// This is retained for backward compatibility in the templates.
			typeTpl.PrimaryKey = f
		}

		// append col to template fields
		typeTpl.Fields = append(typeTpl.Fields, f)
	}

	return nil
}

// LoadForeignKeys loads foreign keys.
func (tl TypeLoader) loadForeignKeys(args *ArgType, allTableMap schemaTypeMap) (schemaFkMap, error) {
	var err error

	allFkMap := make(schemaFkMap)
	for schema, tableMap := range allTableMap {
		fkMap := map[string]*ForeignKey{}
		for _, t := range tableMap {
			// load keys per table
			err = tl.loadTableForeignKeys(args, schema, tableMap, t, fkMap)
			if err != nil {
				return nil, err
			}
		}

		// determine foreign key names
		for _, fk := range fkMap {
			fk.Name = args.ForeignKeyName(fkMap, fk)
		}

		// generate templates
		for _, fk := range fkMap {
			err = args.ExecuteTemplate(ForeignKeyTemplate, schema, fk.Type.Name, fk.ForeignKey.ForeignKeyName, fk)
			if err != nil {
				return nil, err
			}
		}
		allFkMap[schema] = fkMap
	}

	return allFkMap, nil
}

// LoadTableForeignKeys loads schema foreign key definitions per table.
func (tl TypeLoader) loadTableForeignKeys(args *ArgType, schema string, tableMap map[string]*Type, typeTpl *Type, fkMap map[string]*ForeignKey) error {
	var err error

	// load foreign keys
	foreignKeyList, err := tl.ForeignKeyList(args.DB, schema, typeTpl.Table.TableName)
	if err != nil {
		return err
	}

	// loop over foreign keys for table
	for _, fk := range foreignKeyList {
		var refTpl *Type
		var col, refCol *Field

	colLoop:
		// find column
		for _, f := range typeTpl.Fields {
			if f.Col.ColumnName == fk.ColumnName {
				col = f
				break colLoop
			}
		}

	refTplLoop:
		// find ref table
		for _, t := range tableMap {
			if t.Table.TableName == fk.RefTableName {
				refTpl = t
				break refTplLoop
			}
		}

		if refTpl == nil {
			log.Printf("can't create foreign key template for '%v' reference to table '%v', skipping it",
				fk.ForeignKeyName, fk.RefTableSchema+"."+fk.RefTableName)
			continue
		}

		// find ref column
	refColLoop:
		for _, f := range refTpl.Fields {
			if f.Col.ColumnName == fk.RefColumnName {
				refCol = f
				break refColLoop
			}
		}

		// no ref col, but have ref tpl, so use primary key
		if refCol == nil {
			refCol = refTpl.PrimaryKey
		}

		// check everything was found
		if col == nil || refCol == nil {
			return errors.New("could not find col, refTpl, or refCol")
		}

		// foreign key name
		if fk.ForeignKeyName == "" {
			fk.ForeignKeyName = typeTpl.Table.TableName + "_" + col.Col.ColumnName + "_fkey"
		}

		// create foreign key template
		fkMap[fk.ForeignKeyName] = &ForeignKey{
			Schema:     schema,
			Type:       typeTpl,
			Field:      col,
			RefType:    refTpl,
			RefField:   refCol,
			ForeignKey: fk,
		}
	}

	return nil
}

// LoadIndexes loads schema index definitions.
func (tl TypeLoader) loadIndexes(args *ArgType, allTableMap schemaTypeMap) (schemaIndexMap, error) {
	var err error

	allIxMap := make(schemaIndexMap)
	for schema, tableMap := range allTableMap {
		ixMap := map[string]*Index{}
		for _, t := range tableMap {
			// load table indexes
			err = tl.LoadTableIndexes(args, schema, t, ixMap)
			if err != nil {
				return nil, err
			}
		}

		// generate templates
		for _, ix := range ixMap {
			err = args.ExecuteTemplate(IndexTemplate, schema, ix.Type.Name, ix.Index.IndexName, ix)
			if err != nil {
				return nil, err
			}
		}
		allIxMap[schema] = ixMap
	}

	return allIxMap, nil
}

// LoadTableIndexes loads schema index definitions per table.
func (tl TypeLoader) LoadTableIndexes(args *ArgType, schema string, typeTpl *Type, ixMap map[string]*Index) error {
	var err error
	var priIxLoaded bool

	// load indexes
	indexList, err := tl.IndexList(args.DB, schema, typeTpl.Table.TableName)
	if err != nil {
		return err
	}

	// process indexes
	for _, ix := range indexList {
		// save whether or not the primary key index was processed
		priIxLoaded = priIxLoaded || ix.IsPrimary || (ix.Origin == "pk")

		// create index template
		ixTpl := &Index{
			Schema: schema,
			Type:   typeTpl,
			Fields: []*Field{},
			Index:  ix,
		}

		// load index columns
		err = tl.loadIndexColumns(args, schema, ixTpl)
		if err != nil {
			return err
		}

		// build func name
		args.BuildIndexFuncName(ixTpl)

		ixMap[typeTpl.Table.TableName+"_"+ix.IndexName] = ixTpl
	}

	// search for primary key if it was skipped being set in the type
	pk := typeTpl.PrimaryKey
	if pk == nil {
		for _, f := range typeTpl.Fields {
			if f.Col.IsPrimaryKey {
				pk = f
				break
			}
		}
	}

	// if no primary key index loaded, but a primary key column was defined in
	// the type, then create the definition here. this is needed for sqlite, as
	// sqlite doesn't define primary keys in its index list
	if args.LoaderType != "ora" && !priIxLoaded && pk != nil {
		ixName := typeTpl.Table.TableName + "_" + pk.Col.ColumnName + "_pkey"
		ixMap[ixName] = &Index{
			FuncName: typeTpl.Name + "By" + pk.Name,
			Schema:   schema,
			Type:     typeTpl,
			Fields:   []*Field{pk},
			Index: &models.Index{
				IndexName: ixName,
				IsUnique:  true,
				IsPrimary: true,
			},
		}
	}

	return nil
}

// LoadIndexColumns loads the index column information.
func (tl TypeLoader) loadIndexColumns(args *ArgType, schema string, ixTpl *Index) error {
	var err error

	// load index columns
	indexCols, err := tl.IndexColumnList(args.DB, schema, ixTpl.Type.Table.TableName, ixTpl.Index.IndexName)
	if err != nil {
		return err
	}

	// process index columns
	for _, ic := range indexCols {
		var field *Field

	fieldLoop:
		// find field
		for _, f := range ixTpl.Type.Fields {
			if f.Col.ColumnName == ic.ColumnName {
				field = f
				break fieldLoop
			}
		}

		if field == nil {
			continue
		}

		ixTpl.Fields = append(ixTpl.Fields, field)
	}

	return nil
}
