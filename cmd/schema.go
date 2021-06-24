package cmd

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/gedex/inflector"
	"github.com/kenshaw/snaker"
	"github.com/xo/xo/models"
	"github.com/xo/xo/templates"
)

// SchemaGenerator generates code for a database schema.
type SchemaGenerator struct{}

// NewSchemaGenerator creates a new schema generator.
func NewSchemaGenerator() *SchemaGenerator {
	return &SchemaGenerator{}
}

// Exec satisfies the Generator interface.
func (*SchemaGenerator) Exec(ctx context.Context, args *Args) error {
	// load enums
	if _, err := LoadEnums(ctx, args); err != nil {
		return err
	}
	// load procs
	if _, err := LoadProcs(ctx, args); err != nil {
		return err
	}
	// load table types
	tables, err := LoadTypes(ctx, args, "table")
	if err != nil {
		return err
	}
	// load view types
	views, err := LoadTypes(ctx, args, "view")
	if err != nil {
		return err
	}
	// merge views with tables
	for k, view := range views {
		tables[k] = view
	}
	// load foreign keys
	if err := LoadForeignKeys(ctx, args, tables); err != nil {
		return err
	}
	// load indexes
	if err := LoadIndexes(ctx, args, tables); err != nil {
		return err
	}
	return nil
}

// Process satisfies the Generator interface.
func (*SchemaGenerator) Process(ctx context.Context, args *Args) error {
	return templates.Process(
		ctx,
		args.OutParams.Append,
		args.OutParams.Single,
		"enum", "typedef", "index", "foreignkey", "proc",
	)
}

// LoadEnums loads enums.
func LoadEnums(ctx context.Context, args *Args) (map[string]*templates.Enum, error) {
	db, l, schema := DbLoaderSchema(ctx)
	// not supplied, so bail
	if l.Enums == nil {
		return nil, nil
	}
	// load enums
	schemaEnums, err := l.Enums(ctx, db, schema)
	if err != nil {
		return nil, err
	}
	// process enums
	enums := make(map[string]*templates.Enum)
	for _, enum := range schemaEnums {
		v := &templates.Enum{
			Name: singularize(enum.EnumName),
			Enum: enum,
		}
		if err := LoadEnumValues(ctx, args, v); err != nil {
			return nil, err
		}
		if err := templates.AddKnownType(ctx, v.Name); err != nil {
			return nil, err
		}
		enums[v.Name] = v
	}
	// generate enum templates
	for _, enum := range enums {
		if err := templates.Emit(ctx, &templates.Template{
			Set:      "schema",
			Template: "enum",
			Type:     enum.Name,
			Data:     enum,
		}); err != nil {
			return nil, err
		}
	}
	return enums, nil
}

// LoadEnumValues loads enum values.
func LoadEnumValues(ctx context.Context, args *Args, enum *templates.Enum) error {
	db, l, schema := DbLoaderSchema(ctx)
	// load enum values
	enumValues, err := l.EnumValues(ctx, db, schema, enum.Enum.EnumName)
	if err != nil {
		return err
	}
	// process enum values
	for _, val := range enumValues {
		// chop off redundant enum name if applicable
		name := snaker.SnakeToCamelIdentifier(val.EnumValue)
		if strings.HasSuffix(strings.ToLower(name), strings.ToLower(enum.Name)) {
			n := name[:len(name)-len(enum.Name)]
			if len(n) > 0 {
				name = n
			}
		}
		enum.Values = append(enum.Values, &templates.EnumValue{
			Name: name,
			Val:  val,
		})
	}
	return nil
}

// LoadProcs loads stored procedures definitions.
func LoadProcs(ctx context.Context, args *Args) (map[string]*templates.Proc, error) {
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
	m := make(map[string]*templates.Proc)
	for _, proc := range procs {
		// fix the name if it starts with one or more underscores
		name := proc.ProcName
		for strings.HasPrefix(name, "_") {
			name = name[1:]
		}
		// create template
		// parse return type into template
		// TODO: fix this so that nullable types can be returned
		goType, zero, prec, err := l.GoType(ctx, proc.ReturnType, false)
		if err != nil {
			return nil, err
		}
		p := &templates.Proc{
			Name: snaker.SnakeToCamelIdentifier(name),
			Proc: proc,
			Return: &templates.Field{
				Type: goType,
				Zero: zero,
				Prec: prec,
			},
		}
		// load proc parameters
		if err := LoadProcParams(ctx, args, p); err != nil {
			return nil, err
		}
		m[proc.ProcName] = p
	}
	// generate proc templates
	for _, p := range m {
		if err := templates.Emit(ctx, &templates.Template{
			Set:      "schema",
			Template: "proc",
			Type:     "sp_" + p.Name,
			Data:     p,
		}); err != nil {
			return nil, err
		}
	}
	return m, nil
}

// LoadProcParams loads stored procedure parameters.
func LoadProcParams(ctx context.Context, args *Args, proc *templates.Proc) error {
	db, l, schema := DbLoaderSchema(ctx)
	// load proc params
	params, err := l.ProcParams(ctx, db, schema, proc.Proc.ProcName)
	if err != nil {
		return err
	}
	// process
	for i, param := range params {
		// TODO: some databases support named parameters in procs (MySQL)
		// TODO: fix this so that nullable types can be used as parameters
		goType, zero, prec, err := l.GoType(ctx, strings.TrimSpace(param.ParamType), false)
		if err != nil {
			return err
		}
		name := param.ParamName
		if name == "" {
			name = fmt.Sprintf("v%d", i)
		}
		field := &templates.Field{
			Name: name,
			Type: goType,
			Zero: zero,
			Prec: prec,
		}
		// add to proc params
		if proc.ProcParams != "" {
			proc.ProcParams = proc.ProcParams + ", "
		}
		proc.ProcParams = proc.ProcParams + param.ParamType
		proc.Params = append(proc.Params, field)
	}
	return nil
}

// LoadTypes loads types for the kind (ie, table/view definitions).
func LoadTypes(ctx context.Context, args *Args, kind string) (map[string]*templates.Type, error) {
	// load tables
	tables, err := LoadTables(ctx, args, kind)
	if err != nil {
		return nil, err
	}
	// build table map
	m := make(map[string]*templates.Type)
	for _, table := range tables {
		// create template
		typ := &templates.Type{
			Name:  singularize(table.TableName),
			Kind:  kind,
			Table: table,
		}
		// process columns
		if err := LoadColumns(ctx, args, typ); err != nil {
			return nil, err
		}
		m[table.TableName] = typ
	}
	// generate table templates
	for _, table := range m {
		if err := templates.Emit(ctx, &templates.Template{
			Set:      "schema",
			Template: "typedef",
			Type:     table.Name,
			Data:     table,
		}); err != nil {
			return nil, err
		}
	}
	return m, nil
}

// LoadTables loads tables.
func LoadTables(ctx context.Context, args *Args, kind string) ([]*models.Table, error) {
	db, l, schema := DbLoaderSchema(ctx)
	// load tables
	tables, err := l.Tables(ctx, db, schema, kind)
	if err != nil {
		return nil, err
	}
	if l.TableSequences == nil {
		return tables, nil
	}
	// load sequences
	sequences, err := l.TableSequences(ctx, db, schema)
	if err != nil {
		return nil, err
	}
	// build sequence map
	sequenceMap := make(map[string]bool)
	for _, sequence := range sequences {
		sequenceMap[sequence.TableName] = true
	}
	// force manual pk if not defined in sequences
	for _, table := range tables {
		table.ManualPk = table.ManualPk || !sequenceMap[table.TableName]
	}
	return tables, nil
}

// LoadColumns loads table/view columns.
func LoadColumns(ctx context.Context, args *Args, typ *templates.Type) error {
	db, l, schema := DbLoaderSchema(ctx)
	// load columns
	columns, err := l.TableColumns(ctx, db, schema, typ.Table.TableName)
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
		s, z, prec, err := l.GoType(ctx, c.DataType, !c.NotNull)
		if err != nil {
			return err
		}
		field := &templates.Field{
			Name: snaker.SnakeToCamelIdentifier(c.ColumnName),
			Type: s,
			Zero: z,
			Prec: prec,
			Col:  c,
		}
		// set primary key
		if c.IsPrimaryKey {
			typ.PrimaryKeyFields = append(typ.PrimaryKeyFields, field)
			// This is retained for backward compatibility in the templates.
			typ.PrimaryKey = field
		}
		// append col to template fields
		typ.Fields = append(typ.Fields, field)
	}
	return nil
}

// LoadForeignKeys loads foreign keys.
func LoadForeignKeys(ctx context.Context, args *Args, tables map[string]*templates.Type) error {
	fkeys := make(map[string]*templates.ForeignKey)
	for _, table := range tables {
		// load keys per table
		if err := LoadTableForeignKeys(ctx, args, tables, table, fkeys); err != nil {
			return err
		}
	}
	// determine foreign key names
	for _, fkey := range fkeys {
		fkey.Name = resolveFkName(args.SchemaParams.FkMode, fkeys, fkey)
	}
	// generate templates
	for _, fkey := range fkeys {
		if err := templates.Emit(ctx, &templates.Template{
			Set:      "schema",
			Template: "foreignkey",
			Type:     fkey.Type.Name,
			Name:     fkey.ForeignKey.ForeignKeyName,
			Data:     fkey,
		}); err != nil {
			return err
		}
	}
	return nil
}

// LoadTableForeignKeys loads foreign key definitions per table.
func LoadTableForeignKeys(ctx context.Context, args *Args, tables map[string]*templates.Type, typ *templates.Type, fkMap map[string]*templates.ForeignKey) error {
	db, l, schema := DbLoaderSchema(ctx)
	// load foreign keys
	fkeys, err := l.TableForeignKeys(ctx, db, schema, typ.Table.TableName)
	if err != nil {
		return err
	}
	// loop over foreign keys for table
	for _, fkey := range fkeys {
		var refType *templates.Type
		var field, refField *templates.Field
		// find field from columns
		for _, f := range typ.Fields {
			if f.Col.ColumnName == fkey.ColumnName {
				field = f
				break
			}
		}
		// find ref table
		for _, t := range tables {
			if t.Table.TableName == fkey.RefTableName {
				refType = t
				break
			}
		}
		// find ref field from columns
		for _, f := range refType.Fields {
			if f.Col.ColumnName == fkey.RefColumnName {
				refField = f
				break
			}
		}
		// no ref field, but have ref type, so use primary key
		if refType != nil && refField == nil {
			refField = refType.PrimaryKey
		}
		// check everything was found
		if field == nil || refType == nil || refField == nil {
			return fmt.Errorf("table %q %q could not find field, refType, or refField for foreign key: %q", schema, typ.Table.TableName, fkey.ForeignKeyName)
		}
		// foreign key name
		if fkey.ForeignKeyName == "" {
			fkey.ForeignKeyName = typ.Table.TableName + "_" + field.Col.ColumnName + "_fkey"
		}
		// create foreign key template
		fkMap[fkey.ForeignKeyName] = &templates.ForeignKey{
			Type:       typ,
			Field:      field,
			RefType:    refType,
			RefField:   refField,
			ForeignKey: fkey,
		}
	}
	return nil
}

// LoadIndexes loads index definitions.
func LoadIndexes(ctx context.Context, args *Args, tables map[string]*templates.Type) error {
	indexes := make(map[string]*templates.Index)
	for _, table := range tables {
		// load table indexes
		if err := LoadTableIndexes(ctx, args, table, indexes); err != nil {
			return err
		}
	}
	// generate templates
	for _, index := range indexes {
		if err := templates.Emit(ctx, &templates.Template{
			Set:      "schema",
			Template: "index",
			Type:     index.Type.Name,
			Name:     index.Index.IndexName,
			Data:     index,
		}); err != nil {
			return err
		}
	}
	return nil
}

// LoadTableIndexes loads index definitions per table.
func LoadTableIndexes(ctx context.Context, args *Args, typ *templates.Type, indexes map[string]*templates.Index) error {
	db, l, schema := DbLoaderSchema(ctx)
	// load indexes
	tableIndexes, err := l.TableIndexes(ctx, db, schema, typ.Table.TableName)
	if err != nil {
		return err
	}
	// process indexes
	var priIxLoaded bool
	for _, tableIndex := range tableIndexes {
		// save whether or not the primary key index was processed
		priIxLoaded = priIxLoaded || tableIndex.IsPrimary
		// create index template
		index := &templates.Index{
			Type:  typ,
			Index: tableIndex,
		}
		// load index columns
		if err := LoadIndexColumns(ctx, args, index); err != nil {
			return err
		}
		// build func name
		index.FuncName = indexFuncName(index, args.SchemaParams.UseIndexNames)
		indexes[typ.Table.TableName+"_"+tableIndex.IndexName] = index
	}
	// search for primary key if it was skipped being set in the type
	pkey := typ.PrimaryKey
	if pkey == nil {
		for _, field := range typ.Fields {
			if field.Col.IsPrimaryKey {
				pkey = field
				break
			}
		}
	}
	// if no primary key index loaded, but a primary key column was defined in
	// the type, then create the definition here. this is needed for sqlite, as
	// sqlite doesn't define primary keys in its index list
	if l.Driver != "oracle" && !priIxLoaded && pkey != nil {
		indexName := typ.Table.TableName + "_" + pkey.Col.ColumnName + "_pkey"
		indexes[indexName] = &templates.Index{
			FuncName: typ.Name + "By" + pkey.Name,
			Type:     typ,
			Fields:   []*templates.Field{pkey},
			Index: &models.Index{
				IndexName: indexName,
				IsUnique:  true,
				IsPrimary: true,
			},
		}
	}
	return nil
}

// LoadIndexColumns loads the index column information.
func LoadIndexColumns(ctx context.Context, args *Args, index *templates.Index) error {
	db, l, schema := DbLoaderSchema(ctx)
	// load index columns
	cols, err := l.IndexColumns(ctx, db, schema, index.Type.Table.TableName, index.Index.IndexName)
	if err != nil {
		return err
	}
	// process index columns
	for _, col := range cols {
		var field *templates.Field
		// find field
		for _, f := range index.Type.Fields {
			if f.Col.ColumnName == col.ColumnName {
				field = f
				break
			}
		}
		if field == nil {
			continue
		}
		index.Fields = append(index.Fields, field)
	}
	return nil
}

// resolveFkName returns the foreign key name for the passed foreign key.
func resolveFkName(mode string, fkMap map[string]*templates.ForeignKey, fkey *templates.ForeignKey) string {
	switch mode {
	case "parent":
		// parent causes a foreign key field to be named in the form of
		// "<type>.<ParentName>".
		//
		// For example, if you have an `authors` and `books` tables, then the
		// foreign key func will be Book.Author.
		return fkey.RefType.Name
	case "field":
		// field causes a foreign key field to be named in the form of
		// "<type>.<ParentName>By<Field>".
		//
		// For example, if you have an `authors` and `books` tables, then the
		// foreign key func will be Book.AuthorByAuthorID.
		return fkey.RefType.Name + "By" + fkey.Field.Name
	case "key":
		// key causes a foreign key field to be named in the form of
		// "<type>.<ParentName>By<ForeignKeyName>".
		//
		// For example, if you have an `authors` and `books` tables with a foreign
		// key name of 'fk_123', then the foreign key func will be
		// Book.AuthorByFk123.
		return fkey.RefType.Name + "By" + snaker.SnakeToCamelIdentifier(fkey.ForeignKey.ForeignKeyName)
	case "smart":
		// smart is the default.
		//
		// When there are no naming conflicts, smart behaves the same parent,
		// otherwise it behaves the same as field.
		//
		// inspect all foreign keys and use field if conflict found
		for _, v := range fkMap {
			if fkey != v && fkey.Type.Name == v.Type.Name && fkey.RefType.Name == v.RefType.Name {
				return resolveFkName("field", fkMap, fkey)
			}
		}
		// no conflict, so use parent mode
		return resolveFkName("parent", fkMap, fkey)
	}
	panic(fmt.Sprintf("invalid mode %q", mode))
}

// indexFuncName creates the func name for an index and its supplied fields.
func indexFuncName(index *templates.Index, useIndexNames bool) string {
	// func name
	s := index.Type.Name
	if !index.Index.IsUnique {
		s = inflector.Pluralize(index.Type.Name)
	}
	s += "By"
	// add param names
	var paramNames []string
	name := indexName(index.Index.IndexName, index.Type.Table.TableName)
	if useIndexNames && name != "" {
		paramNames = append(paramNames, name)
	} else {
		for _, field := range index.Fields {
			paramNames = append(paramNames, field.Name)
		}
	}
	return s + strings.Join(paramNames, "")
}

// indexName determines the name for an index, returning in CamelCase.
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
	return snaker.SnakeToCamelIdentifier(name)
}

// indexSuffixRE is the regexp of index name suffixes that will be chopped off.
var indexSuffixRE = regexp.MustCompile(`(?i)_(ix|idx|index|pkey|ukey|key)$`)

// singuralize will singularize a identifier, returning in CamelCase.
func singularize(s string) string {
	if i := lastIndex(s, '_'); i != -1 {
		s = s[:i] + "_" + inflector.Singularize(s[i+1:])
	} else {
		s = inflector.Singularize(s)
	}
	return snaker.SnakeToCamelIdentifier(s)
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
