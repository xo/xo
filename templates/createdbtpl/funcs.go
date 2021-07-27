package createdbtpl

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	xo "github.com/xo/xo/types"
)

// Funcs is a set of template funcs.
type Funcs struct {
	driver     string
	enumMap    map[string]xo.Enum
	constraint bool
	escCols    bool
	escTypes   bool
	engine     string
}

// NewFuncs creates a new Funcs.
func NewFuncs(ctx context.Context, enums []xo.Enum) *Funcs {
	driver, _, _ := xo.DriverSchemaNthParam(ctx)
	enumMap := make(map[string]xo.Enum)
	if driver == "mysql" {
		for _, e := range enums {
			enumMap[e.Name] = e
		}
	}
	return &Funcs{
		driver:     driver,
		enumMap:    enumMap,
		constraint: Constraint(ctx),
		escCols:    Esc(ctx, "columns"),
		escTypes:   Esc(ctx, "types"),
		engine:     Engine(ctx),
	}
}

// FuncMap returns the func map.
func (f *Funcs) FuncMap() template.FuncMap {
	return template.FuncMap{
		"coldef":          f.coldef,
		"viewdef":         f.viewdef,
		"procdef":         f.procdef,
		"driver":          f.driverfn,
		"constraint":      f.constraintfn,
		"esc":             f.escType,
		"fields":          f.fields,
		"engine":          f.enginefn,
		"literal":         f.literal,
		"comma":           f.comma,
		"isEndConstraint": f.isEndConstraint,
	}
}

func (f *Funcs) coldef(table xo.Table, field xo.Field) string {
	// normalize type
	typ := f.normalize(field.Datatype)
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
	if !field.Datatype.Nullable && !field.IsSequence {
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
		if !field.Datatype.Nullable {
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
	if !strings.HasSuffix(def, ";") {
		def += ";"
	}
	return def
}

func (f *Funcs) procdef(proc xo.Proc) string {
	// don't escape trailing suffix for sqlserver
	if f.driver == "sqlserver" {
		proc.Definition = strings.TrimSuffix(proc.Definition, ";")
	}
	// escape semicolons
	// FIXME: do not escape semicolons in string literals
	if f.driver != "postgres" {
		proc.Definition = strings.ReplaceAll(proc.Definition, ";", "\\;")
	}
	// definition already provided
	switch f.driver {
	case "oracle":
		proc.Definition = "CREATE " + proc.Definition
		fallthrough
	case "sqlserver":
		return proc.Definition
	}
	// specify query language when using postgres
	var suffix string
	if f.driver == "postgres" {
		suffix = "\n$$ LANGUAGE plpgsql"
	}
	return f.procSignature(proc) + "\n" + proc.Definition + suffix
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
		params = append(params, fmt.Sprintf("%s %s", f.escCol(field.Name), f.normalize(field.Datatype)))
	}
	// add return values
	if len(proc.Returns) == 1 && proc.Returns[0].Name == "r0" {
		end += " RETURNS " + f.normalize(proc.Returns[0].Datatype)
	} else {
		for _, field := range proc.Returns {
			params = append(params, fmt.Sprintf("OUT %s %s", f.escCol(field.Name), f.normalize(field.Datatype)))
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

func (f *Funcs) normalize(datatype xo.Datatype) string {
	typ := f.convert(datatype.Type)
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

func (f *Funcs) convert(typ string) string {
	// mysql enums
	if e, ok := f.enumMap[typ]; f.driver == "mysql" && ok {
		var enums []string
		for _, v := range e.Values {
			enums = append(enums, fmt.Sprintf("'%s'", v.Name))
		}
		return fmt.Sprintf("ENUM(%s)", strings.Join(enums, ", "))
	}
	// check aliases
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

func (f *Funcs) comma(i int, v interface{}) string {
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

func (f *Funcs) isEndConstraint(idx *xo.Index) bool {
	if f.driver == "sqlite3" && idx.Fields[0].IsSequence {
		return false
	}
	return idx.IsPrimary || idx.IsUnique
}

var escapeValues = map[string][2]string{
	"postgres":  {`"`, `"`},
	"mysql":     {"`", "`"},
	"sqlite3":   {"[", "]"},
	"sqlserver": {`"`, `"`},
	"oracle":    {`"`, `"`},
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
