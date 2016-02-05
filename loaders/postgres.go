package loaders

import (
	"bytes"
	"database/sql"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/gedex/inflector"
	"github.com/knq/xo/internal"
	"github.com/knq/xo/models"
	"github.com/knq/xo/templates"
	"github.com/serenize/snaker"
)

// PgLoadTypes loads the postgres type definitions from a database.
func PgLoadTypes(args *internal.ArgType, db *sql.DB, typeMap map[string]*bytes.Buffer) error {
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

	// load idx
	_, err = PgLoadIdx(args, db, typeMap, tableMap)

	//
	//_, err = PgLoadForeignIdx(args, db, typeMap)

	// load procs
	_, err = PgLoadProcs(args, db, typeMap)
	if err != nil {
		return err
	}

	tableMap = tableMap

	// loop over tables
	//_, err = loadIndexes(db, typeMap, tableMap)

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
	enumMap := make(map[string]*templates.EnumTemplate)
	for _, e := range enums {
		// calc go type and add to known types
		goType := inflector.Singularize(snaker.SnakeToCamel(e.EnumType))
		templates.KnownTypeMap[goType] = true

		// copy values back into model
		e.Value = snaker.SnakeToCamel(strings.ToLower(e.EnumValue))
		e.Type = goType

		// set value in enum map if not present
		typ := strings.ToLower(goType)
		if _, ok := enumMap[typ]; !ok {
			enumMap[typ] = &templates.EnumTemplate{
				Type:     goType,
				EnumType: e.EnumType,
				Values:   make([]*models.Enum, 0),
			}
		}

		// append enum to template vals
		enumMap[typ].Values = append(enumMap[typ].Values, e)
	}

	// generate enum templates
	for typ, em := range enumMap {
		buf := GetBuf(typeMap, typ)
		err = templates.Tpls["postgres.enum.go.tpl"].Execute(buf, em)
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
	procMap := make(map[string]*templates.ProcTemplate)
	for _, p := range procs {
		// fix the name if it starts with underscore
		name := p.Name
		if name[0:1] == "_" {
			name = name[1:]
		}

		// create template
		procTpl := templates.ProcTemplate{
			Name:               snaker.SnakeToCamel(name),
			TableSchema:        args.Schema,
			ProcName:           p.Name,
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

		procMap[strings.ToLower("sp_"+p.Name)] = &procTpl
	}

	// generate proc templates
	for typ, pm := range procMap {
		buf := GetBuf(typeMap, typ)
		err = templates.Tpls["postgres.proc.go.tpl"].Execute(buf, pm)
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
	cols, err := models.ColumnsByTableSchema(db, args.Schema)
	if err != nil {
		return nil, err
	}

	// process columns
	fieldMap := make(map[string]map[string]bool)
	tableMap := make(map[string]*templates.TableTemplate)
	for _, c := range cols {
		if _, ok := fieldMap[c.TableName]; !ok {
			fieldMap[c.TableName] = make(map[string]bool)
		}

		tableType := inflector.Singularize(snaker.SnakeToCamel(c.TableName))
		typ := strings.ToLower(tableType)

		// set col info
		c.Field = snaker.SnakeToCamel(c.ColumnName)
		c.Len, c.NilType, c.Type = PgParseType(args, c.DataType, c.IsNullable)

		// set value in table map if not present
		if _, ok := tableMap[typ]; !ok {
			tableMap[typ] = &templates.TableTemplate{
				Type:        tableType,
				TableSchema: args.Schema,
				TableName:   c.TableName,
				Fields:      make([]*models.Column, 0),
			}
		}

		// set primary key
		if c.IsPrimaryKey {
			tableMap[typ].PrimaryKey = c.ColumnName
			tableMap[typ].PrimaryKeyField = c.Field
			tableMap[typ].PrimaryKeyType = c.Type
		}

		// append col to template fields
		if _, ok := fieldMap[c.TableName][c.ColumnName]; !ok {
			tableMap[typ].Fields = append(tableMap[typ].Fields, c)
		}

		fieldMap[c.TableName][c.ColumnName] = true
	}

	// generate table templates
	for typ, t := range tableMap {
		buf := GetBuf(typeMap, typ)
		err = templates.Tpls["postgres.model.go.tpl"].Execute(buf, t)
		if err != nil {
			return nil, err
		}
	}

	return tableMap, nil
}

// PgLoadIdx loads indexes from the database.
func PgLoadIdx(args *internal.ArgType, db *sql.DB, typeMap map[string]*bytes.Buffer, tableMap map[string]*templates.TableTemplate) (map[string]*templates.IdxTemplate, error) {
	var err error

	// load idx's
	idxMap := make(map[string]*templates.IdxTemplate)
	for _, t := range tableMap {
		// find relevant columns
		fields := make([]*models.Column, 0)
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
