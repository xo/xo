package cmd

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/gedex/inflector"
	xo "github.com/xo/xo/types"
)

// BuildSchema builds a schema.
func BuildSchema(ctx context.Context, args *Args, dest *xo.XO) error {
	// driver info
	_, l, schema := DbLoaderSchema(ctx)
	s := xo.Schema{
		Driver: l.Driver,
		Name:   schema,
	}
	var err error
	// load enums, procs, tables, views
	if s.Enums, err = LoadEnums(ctx, args); err != nil {
		return err
	}
	if s.Procs, err = LoadProcs(ctx, args); err != nil {
		return err
	}
	if s.Tables, err = LoadTables(ctx, args, "table"); err != nil {
		return err
	}
	if s.Views, err = LoadTables(ctx, args, "view"); err != nil {
		return err
	}
	// emit
	dest.Emit(s)
	return nil
}

// LoadEnums loads enums.
func LoadEnums(ctx context.Context, args *Args) ([]xo.Enum, error) {
	db, l, schema := DbLoaderSchema(ctx)
	// not supplied, so bail
	if l.Enums == nil {
		return nil, nil
	}
	// load enums
	enumNames, err := l.Enums(ctx, db, schema)
	if err != nil {
		return nil, err
	}
	sort.Slice(enumNames, func(i, j int) bool {
		return enumNames[i].EnumName < enumNames[j].EnumName
	})
	// process enums
	var enums []xo.Enum
	for _, name := range enumNames {
		enum := &xo.Enum{
			Name: name.EnumName,
		}
		if err := LoadEnumValues(ctx, args, enum); err != nil {
			return nil, err
		}
		enums = append(enums, *enum)
	}
	return enums, nil
}

// LoadEnumValues loads enum values.
func LoadEnumValues(ctx context.Context, args *Args, enum *xo.Enum) error {
	db, l, schema := DbLoaderSchema(ctx)
	// load enum values
	enumValues, err := l.EnumValues(ctx, db, schema, enum.Name)
	if err != nil {
		return err
	}
	// process enum values
	for _, val := range enumValues {
		enum.Values = append(enum.Values, xo.Field{
			Name:       val.EnumValue,
			ConstValue: &val.ConstValue,
		})
	}
	return nil
}

// LoadProcs loads stored procedures definitions.
func LoadProcs(ctx context.Context, args *Args) ([]xo.Proc, error) {
	db, l, schema := DbLoaderSchema(ctx)
	// not supplied, so bail
	if l.Procs == nil {
		return nil, nil
	}
	// load procs
	procs, err := l.Procs(ctx, db, schema)
	if err != nil {
		return nil, err
	}
	// process procs
	procMap := make(map[string]xo.Proc)
	for _, proc := range procs {
		// skip triggers
		if proc.ReturnType == "trigger" {
			continue
		}
		// parse return type into template
		// TODO: fix this so that nullable types can be returned
		typ, prec, scale, array, err := parseType(l.Driver, proc.ReturnType)
		if err != nil {
			return nil, err
		}
		var returnFields []xo.Field
		// if already in map, proc has >1 return value
		if p, ok := procMap[proc.ProcName]; ok {
			returnFields = p.Returns
		}
		name := proc.ReturnName
		if name == "" || name == "-" {
			name = fmt.Sprintf("r%d", len(returnFields))
		}
		p := &xo.Proc{
			Name: proc.ProcName,
			Kind: proc.ProcKind,
			Returns: append(returnFields, xo.Field{
				Name: name,
				Datatype: xo.Datatype{
					Type:  typ,
					Prec:  prec,
					Scale: scale,
					Array: array,
				},
			}),
		}
		// load proc parameters
		if err := LoadProcParams(ctx, args, p); err != nil {
			return nil, err
		}
		procMap[proc.ProcName] = *p
	}
	var m []xo.Proc
	for _, proc := range procMap {
		if len(proc.Returns) == 1 && proc.Returns[0].Datatype.Type == "void" {
			proc.Returns = nil
			proc.Void = true
		}
		m = append(m, proc)
	}
	sort.Slice(m, func(i, j int) bool {
		return m[i].Name < m[j].Name
	})
	return m, nil
}

// LoadProcParams loads stored procedure parameters.
func LoadProcParams(ctx context.Context, args *Args, proc *xo.Proc) error {
	db, l, schema := DbLoaderSchema(ctx)
	// load proc params
	params, err := l.ProcParams(ctx, db, schema, proc.Name)
	if err != nil {
		return err
	}
	// process
	for i, param := range params {
		typ, prec, scale, array, err := parseType(l.Driver, param.ParamType)
		if err != nil {
			return err
		}
		name := param.ParamName
		if name == "" {
			name = fmt.Sprintf("p%d", i)
		}
		proc.Params = append(proc.Params, xo.Field{
			Name: name,
			Datatype: xo.Datatype{
				Type:  typ,
				Prec:  prec,
				Scale: scale,
				Array: array,
			},
		})
	}
	return nil
}

// LoadTables loads types for the kind (ie, table/view definitions).
func LoadTables(ctx context.Context, args *Args, kind string) ([]xo.Table, error) {
	db, l, schema := DbLoaderSchema(ctx)
	// load tables
	tables, err := l.Tables(ctx, db, schema, kind)
	if err != nil {
		return nil, err
	}
	sort.Slice(tables, func(i, j int) bool {
		return tables[i].TableName < tables[j].TableName
	})
	// create types
	var m []xo.Table
	for _, t := range tables {
		// create template
		table := &xo.Table{
			Type:   kind,
			Name:   t.TableName,
			Manual: true,
		}
		// process columns
		if err := LoadColumns(ctx, args, table); err != nil {
			return nil, err
		}
		// load indexes
		if err := LoadTableIndexes(ctx, args, table); err != nil {
			return nil, err
		}
		m = append(m, *table)
	}
	// load foreign keys
	for i, table := range m {
		fkeys, err := LoadTableForeignKeys(ctx, args, m, table)
		if err != nil {
			return nil, err
		}
		for _, fkey := range fkeys {
			// manual foreign key name generation if name not found
			if fkey.Name == "" {
				var names []string
				for _, field := range fkey.Fields {
					names = append(names, field.Name)
				}
				fkey.Name = table.Name + "_" + strings.Join(names, "_") + "_fkey"
			}
			// foreign key called func name
			fkey.RefFuncName = indexFuncName(fkey.RefTable, xo.Index{
				IsUnique: true,
				Fields:   fkey.RefFields,
			}, false)
			// determine foreign key func name
			fkey.ResolvedName = resolveFkName(args.SchemaParams.FkMode, table, fkey)
			table.ForeignKeys = append(table.ForeignKeys, fkey)
		}
		sort.Slice(table.ForeignKeys, func(i, j int) bool {
			return table.ForeignKeys[i].Name < table.ForeignKeys[j].Name
		})
		m[i] = table
	}
	return m, nil
}

// LoadColumns loads table/view columns.
func LoadColumns(ctx context.Context, args *Args, table *xo.Table) error {
	db, l, schema := DbLoaderSchema(ctx)
	// load sequences
	sequences, err := l.TableSequences(ctx, db, schema, table.Name)
	if err != nil {
		return err
	}
	sqMap := make(map[string]bool)
	for _, s := range sequences {
		table.Manual = false
		sqMap[s.ColumnName] = true
	}
	// load columns
	columns, err := l.TableColumns(ctx, db, schema, table.Name)
	if err != nil {
		return err
	}
	// process columns
	for _, c := range columns {
		ignore := false
		for _, ignoreField := range args.SchemaParams.Ignore {
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
		typ, prec, scale, array, err := parseType(l.Driver, c.DataType)
		if err != nil {
			return err
		}
		var d *string = nil
		if c.DefaultValue.Valid {
			d = &c.DefaultValue.String
		}
		col := xo.Field{
			Name: c.ColumnName,
			Datatype: xo.Datatype{
				Type:     typ,
				Nullable: !c.NotNull,
				Prec:     prec,
				Scale:    scale,
				Array:    array,
			},
			Default:    d,
			IsPrimary:  c.IsPrimaryKey,
			IsSequence: sqMap[c.ColumnName],
		}
		table.Columns = append(table.Columns, col)
		if col.IsPrimary {
			table.PrimaryKeys = append(table.PrimaryKeys, col)
		}
	}
	return nil
}

// LoadTableIndexes loads index definitions per table.
func LoadTableIndexes(ctx context.Context, args *Args, table *xo.Table) error {
	db, l, schema := DbLoaderSchema(ctx)
	// load indexes
	indexes, err := l.TableIndexes(ctx, db, schema, table.Name)
	if err != nil {
		return err
	}
	sort.Slice(indexes, func(i, j int) bool {
		return indexes[i].IndexName < indexes[j].IndexName
	})
	// process indexes
	var priIxLoaded bool
	for _, index := range indexes {
		// save whether or not the primary key index was processed
		priIxLoaded = priIxLoaded || index.IsPrimary
		// create index template
		index := &xo.Index{
			Name:      index.IndexName,
			IsPrimary: index.IsPrimary,
			IsUnique:  index.IsUnique,
		}
		// load index columns
		if err := LoadIndexColumns(ctx, args, table, index); err != nil {
			return err
		}
		// load index func name
		index.FuncName = indexFuncName(table.Name, *index, args.SchemaParams.UseIndexNames)
		table.Indexes = append(table.Indexes, *index)
	}
	pkeys := table.PrimaryKeys
	// if no primary key index loaded, but a primary key column was defined in
	// the type, then create the definition here. this is needed for sqlite, as
	// sqlite doesn't define primary keys in its index list.
	// however, oracle is omitted because indexes for primary keys are not marked
	// as such from introspection queries.
	if l.Driver != "oracle" && !priIxLoaded && len(pkeys) != 0 {
		indexName := table.Name + "_"
		for _, pkey := range pkeys {
			indexName += pkey.Name + "_"
		}
		index := xo.Index{
			Name:      indexName + "pkey",
			Fields:    pkeys,
			IsUnique:  true,
			IsPrimary: true,
		}
		index.FuncName = indexFuncName(table.Name, index, args.SchemaParams.UseIndexNames)
		table.Indexes = append(table.Indexes, index)
	}
	return nil
}

// LoadIndexColumns loads the index column information.
func LoadIndexColumns(ctx context.Context, args *Args, table *xo.Table, index *xo.Index) error {
	db, l, schema := DbLoaderSchema(ctx)
	// load index columns
	cols, err := l.IndexColumns(ctx, db, schema, table.Name, index.Name)
	if err != nil {
		return err
	}
	// process index columns
	for _, col := range cols {
		var field *xo.Field
		// find field
		for _, f := range table.Columns {
			if f.Name == col.ColumnName {
				field = &f
				break
			}
		}
		// no corresponding field found
		if field == nil {
			continue
		}
		index.Fields = append(index.Fields, *field)
	}
	return nil
}

// LoadTableForeignKeys loads foreign key definitions per table.
func LoadTableForeignKeys(ctx context.Context, args *Args, tables []xo.Table, table xo.Table) (map[string]xo.ForeignKey, error) {
	db, l, schema := DbLoaderSchema(ctx)
	// load foreign keys
	fkeys, err := l.TableForeignKeys(ctx, db, schema, table.Name)
	if err != nil {
		return nil, err
	}
	sort.Slice(fkeys, func(i, j int) bool {
		return fkeys[i].ForeignKeyName < fkeys[j].ForeignKeyName
	})
	fkMap := make(map[string]xo.ForeignKey)
	// loop over foreign keys for table
	for _, fkey := range fkeys {
		var refTable *xo.Table
		var field, refField *xo.Field
		// find field from columns
		for _, c := range table.Columns {
			if c.Name == fkey.ColumnName {
				field = &c
				break
			}
		}
		// find ref table
		for _, t := range tables {
			if t.Name == fkey.RefTableName {
				refTable = &t
				break
			}
		}
		// find ref field from columns
		for _, f := range refTable.Columns {
			if f.Name == fkey.RefColumnName {
				refField = &f
				break
			}
		}
		// no ref field, but have ref table, so use primary key
		if refTable != nil && refField == nil && len(refTable.PrimaryKeys) == 1 {
			refField = &refTable.PrimaryKeys[0]
		}
		// check everything was found
		if field == nil || refTable == nil || refField == nil {
			return nil, fmt.Errorf(
				"table %q.%q could not find field, refType, or refField for foreign key: %q",
				schema,
				table.Name,
				fkey.ForeignKeyName,
			)
		}
		// ForeignKeyName should only be empty on SQLite. When this happens, we
		// resort to using the keyid (which is unique to each foreign key, even
		// if it references multiple columns) as the map for the foreign key.
		mapKey := fkey.ForeignKeyName
		if fkey.ForeignKeyName == "" {
			mapKey = strconv.Itoa(fkey.KeyID)
		}
		// create fkey or append (ref)fields
		var f xo.ForeignKey
		var ok bool
		if f, ok = fkMap[mapKey]; !ok {
			f = xo.ForeignKey{
				Name:     fkey.ForeignKeyName,
				RefTable: refTable.Name,
			}
		}
		f.Fields = append(f.Fields, *field)
		f.RefFields = append(f.RefFields, *refField)
		fkMap[mapKey] = f
	}
	return fkMap, nil
}

// parseType parses "type[ (precision[,scale])][\[\]]" strings returning the
// parsed precision, scale, and if the type is an array or not.
//
// Expected formats:
//
//	type
//	type(precision)
//	type(precision, scale)
//	type(precision, scale)[]
//	timestamp(n) with [local] time zone (oracle only)
//
// The returned type is stripped of precision and scale.
func parseType(driver, typ string) (string, int, int, bool, error) {
	// special case for oracle timestamp(n) with [local] time zone
	if m := oracleTimestampRE.FindStringSubmatch(typ); driver == "oracle" && m != nil {
		prec, err := strconv.Atoi(m[1])
		if err != nil {
			return "", 0, 0, false, fmt.Errorf("could not parse precision: %w", err)
		}
		return "timestamp " + m[2], prec, 0, false, nil
	}
	// handle normal
	var prec, scale int
	if m := precRE.FindStringIndex(typ); m != nil {
		s := typ[m[0]+1 : m[1]-1]
		if i := strings.LastIndex(s, ","); i != -1 {
			var err error
			if scale, err = strconv.Atoi(strings.TrimSpace(s[i+1:])); err != nil {
				return "", 0, 0, false, fmt.Errorf("could not parse scale: %w", err)
			}
			s = s[:i]
		}
		// extract precision
		var err error
		if prec, err = strconv.Atoi(strings.TrimSpace(s)); err != nil {
			return "", 0, 0, false, fmt.Errorf("could not parse precision: %w", err)
		}
		typ = typ[:m[0]]
	}
	typ = strings.TrimSpace(typ)
	isArray := strings.HasSuffix(typ, "[]")
	return strings.ToLower(strings.TrimSuffix(typ, "[]")), prec, scale, isArray, nil
}

// oracleTimestampRE is the regexp that matches "timestamp(precision) with [local]
// time zone" definitions in oracle databases
var oracleTimestampRE = regexp.MustCompile(`^timestamp\((\d)\) (with(?: local)? time zone)$`)

// precRE is the regexp that matches "(precision[,scale])" definitions in a
// database.
var precRE = regexp.MustCompile(`\(([0-9]+)(\s*,\s*[0-9]+\s*)?\)$`)

// resolveFkName returns the foreign key name for the passed foreign key.
// The function converts all names to snake_case.
func resolveFkName(mode string, table xo.Table, fkey xo.ForeignKey) string {
	tableName := singularize(fkey.RefTable)
	switch mode {
	case "parent":
		// parent causes a foreign key field to be named in the form of
		// "<type>.<ParentName>".
		//
		// For example, if you have an `authors` and `books` tables, then the
		// foreign key func will be Book.Author.
		return tableName
	case "field":
		// field causes a foreign key field to be named in the form of
		// "<type>.<ParentName>_by_<Field1>_<Field2>".
		//
		// For example, if you have an `authors` and `books` tables, then the
		// foreign key func will be book.AuthorByAuthorIDAuthorName
		var names []string
		for _, f := range fkey.Fields {
			names = append(names, f.Name)
		}
		return tableName + "_by_" + strings.Join(names, "_")
	case "key":
		// key causes a foreign key field to be named in the form of
		// "<type>.<ParentName>By<ForeignKeyName>".
		//
		// For example, if you have `authors` and `books` tables with a foreign
		// key name of 'fk_123', then the foreign key func will be
		// Book.AuthorByFk123
		return tableName + "_by_" + fkey.Name
	case "smart":
		// smart is the default.
		//
		// When there are no naming conflicts, smart behaves like parent,
		// otherwise it behaves the same as field.
		//
		// inspect all foreign keys and use field if conflict found
		for _, v := range table.ForeignKeys {
			if fkey.Name != v.Name && fkey.RefTable == v.RefTable {
				return resolveFkName("field", table, fkey)
			}
		}
		// no conflict, so use parent mode
		return resolveFkName("parent", table, fkey)
	}
	panic(fmt.Sprintf("invalid mode %q", mode))
}

// indexFuncName creates the func name for an index and its supplied fields.
func indexFuncName(tableName string, index xo.Index, useIndexNames bool) string {
	// func name
	if index.IsUnique {
		tableName = inflector.Singularize(tableName)
	}
	name := indexName(index.Name, tableName)
	if useIndexNames && name != "" {
		return tableName + "_by_" + name
	}
	names := []string{tableName, "by"}
	// add param names
	for _, field := range index.Fields {
		names = append(names, field.Name)
	}
	return strings.Join(names, "_")
}

// indexName determines the name for an index.
func indexName(name, tableName string) string {
	// remove suffix _ix, _idx, _index, _pkey, or _key
	if m := indexSuffixRE.FindStringIndex(name); m != nil {
		name = name[:m[0]]
	}
	// check tableName
	if name == tableName {
		return ""
	}
	// chop off tablename_
	if strings.HasPrefix(name, tableName+"_") {
		name = name[len(tableName)+1:]
	}
	// camel case name
	return name
}

// indexSuffixRE is the regexp of index name suffixes that will be chopped off.
var indexSuffixRE = regexp.MustCompile(`(?i)_(ix|idx|index|pkey|ukey|key)$`)

// singuralize will singularize a identifier, returning in CamelCase.
func singularize(s string) string {
	if i := lastIndex(s, '_'); i != -1 {
		return s[:i] + "_" + inflector.Singularize(s[i+1:])
	}
	return inflector.Singularize(s)
}

// lastIndex finds the last rune r in s, returning -1 if not present.
func lastIndex(s string, c rune) int {
	r := []rune(s)
	for i := len(r) - 1; i >= 0; i-- {
		if r[i] == c {
			return i
		}
	}
	return -1
}
