package cmd

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/kenshaw/inflector"
	"github.com/xo/xo/loader"
	"github.com/xo/xo/models"
	xo "github.com/xo/xo/types"
)

// LoadSchema loads a schema from a database.
func LoadSchema(ctx context.Context, set *xo.Set, args *Args) error {
	driver, _, schemaName := xo.DriverDbSchema(ctx)
	schema := xo.Schema{
		Driver: driver,
		Name:   schemaName,
	}
	var err error
	// load enums, procs, tables, views
	if schema.Enums, err = LoadEnums(ctx, args); err != nil {
		return err
	}
	if schema.Procs, err = LoadProcs(ctx, args); err != nil {
		return err
	}
	if schema.Tables, err = LoadTables(ctx, args, "table"); err != nil {
		return err
	}
	if schema.Views, err = LoadTables(ctx, args, "view"); err != nil {
		return err
	}
	// fix enums for mysql
	if driver == "mysql" {
		for i := 0; i < len(schema.Tables); i++ {
			for j := 0; j < len(schema.Tables[i].Columns); j++ {
				if e := schema.EnumByName(schema.Tables[i].Columns[j].Type.Type); e != nil {
					schema.Tables[i].Columns[j].Type.Enum = e
				}
			}
		}
	}
	// emit
	set.Schemas = append(set.Schemas, schema)
	return nil
}

// LoadEnums loads enums.
func LoadEnums(ctx context.Context, args *Args) ([]xo.Enum, error) {
	// load enums
	enumNames, err := loader.Enums(ctx)
	if err != nil {
		return nil, err
	}
	sort.Slice(enumNames, func(i, j int) bool {
		return enumNames[i].EnumName < enumNames[j].EnumName
	})
	// process enums
	var enums []xo.Enum
	for _, enum := range enumNames {
		if !validType(args, false, enum.EnumName) {
			continue
		}
		e := &xo.Enum{
			Name: enum.EnumName,
		}
		if err := LoadEnumValues(ctx, args, e); err != nil {
			return nil, err
		}
		enums = append(enums, *e)
	}
	return enums, nil
}

// LoadEnumValues loads enum values.
func LoadEnumValues(ctx context.Context, args *Args, enum *xo.Enum) error {
	// load enum values
	enumValues, err := loader.EnumValues(ctx, enum.Name)
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
	driver, _, _ := xo.DriverDbSchema(ctx)
	// load procs
	procs, err := loader.Procs(ctx)
	if err != nil {
		return nil, err
	}
	// process procs
	procMap := make(map[string]xo.Proc)
	for _, proc := range procs {
		if !validType(args, false, proc.ProcName) {
			continue
		}
		// parse return type into template
		// TODO: fix this so that nullable types can be returned
		d, err := xo.ParseType(proc.ReturnType, driver)
		if err != nil {
			return nil, err
		}
		var returnFields []xo.Field
		// if already in map, proc has >1 return value
		if p, ok := procMap[proc.ProcID]; ok {
			returnFields = p.Returns
		}
		name := proc.ReturnName
		if name == "" || name == "-" {
			name = fmt.Sprintf("r%d", len(returnFields))
		}
		p := &xo.Proc{
			Type: proc.ProcType,
			ID:   proc.ProcID,
			Name: proc.ProcName,
			Returns: append(returnFields, xo.Field{
				Name: name,
				Type: d,
			}),
			Definition: strings.TrimSpace(proc.ProcDef),
		}
		// load proc parameters
		if err := LoadProcParams(ctx, args, p); err != nil {
			return nil, err
		}
		procMap[proc.ProcID] = *p
	}
	var m []xo.Proc
	for _, proc := range procMap {
		if len(proc.Returns) == 1 && proc.Returns[0].Type.Type == "void" {
			proc.Void, proc.Returns = true, proc.Returns[1:]
		}
		m = append(m, proc)
	}
	sort.Slice(m, func(i, j int) bool {
		if m[i].Name == m[j].Name {
			return m[i].ID < m[j].ID
		}
		return m[i].Name < m[j].Name
	})
	return m, nil
}

// LoadProcParams loads stored procedure parameters.
func LoadProcParams(ctx context.Context, args *Args, proc *xo.Proc) error {
	driver, _, _ := xo.DriverDbSchema(ctx)
	// load proc params
	params, err := loader.ProcParams(ctx, proc.ID)
	if err != nil {
		return err
	}
	// process
	for i, param := range params {
		d, err := xo.ParseType(param.ParamType, driver)
		if err != nil {
			return err
		}
		name := param.ParamName
		if name == "" {
			name = fmt.Sprintf("p%d", i)
		}
		proc.Params = append(proc.Params, xo.Field{
			Name: name,
			Type: d,
		})
	}
	return nil
}

// LoadTables loads types for the type (ie, table/view definitions).
func LoadTables(ctx context.Context, args *Args, typ string) ([]xo.Table, error) {
	// load tables
	tables, err := loader.Tables(ctx, typ)
	if err != nil {
		return nil, err
	}
	sort.Slice(tables, func(i, j int) bool {
		return tables[i].TableName < tables[j].TableName
	})
	// create types
	var m []xo.Table
	for _, table := range tables {
		if !validType(args, false, table.TableName) {
			continue
		}
		// create table
		t := &xo.Table{
			Type:       typ,
			Name:       table.TableName,
			Manual:     true,
			Definition: strings.TrimSpace(table.ViewDef),
		}
		// fix multi-line comments
		if t.Definition != "" {
			t.Definition = strings.Replace(t.Definition, "\n", " ", -1)
		}

		// process columns
		if err := LoadColumns(ctx, args, t); err != nil {
			return nil, err
		}
		// load indexes
		if err := LoadTableIndexes(ctx, args, t); err != nil {
			return nil, err
		}
		m = append(m, *t)
	}
	// load foreign keys
	for i, table := range m {
		if m[i].ForeignKeys, err = LoadTableForeignKeys(ctx, args, m, table); err != nil {
			return nil, err
		}
	}
	return m, nil
}

// LoadColumns loads table/view columns.
func LoadColumns(ctx context.Context, args *Args, table *xo.Table) error {
	driver, _, _ := xo.DriverDbSchema(ctx)
	// load sequences
	sequences, err := loader.TableSequences(ctx, table.Name)
	if err != nil {
		return err
	}
	sqMap := make(map[string]bool)
	for _, s := range sequences {
		table.Manual = false
		sqMap[s.ColumnName] = true
	}
	// load columns
	columns, err := loader.TableColumns(ctx, table.Name)
	if err != nil {
		return err
	}
	// process columns
	for _, c := range columns {
		if !validType(args, true, table.Name, c.ColumnName) {
			continue
		}
		// set col info
		d, err := xo.ParseType(c.DataType, driver)
		if err != nil {
			return err
		}
		d.Nullable = !c.NotNull
		defaultValue := c.DefaultValue.String
		if defaultValue == "NULL" || sqMap[c.ColumnName] {
			defaultValue = ""
		}
		col := xo.Field{
			Name:       c.ColumnName,
			Type:       d,
			Default:    defaultValue,
			IsPrimary:  c.IsPrimaryKey,
			IsSequence: sqMap[c.ColumnName],
			Comment:    strings.TrimSpace(c.Comment.String),
		}
		// fix multi-line comments
		if col.Comment != "" {
			col.Comment = strings.Replace(col.Comment, "\n", " ", -1)
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
	// load indexes
	indexes, err := loader.TableIndexes(ctx, table.Name)
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
		index.Func = indexFuncName(*index, table.Name, args.SchemaParams.UseIndexNames)
		table.Indexes = append(table.Indexes, *index)
	}
	pkeys := table.PrimaryKeys
	// if no primary key index loaded, but a primary key column was defined in
	// the type, then create the definition here. this is needed for sqlite, as
	// sqlite doesn't define primary keys in its index list.
	// however, oracle is omitted because indexes for primary keys are not marked
	// as such from introspection queries.
	driver, _, _ := xo.DriverDbSchema(ctx)
	if driver != "oracle" && !priIxLoaded && len(pkeys) != 0 {
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
		index.Func = indexFuncName(index, table.Name, args.SchemaParams.UseIndexNames)
		table.Indexes = append(table.Indexes, index)
	} else if driver == "oracle" && len(table.PrimaryKeys) != 0 {
	loop:
		for i, index := range table.Indexes {
			for _, field := range index.Fields {
				if !field.IsPrimary {
					continue loop
				}
			}
			table.Indexes[i].IsPrimary = true
			break
		}
	}
	return nil
}

// LoadIndexColumns loads the index column information.
func LoadIndexColumns(ctx context.Context, args *Args, table *xo.Table, index *xo.Index) error {
	// load index columns
	cols, err := loader.IndexColumns(ctx, table.Name, index.Name)
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
func LoadTableForeignKeys(ctx context.Context, args *Args, tables []xo.Table, table xo.Table) ([]xo.ForeignKey, error) {
	// load foreign keys
	foreignKeys, err := loader.TableForeignKeys(ctx, table.Name)
	if err != nil {
		return nil, err
	}
	fkMap := make(map[string]xo.ForeignKey)
	// loop over foreign keys for table
	for _, fkey := range foreignKeys {
		// if the referenced table is excluded, we don't want to omit it
		if !validType(args, false, fkey.RefTableName) {
			fmt.Fprintf(os.Stderr, "WARNING: skipping table %q foreign key %q (%q previously excluded)\n", table.Name, fkey.ForeignKeyName, fkey.RefTableName)
			continue
		}
		// check foreign key
		field, refTable, refField := xo.Field{}, xo.Table{}, xo.Field{}
		if err := checkFk(tables, table, fkey, &field, &refTable, &refField); err != nil {
			return nil, err
		}
		// ForeignKeyName should only be empty on SQLite. When this happens, we
		// resort to using the keyid (which is unique to each foreign key, even
		// if it references multiple columns) as the map for the foreign key
		key := fkey.ForeignKeyName
		if fkey.ForeignKeyName == "" {
			key = strconv.Itoa(fkey.KeyID)
		}
		f := fkMap[key]
		fkMap[key] = xo.ForeignKey{
			Name:      fkey.ForeignKeyName,
			Fields:    append(f.Fields, field),
			RefTable:  refTable.Name,
			RefFields: append(f.RefFields, refField),
		}
	}
	// convert from map to slice
	var fkeys []xo.ForeignKey
	for _, fkey := range fkMap {
		// manual foreign key name generation if name not found
		if fkey.Name == "" {
			var names []string
			for _, field := range fkey.Fields {
				names = append(names, field.Name)
			}
			fkey.Name = table.Name + "_" + strings.Join(names, "_") + "_fkey"
		}
		// determine foreign key func name
		fkey.Func = resolveFkName(fkey, table, args.SchemaParams.FkMode.AsString())
		// foreign key called func name
		fkey.RefFunc = indexFuncName(xo.Index{
			IsUnique: true,
			Fields:   fkey.RefFields,
		}, fkey.RefTable, false)
		fkeys = append(fkeys, fkey)
	}
	// sort fkeys
	sort.Slice(fkeys, func(i, j int) bool {
		return fkeys[i].Name < fkeys[j].Name
	})
	return fkeys, nil
}

// validType returns whether the type name given is valid, given the --include
// and --exclude options provided by the user.
func validType(args *Args, skipIncludes bool, names ...string) bool {
	include, exclude := args.SchemaParams.Include.AsGlob(), args.SchemaParams.Exclude.AsGlob()
	if len(include) == 0 && len(exclude) == 0 {
		return true
	}
	target := strings.Join(names, ".")
	for _, pattern := range exclude {
		if pattern.Match(target) {
			return false
		}
	}
	if len(include) == 0 || skipIncludes {
		return true
	}
	for _, pattern := range include {
		if pattern.Match(target) {
			return true
		}
	}
	return false
}

// checkFk checks that the foreign key has a matching field, ref table, and ref
// field
func checkFk(tables []xo.Table, table xo.Table, fkey *models.ForeignKey, field *xo.Field, refTable *xo.Table, refField *xo.Field) error {
	// find field from columns
	var fieldFound, refTableFound, refFieldFound bool
	for _, c := range table.Columns {
		if c.Name == fkey.ColumnName {
			fieldFound, *field = true, c
			break
		}
	}
	// find ref table
	for _, t := range tables {
		if t.Name == fkey.RefTableName {
			refTableFound, *refTable = true, t
			break
		}
	}
	// find ref field from columns
	for _, f := range refTable.Columns {
		if f.Name == fkey.RefColumnName {
			refFieldFound, *refField = true, f
			break
		}
	}
	// no ref field, but have ref table, so use primary key
	if refTable != nil && refField == nil && len(refTable.PrimaryKeys) == 1 {
		refFieldFound, *refField = true, refTable.PrimaryKeys[0]
	}
	// check everything was found
	if !fieldFound || !refTableFound || !refFieldFound {
		return fmt.Errorf("could not find field, refTable, or refField for table %q foreign key %q", table.Name, fkey.ForeignKeyName)
	}
	return nil
}

// resolveFkName returns the foreign key name for the passed foreign key.
// The function converts all names to snake_case.
func resolveFkName(fkey xo.ForeignKey, table xo.Table, mode string) string {
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
				return resolveFkName(fkey, table, "field")
			}
		}
		// no conflict, so use parent mode
		return resolveFkName(fkey, table, "parent")
	}
	panic(fmt.Sprintf("invalid mode %q", mode))
}

// indexFuncName creates the func name for an index and its supplied fields.
func indexFuncName(index xo.Index, tableName string, useIndexNames bool) string {
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
	if i := strings.LastIndex(s, "_"); i != -1 {
		return s[:i+1] + inflector.Singularize(s[i+1:])
	}
	return inflector.Singularize(s)
}
