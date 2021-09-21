// Package funcs provides custom template funcs.
package funcs

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"github.com/xo/xo/templates/createdbtpl"
	xo "github.com/xo/xo/types"
)

// Init intializes the custom template funcs.
func Init(ctx context.Context) (template.FuncMap, error) {
	driver, _, _ := xo.DriverDbSchema(ctx)
	funcs := &Funcs{
		driver:      driver,
		constraint:  createdbtpl.Constraint(ctx),
		escCols:     createdbtpl.Esc(ctx, "columns"),
		escTypes:    createdbtpl.Esc(ctx, "types"),
		engine:      createdbtpl.Engine(ctx),
		trimComment: createdbtpl.TrimComment(ctx),
	}
	return template.FuncMap{
		"coldef":          funcs.coldef,
		"viewdef":         funcs.viewdef,
		"procdef":         funcs.procdef,
		"driver":          funcs.driverfn,
		"constraint":      funcs.constraintfn,
		"esc":             funcs.escType,
		"fields":          funcs.fields,
		"engine":          funcs.enginefn,
		"literal":         funcs.literal,
		"isEndConstraint": funcs.isEndConstraint,
		"comma":           comma,
	}, nil
}

// Funcs is a set of template funcs.
type Funcs struct {
	driver      string
	constraint  bool
	escCols     bool
	escTypes    bool
	engine      string
	trimComment bool
}

func (f *Funcs) coldef(table xo.Table, field xo.Field) string {
	// normalize type
	typ := f.normalize(field.Type)
	// add sequence definition
	if field.IsSequence {
		typ = f.resolveSequence(typ, field)
	}
	// column def
	def := []string{f.escCol(field.Name), typ}
	// add default value
	if field.Default != "" && !field.IsSequence {
		def = append(def, "DEFAULT", f.alterDefault(field.Default))
	}
	if !field.Type.Nullable && !field.IsSequence {
		def = append(def, "NOT NULL")
	}
	// add constraints
	if fk := f.colFKey(table, field); fk != "" {
		def = append(def, fk)
	}
	return strings.Join(def, " ")
}

// alterDefault parses and alters default column values based on the driver.
func (f *Funcs) alterDefault(s string) string {
	switch f.driver {
	case "postgres":
		if m := postgresDefaultCastRE.FindStringSubmatch(s); m != nil {
			return m[1]
		}
	case "mysql":
		if v := strings.ToUpper(s); v == "CURRENT_TIMESTAMP()" {
			return "CURRENT_TIMESTAMP"
		}
	case "sqlite3":
		if !sqliteDefaultNeedsParenRE.MatchString(s) {
			return "(" + s + ")"
		}
	}
	return s
}

// postgresDefaultCastRE is the regexp to strip the datatype cast from the
// postgres default value.
var postgresDefaultCastRE = regexp.MustCompile(`(.*)::[a-zA-Z_ ]*(\[\])?$`)

// sqliteDefaultNeedsParen is the regexp to test whether the given value is
// correctly surrounded with parenthesis
//
// If it starts and ends with a parenthesis or a single or double quote, it
// does not need to be quoted with parenthesis.
var sqliteDefaultNeedsParenRE = regexp.MustCompile(`^([\('"].*[\)'"]|\d+)$`)

func (f *Funcs) resolveSequence(typ string, field xo.Field) string {
	switch f.driver {
	case "postgres":
		switch typ {
		case "SMALLINT":
			return "SMALLSERIAL"
		case "INTEGER":
			return "SERIAL"
		case "BIGINT":
			return "BIGSERIAL"
		}
	case "mysql":
		return typ + " AUTO_INCREMENT"
	case "sqlite3":
		ext := " PRIMARY KEY AUTOINCREMENT"
		if !field.Type.Nullable {
			ext = " NOT NULL" + ext
		}
		return typ + ext
	case "sqlserver":
		return typ + " IDENTITY(1, 1)"
	case "oracle":
		return typ + " GENERATED ALWAYS AS IDENTITY"
	}
	return ""
}

func (f *Funcs) colFKey(table xo.Table, field xo.Field) string {
	for _, fk := range table.ForeignKeys {
		if len(fk.Fields) == 1 && fk.Fields[0] == field {
			tblName, fieldName := f.escType(fk.RefTable), fk.RefFields[0].Name
			return fmt.Sprintf("%sREFERENCES %s (%s)", f.constraintfn(fk.Name), tblName, fieldName)
		}
	}
	return ""
}

func (f *Funcs) viewdef(view xo.Table) string {
	def := view.Definition
	switch f.driver {
	case "postgres", "mysql", "oracle":
		def = fmt.Sprintf("CREATE VIEW %s AS\n%s", f.escType(view.Name), view.Definition)
	}
	if f.trimComment {
		if strings.HasPrefix(def, "--") {
			def = def[strings.Index(def, "\n")+1:]
		}
	}
	return strings.TrimSuffix(def, ";")
}

func (f *Funcs) procdef(proc xo.Proc) string {
	def := f.cleanProcDef(proc.Definition)
	// prepend signature definition
	if f.driver == "postgres" || f.driver == "mysql" {
		def = f.procSignature(proc) + "\n" + def
	}
	return def
}

func (f *Funcs) cleanProcDef(def string) string {
	switch f.driver {
	// nothing needs to be done for postgres
	// only add the query language suffix
	case "postgres":
		return def + "\n$$ LANGUAGE plpgsql"
	// the trailing semicolon shouldn't be escaped for sqlserver
	case "sqlserver":
		def = strings.TrimSuffix(def, ";")
	// oracle only just needs the CREATE prefix
	case "oracle":
		def = "CREATE " + def
	}
	if f.trimComment {
		if strings.HasPrefix(def, "--") {
			def = def[strings.Index(def, "\n")+1:]
		}
	}
	return strings.ReplaceAll(def, ";", "\\;")
}

func (f *Funcs) procSignature(proc xo.Proc) string {
	// create function signature
	typ := "PROCEDURE"
	if proc.Type == "function" {
		typ = "FUNCTION"
	}
	var params []string
	var end string
	// add params
	for _, field := range proc.Params {
		params = append(params, fmt.Sprintf("%s %s", f.escCol(field.Name), f.normalize(field.Type)))
	}
	// add return values
	if len(proc.Returns) == 1 && proc.Returns[0].Name == "r0" {
		end += " RETURNS " + f.normalize(proc.Returns[0].Type)
	} else {
		for _, field := range proc.Returns {
			params = append(params, fmt.Sprintf("OUT %s %s", f.escCol(field.Name), f.normalize(field.Type)))
		}
	}
	signature := fmt.Sprintf("CREATE %s %s(%s)%s", typ, f.escType(proc.Name), strings.Join(params, ", "), end)
	if f.driver == "postgres" {
		signature += " AS $$"
	}
	return signature
}

func (f *Funcs) driverfn(allowed ...string) bool {
	for _, d := range allowed {
		if f.driver == d {
			return true
		}
	}
	return false
}

func (f *Funcs) constraintfn(name string) string {
	if f.constraint || f.driver == "sqlserver" || f.driver == "oracle" {
		return fmt.Sprintf("CONSTRAINT %s ", f.escType(name))
	}
	return ""
}

func (f *Funcs) fields(v interface{}) string {
	switch x := v.(type) {
	case []xo.Field:
		var fs []string
		for _, field := range x {
			fs = append(fs, f.escCol(field.Name))
		}
		return strings.Join(fs, ", ")
	}
	return fmt.Sprintf("[[ UNKNOWN TYPE %T ]]", v)
}

func (f *Funcs) esc(value string, escape bool) string {
	if !escape {
		return value
	}
	var start, end string
	switch f.driver {
	case "postgres", "sqlite3", "oracle":
		start, end = `"`, `"`
	case "mysql":
		start, end = "`", "`"
	case "sqlserver":
		start, end = "[", "]"
	}
	return start + value + end
}

func (f *Funcs) escType(value string) string {
	return f.esc(value, f.escTypes)
}

func (f *Funcs) escCol(value string) string {
	return f.esc(value, f.escCols)
}

func (f *Funcs) enginefn() string {
	if f.driver != "mysql" || f.engine == "" {
		return ""
	}
	return fmt.Sprintf(" ENGINE=%s", f.engine)
}

func (f *Funcs) normalize(datatype xo.Type) string {
	typ := f.convert(datatype)
	if datatype.Scale > 0 && !omitPrecision[f.driver][typ] {
		typ += fmt.Sprintf("(%d, %d)", datatype.Prec, datatype.Scale)
	} else if datatype.Prec > 0 && !omitPrecision[f.driver][typ] {
		typ += fmt.Sprintf("(%d)", datatype.Prec)
	}
	if datatype.Unsigned {
		typ += " UNSIGNED"
	}
	if datatype.IsArray {
		typ += "[]"
	}
	return typ
}

func (f *Funcs) convert(datatype xo.Type) string {
	// mysql enums
	if f.driver == "mysql" && datatype.Enum != nil {
		var enums []string
		for _, v := range datatype.Enum.Values {
			enums = append(enums, fmt.Sprintf("'%s'", v.Name))
		}
		return fmt.Sprintf("ENUM(%s)", strings.Join(enums, ", "))
	}
	// check aliases
	typ := datatype.Type
	if alias, ok := typeAliases[f.driver][typ]; ok {
		typ = alias
	}
	return strings.ToUpper(typ)
}

// literal properly escapes string literals within single quotes
// (Used for enum values in postgres)
func (f *Funcs) literal(literal string) string {
	return fmt.Sprint("'", strings.ReplaceAll(literal, "'", "''"), "'")
}

func (f *Funcs) isEndConstraint(idx xo.Index) bool {
	if f.driver == "sqlite3" && idx.Fields[0].IsSequence {
		return false
	}
	return idx.IsPrimary || idx.IsUnique
}

var typeAliases = map[string]map[string]string{
	"postgres": {
		"character varying":           "varchar",
		"character":                   "char",
		"time without time zone":      "time",
		"timestamp without time zone": "timestamp",
		"time with time zone":         "timetz",
		"timestamp with time zone":    "timestamptz",
	},
}

var omitPrecision = map[string]map[string]bool{
	"sqlserver": {
		"TINYINT":        true,
		"SMALLINT":       true,
		"INT":            true,
		"BIGINT":         true,
		"REAL":           true,
		"SMALLMONEY":     true,
		"MONEY":          true,
		"BIT":            true,
		"DATE":           true,
		"TIME":           true,
		"DATETIME":       true,
		"DATETIME2":      true,
		"SMALLDATETIME":  true,
		"DATETIMEOFFSET": true,
	},
	"oracle": {
		"TIMESTAMP":                      true,
		"TIMESTAMP WITH TIME ZONE":       true,
		"TIMESTAMP WITH LOCAL TIME ZONE": true,
	},
}

func comma(i int, v interface{}) string {
	var l int
	switch x := v.(type) {
	case []xo.Field:
		l = len(x)
	}
	if i+1 < l {
		return ","
	}
	return ""
}
