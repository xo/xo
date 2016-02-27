package internal

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/gedex/inflector"
	"github.com/serenize/snaker"

	"github.com/knq/xo/models"
)

// Loader is the common interface for database drivers that can generate code
// from a database schema.
type Loader interface {
	// IsSupported processes the passed url.URL, returning the sql.Open
	// driverName and dataSourceName and whether or not the Loader supports the
	// scheme in the url.
	IsSupported(*url.URL) (string, bool)

	// NthParam returns the 0-based Nth param for the Loader.
	NthParam(i int) string

	// Mask returns the mask.
	Mask() string

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

// TypeLoader provides a common Loader implementation used by the built in
// schema/query loaders.
type TypeLoader struct {
	Schemes         []string
	ProcessDSN      func(*url.URL, string) string
	ParamN          func(int) string
	MaskFunc        func() string
	ProcessRelkind  func(RelType) string
	Schema          func(*ArgType) (string, error)
	ParseType       func(*ArgType, string, bool) (int, string, string)
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

// IsSupported processes the passed url.URL, returning the sql.Open driverName
// and dataSourceName and whether or not the Loader supports the scheme in the
// url.
func (tl TypeLoader) IsSupported(u *url.URL) (string, bool) {
	uscheme := strings.ToLower(u.Scheme)
	protocol := "tcp"

	// check if +unix or whatever is in the scheme
	if strings.Contains(uscheme, "+") {
		p := strings.SplitN(uscheme, "+", 2)
		uscheme = p[0]
		protocol = p[1]
	}

	// determine if this type loader works for the url scheme
	var found bool
	for _, s := range tl.Schemes {
		if uscheme == strings.ToLower(s) {
			found = true
			break
		}
	}

	// bail if not found
	if !found {
		return "", false
	}

	// fix scheme
	u.Scheme = tl.Schemes[0]

	// process dsn if func is non-nil
	if tl.ProcessDSN != nil {
		return tl.ProcessDSN(u, protocol), true
	}

	return u.String(), true
}

// NthParam satisifies Loader's NthParam.
func (tl TypeLoader) NthParam(i int) string {
	if tl.ParamN != nil {
		return tl.ParamN(i)
	}

	return fmt.Sprintf("$%d", i+1)
}

// Mask returns the parameter mask
func (tl TypeLoader) Mask() string {
	if tl.MaskFunc != nil {
		return tl.MaskFunc()
	}

	return "$%d"
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
				Name: SnakeToCamel(strings.ToLower(c.ColumnName)),
				Col:  c,
			}
			f.Len, f.NilType, f.Type = tl.ParseType(args, c.DataType, false)
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
	err = args.ExecuteTemplate(QueryTypeTemplate, args.QueryType, "", typeTpl)
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
	err = args.ExecuteTemplate(QueryTemplate, args.QueryType, "", queryTpl)
	if err != nil {
		return err
	}

	return nil
}

// LoadSchema loads schema definitions.
func (tl TypeLoader) LoadSchema(args *ArgType) error {
	var err error

	// load enums
	_, err = tl.LoadEnums(args)
	if err != nil {
		return err
	}

	// load procs
	_, err = tl.LoadProcs(args)
	if err != nil {
		return err
	}

	// load tables
	tableMap, err := tl.LoadRelkind(args, Table)
	if err != nil {
		return err
	}

	// load foreign keys
	_, err = tl.LoadForeignKeys(args, tableMap)
	if err != nil {
		return err
	}

	// load indexes
	_, err = tl.LoadIndexes(args, tableMap)
	if err != nil {
		return err
	}

	return nil
}

// LoadEnums loads schema enums.
func (tl TypeLoader) LoadEnums(args *ArgType) (map[string]*Enum, error) {
	var err error

	// not supplied, so bail
	if tl.EnumList == nil {
		return nil, nil
	}

	// load enums
	enumList, err := tl.EnumList(args.DB, args.Schema)
	if err != nil {
		return nil, err
	}

	// process enums
	enumMap := map[string]*Enum{}
	for _, e := range enumList {
		enumTpl := &Enum{
			Name:   inflector.Singularize(SnakeToCamel(e.EnumName)),
			Schema: args.Schema,
			Values: []*EnumValue{},
			Enum:   e,
		}

		err = tl.LoadEnumValues(args, enumTpl)
		if err != nil {
			return nil, err
		}

		enumMap[enumTpl.Name] = enumTpl
		args.KnownTypeMap[enumTpl.Name] = true
	}

	// generate enum templates
	for _, e := range enumMap {
		err = args.ExecuteTemplate(EnumTemplate, e.Name, "", e)
		if err != nil {
			return nil, err
		}
	}

	return enumMap, nil
}

// LoadEnumValues loads schema enum values.
func (tl TypeLoader) LoadEnumValues(args *ArgType, enumTpl *Enum) error {
	var err error

	// load enum values
	enumValues, err := tl.EnumValueList(args.DB, args.Schema, enumTpl.Enum.EnumName)
	if err != nil {
		return err
	}

	// process enum values
	for _, ev := range enumValues {
		// chop off redundant enum name if applicable
		name := SnakeToCamel(strings.ToLower(ev.EnumValue))
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
func (tl TypeLoader) LoadProcs(args *ArgType) (map[string]*Proc, error) {
	var err error

	// not supplied, so bail
	if tl.ProcList == nil {
		return nil, nil
	}

	// load procs
	procList, err := tl.ProcList(args.DB, args.Schema)
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
			Name:   SnakeToCamel(strings.ToLower(name)),
			Schema: args.Schema,
			Params: []*Field{},
			Return: &Field{},
			Proc:   p,
		}

		// parse return type into template
		_, procTpl.Return.NilType, procTpl.Return.Type = tl.ParseType(args, p.ReturnType, false)

		// load proc parameters
		err = tl.LoadProcParams(args, procTpl)
		if err != nil {
			return nil, err
		}

		procMap[p.ProcName] = procTpl
	}

	// generate proc templates
	for _, p := range procMap {
		err = args.ExecuteTemplate(ProcTemplate, "sp_"+p.Name, "", p)
		if err != nil {
			return nil, err
		}
	}

	return procMap, nil
}

// LoadProcParams loads schema stored procedure parameters.
func (tl TypeLoader) LoadProcParams(args *ArgType, procTpl *Proc) error {
	var err error

	// load proc params
	paramList, err := tl.ProcParamList(args.DB, args.Schema, procTpl.Proc.ProcName)
	if err != nil {
		return err
	}

	// process params
	for i, p := range paramList {
		// TODO: some databases support named parameters in procs (MySQL)
		paramTpl := &Field{
			Name: fmt.Sprintf("v%d", i),
		}
		_, _, paramTpl.Type = tl.ParseType(args, strings.TrimSpace(p.ParamType), false)

		// add to proc params
		if procTpl.ProcParams != "" {
			procTpl.ProcParams = procTpl.ProcParams + ", "
		}
		procTpl.ProcParams = procTpl.ProcParams + p.ParamType

		procTpl.Params = append(procTpl.Params, paramTpl)
	}

	return nil
}

// LoadRelkind loads a schema table/view definition.
func (tl TypeLoader) LoadRelkind(args *ArgType, relType RelType) (map[string]*Type, error) {
	var err error

	// load tables
	tableList, err := tl.TableList(args.DB, args.Schema, tl.Relkind(relType))
	if err != nil {
		return nil, err
	}

	// tables
	tableMap := make(map[string]*Type)
	for _, ti := range tableList {
		// create template
		typeTpl := &Type{
			Name:    inflector.Singularize(SnakeToCamel(strings.ToLower(ti.TableName))),
			Schema:  args.Schema,
			RelType: relType,
			Fields:  []*Field{},
			Table:   ti,
		}

		// process columns
		err = tl.LoadColumns(args, typeTpl)
		if err != nil {
			return nil, err
		}

		tableMap[ti.TableName] = typeTpl
	}

	// generate table templates
	for _, t := range tableMap {
		err = args.ExecuteTemplate(TypeTemplate, t.Name, "", t)
		if err != nil {
			return nil, err
		}
	}

	return tableMap, nil
}

// LoadColumns loads schema table/view columns.
func (tl TypeLoader) LoadColumns(args *ArgType, typeTpl *Type) error {
	var err error

	// load columns
	columnList, err := tl.ColumnList(args.DB, args.Schema, typeTpl.Table.TableName)
	if err != nil {
		return err
	}

	// process columns
	for _, c := range columnList {
		// set col info
		f := &Field{
			Name: SnakeToCamel(strings.ToLower(c.ColumnName)),
			Col:  c,
		}
		f.Len, f.NilType, f.Type = tl.ParseType(args, c.DataType, !c.NotNull)

		// set primary key
		if c.IsPrimaryKey {
			typeTpl.PrimaryKey = f
		}

		// append col to template fields
		typeTpl.Fields = append(typeTpl.Fields, f)
	}

	return nil
}

// LoadForeignKeys loads foreign keys.
func (tl TypeLoader) LoadForeignKeys(args *ArgType, tableMap map[string]*Type) (map[string]*ForeignKey, error) {
	var err error

	fkMap := map[string]*ForeignKey{}
	for _, t := range tableMap {
		// load keys per table
		err = tl.LoadTableForeignKeys(args, tableMap, t, fkMap)
		if err != nil {
			return nil, err
		}
	}

	// generate templates
	for _, fk := range fkMap {
		err = args.ExecuteTemplate(ForeignKeyTemplate, fk.Type.Name, fk.ForeignKey.ForeignKeyName, fk)
		if err != nil {
			return nil, err
		}
	}

	return fkMap, nil
}

// LoadTableForeignKeys loads schema foreign key definitions per table.
func (tl TypeLoader) LoadTableForeignKeys(args *ArgType, tableMap map[string]*Type, typeTpl *Type, fkMap map[string]*ForeignKey) error {
	var err error

	// load foreign keys
	foreignKeyList, err := tl.ForeignKeyList(args.DB, args.Schema, typeTpl.Table.TableName)
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

	refColLoop:
		// find ref column
		for _, f := range refTpl.Fields {
			if f.Col.ColumnName == fk.RefColumnName {
				refCol = f
				break refColLoop
			}
		}

		// no ref col, but have ref tpl, so use primary key
		if refTpl != nil && refCol == nil {
			refCol = refTpl.PrimaryKey
		}

		// check everything was found
		if col == nil || refTpl == nil || refCol == nil {
			return errors.New("could not find col, refTpl, or refCol")
		}

		// foreign key name
		if fk.ForeignKeyName == "" {
			fk.ForeignKeyName = typeTpl.Table.TableName + "_" + col.Col.ColumnName + "_fkey"
		}

		// create foreign key template
		fkMap[fk.ForeignKeyName] = &ForeignKey{
			Schema:     args.Schema,
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
func (tl TypeLoader) LoadIndexes(args *ArgType, tableMap map[string]*Type) (map[string]*Index, error) {
	var err error

	ixMap := map[string]*Index{}
	for _, t := range tableMap {
		// load table indexes
		err = tl.LoadTableIndexes(args, t, ixMap)
		if err != nil {
			return nil, err
		}
	}

	// generate templates
	for _, ix := range ixMap {
		err = args.ExecuteTemplate(IndexTemplate, ix.Type.Name, ix.Index.IndexName, ix)
		if err != nil {
			return nil, err
		}
	}

	return ixMap, nil
}

// LoadTableIndexes loads schema index definitions per table.
func (tl TypeLoader) LoadTableIndexes(args *ArgType, typeTpl *Type, ixMap map[string]*Index) error {
	var err error
	var priIxLoaded bool

	// load indexes
	indexList, err := tl.IndexList(args.DB, args.Schema, typeTpl.Table.TableName)
	if err != nil {
		return err
	}

	// process indexes
	for _, ix := range indexList {
		ixName := ix.IndexName

		// save that the primary key index was loaded
		priIxLoaded = priIxLoaded || ix.IsPrimary

		// chop off tablename_
		if strings.HasPrefix(ixName, typeTpl.Table.TableName+"_") {
			ixName = ixName[len(typeTpl.Table.TableName)+1:]
		}

		// chop off _ix, _idx, or _index
		switch {
		case strings.HasSuffix(ixName, "_ix"):
			ixName = ixName[:len(ixName)-len("_ix")]
		case strings.HasSuffix(ixName, "_idx"):
			ixName = ixName[:len(ixName)-len("_idx")]
		case strings.HasSuffix(ixName, "_index"):
			ixName = ixName[:len(ixName)-len("_index")]
		}

		// determine the type name
		typeName := typeTpl.Name
		if !ix.IsUnique {
			typeName = inflector.Pluralize(typeTpl.Name)
		}

		// create index template
		ixTpl := &Index{
			Name:     SnakeToCamel(strings.ToLower(ixName)),
			TypeName: typeName,
			Schema:   args.Schema,
			Type:     typeTpl,
			Fields:   []*Field{},
			Index:    ix,
		}

		// load index columns
		err = tl.LoadIndexColumns(args, ixTpl)
		if err != nil {
			return err
		}

		ixMap[ix.IndexName] = ixTpl
	}

	// if no primary key index loaded, but a primary key column was defined in
	// the type, then create the definition here. this is needed for sqlite, as
	// sqlite doesn't define primary keys in its index list
	if args.LoaderType != "ora" && !priIxLoaded && typeTpl.PrimaryKey != nil {
		ixName := typeTpl.Table.TableName + "_" + typeTpl.PrimaryKey.Col.ColumnName + "_pkey"
		ixMap[ixName] = &Index{
			Name:     SnakeToCamel(strings.ToLower(ixName)),
			TypeName: typeTpl.Name,
			Schema:   args.Schema,
			Type:     typeTpl,
			Fields:   []*Field{typeTpl.PrimaryKey},
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
func (tl TypeLoader) LoadIndexColumns(args *ArgType, ixTpl *Index) error {
	var err error

	// load index columns
	indexCols, err := tl.IndexColumnList(args.DB, args.Schema, ixTpl.Type.Table.TableName, ixTpl.Index.IndexName)
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
