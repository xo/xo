// Package types contains xo internal types.
package types

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

// Set is a set of queries and schemas.
type Set struct {
	Queries []Query  `json:"queries,omitempty"`
	Schemas []Schema `json:"schemas,omitempty"`
}

// Query is a query.
type Query struct {
	Driver       string   `json:"driver,omitempty"`
	Name         string   `json:"name,omitempty"`
	Comment      string   `json:"comment,omitempty"`
	Exec         bool     `json:"exec,omitempty"`
	Flat         bool     `json:"flat,omitempty"`
	One          bool     `json:"one,omitempty"`
	Interpolate  bool     `json:"interpolate,omitempty"`
	Type         string   `json:"type,omitempty"`
	TypeComment  string   `json:"type_comment,omitempty"`
	Fields       []Field  `json:"fields,omitempty"`
	ManualFields bool     `json:"manual_fields,omitempty"` // fields generated or provided by user
	Params       []Field  `json:"params,omitempty"`
	Query        []string `json:"query,omitempty"`
	Comments     []string `json:"comments,omitempty"`
}

// MarshalYAML satisfies the yaml.Marshaler interface.
func (q Query) MarshalYAML() (interface{}, error) {
	v := q
	v.Comment = forceLineEnd(v.Comment)
	return reflectStruct(v)
}

// Schema is a SQL schema.
type Schema struct {
	Driver string  `json:"type,omitempty"`
	Name   string  `json:"name,omitempty"`
	Enums  []Enum  `json:"enums,omitempty"`
	Procs  []Proc  `json:"procs,omitempty"`
	Tables []Table `json:"tables,omitempty"`
	Views  []Table `json:"views,omitempty"`
}

// EnumByName returns a enum by its name.
func (s Schema) EnumByName(name string) *Enum {
	for _, e := range s.Enums {
		if e.Name == name {
			return &e
		}
	}
	return nil
}

// Enum is a enum type.
type Enum struct {
	Name   string  `json:"name,omitempty"`
	Values []Field `json:"values,omitempty"`
}

// Proc is a stored procedure.
type Proc struct {
	ID         string  `json:"-"`
	Type       string  `json:"type,omitempty"` // 'procedure' or 'function'
	Name       string  `json:"name,omitempty"`
	Params     []Field `json:"params,omitempty"`
	Returns    []Field `json:"return,omitempty"`
	Void       bool    `json:"void,omitempty"`
	Definition string  `json:"definition,omitempty"`
}

// MarshalYAML satisfies the yaml.Marshaler interface.
func (p Proc) MarshalYAML() (interface{}, error) {
	v := p
	v.Definition = forceLineEnd(v.Definition)
	return reflectStruct(v)
}

// Table is a table or view.
type Table struct {
	Type        string       `json:"type,omitempty"` // 'table' or 'view'
	Name        string       `json:"name,omitempty"`
	Columns     []Field      `json:"columns,omitempty"`
	PrimaryKeys []Field      `json:"primary_keys,omitempty"`
	Indexes     []Index      `json:"indexes,omitempty"`
	ForeignKeys []ForeignKey `json:"foreign_keys,omitempty"`
	Manual      bool         `json:"manual,omitempty"`
	Definition  string       `json:"definition,omitempty"` // empty for tables
}

// MarshalYAML satisfies the yaml.Marshaler interface.
func (t Table) MarshalYAML() (interface{}, error) {
	v := t
	v.Definition = forceLineEnd(v.Definition)
	return reflectStruct(v)
}

// Index is a index.
type Index struct {
	Name      string  `json:"name,omitempty"`
	Fields    []Field `json:"fields,omitempty"`
	IsUnique  bool    `json:"is_unique,omitempty"`
	IsPrimary bool    `json:"is_primary,omitempty"`
	Func      string  `json:"-"`
}

// ForeignKey is a foreign key.
type ForeignKey struct {
	Name      string  `json:"name,omitempty"`       // constraint name
	Fields    []Field `json:"column,omitempty"`     // column that has the key on it
	RefTable  string  `json:"ref_table,omitempty"`  // table the foreign key refers to
	RefFields []Field `json:"ref_column,omitempty"` // column in ref table the index refers to
	Func      string  `json:"-"`                    // foreign key func name (based on fkey mode)
	RefFunc   string  `json:"-"`                    // func name from ref index
}

// Field is a column, index, enum value, or stored procedure parameter.
type Field struct {
	Name        string `json:"name,omitempty"`
	Type        Type   `json:"datatype,omitempty"`
	Default     string `json:"default,omitempty"`
	IsPrimary   bool   `json:"is_primary,omitempty"`
	IsSequence  bool   `json:"is_sequence,omitempty"`
	ConstValue  *int   `json:"const_value,omitempty"`
	Interpolate bool   `json:"interpolate,omitempty"`
	Join        bool   `json:"join,omitempty"`
	Comment     string `json:"comment,omitempty"`
}

// Type holds information for a database type.
type Type struct {
	Type     string `json:"type,omitempty"`
	Prec     int    `json:"prec,omitempty"`
	Scale    int    `json:"scale,omitempty"`
	Nullable bool   `json:"nullable,omitempty"`
	IsArray  bool   `json:"array,omitempty"`
	Unsigned bool   `json:"unsigned,omitempty"`
	Enum     *Enum  `json:"-"`
}

// ParseType parses "type[ (precision[,scale])][\[\]]" strings returning the
// parsed precision, scale, and if the type is an array or not.
//
// Expected formats:
//
//		type
//		type(precision)
//		type(precision, scale)
//	 type(precision, scale) unsigned
//		timestamp(n) with [local] time zone (oracle only)
//
// The returned type is stripped of precision and scale.
func ParseType(typ, driver string) (Type, error) {
	// special case for oracle timestamp(n) with [local] time zone
	if m := oracleTimestampRE.FindStringSubmatch(typ); driver == "oracle" && m != nil {
		prec, err := strconv.Atoi(m[1])
		if err != nil {
			return Type{}, fmt.Errorf("could not parse precision: %w", err)
		}
		return Type{
			Type: "timestamp " + m[2],
			Prec: prec,
		}, nil
	}
	// extract is array
	isArray := false
	if strings.HasSuffix(typ, "[]") {
		typ, isArray = typ[:len(typ)-len("[]")], true
	}
	// extract unsigned
	unsigned := false
	if strings.HasSuffix(typ, " unsigned") {
		typ, unsigned = typ[:len(typ)-len(" unsigned")], true
	}
	// handle normal
	var prec, scale int
	if m := precRE.FindStringIndex(typ); m != nil {
		s := typ[m[0]+1 : m[1]-1]
		if i := strings.LastIndex(s, ","); i != -1 {
			var err error
			if scale, err = strconv.Atoi(strings.TrimSpace(s[i+1:])); err != nil {
				return Type{}, fmt.Errorf("could not parse scale: %w", err)
			}
			s = s[:i]
		}
		// extract precision
		var err error
		if prec, err = strconv.Atoi(strings.TrimSpace(s)); err != nil {
			return Type{}, fmt.Errorf("could not parse precision: %w", err)
		}
		typ = typ[:m[0]]
	}
	return Type{
		Type:     strings.ToLower(strings.TrimSpace(typ)),
		Prec:     prec,
		Scale:    scale,
		IsArray:  isArray,
		Unsigned: unsigned,
	}, nil
}

// oracleTimestampRE is the regexp that matches "timestamp(precision) with [local]
// time zone" definitions in oracle databases.
var oracleTimestampRE = regexp.MustCompile(`^timestamp\((\d)\) (with(?: local)? time zone)$`)

// precRE is the regexp that matches "(precision[,scale])" definitions in a
// database.
var precRE = regexp.MustCompile(`\(([0-9]+)(\s*,\s*[0-9]+\s*)?\)$`)

// TemplateType is a template type.
type TemplateType struct {
	// Name is the template name.
	Name string
	// Modes are the command modes the template is available for.
	Modes []string
	// Flags are additional template flags.
	Flags []Flag
	// Order returns the order of template type processing.
	Order func(ctx context.Context, mode string) []string
	// Funcs provides template funcs for use by templates.
	Funcs func(ctx context.Context, mode string) (template.FuncMap, error)
	// NewContext provides a way for templates to inject additional, global
	// context values, prior to processing.
	NewContext func(ctx context.Context, mode string) context.Context
	// Pre performs pre processing of generated content, such as emitting
	// static files.
	Pre func(ctx context.Context, mode string, set *Set, outFolder fs.FS, emit func(Template)) error
	// Process performs the processing templates for the set.
	Process func(ctx context.Context, mode string, set *Set, emit func(Template)) error
	// Post performs post processing of generated content.
	Post func(ctx context.Context, mode string, files map[string][]byte, emit func(string, []byte)) error
}

// Template holds template information.
type Template struct {
	// Src is the source of the template and will be used when it is non-empty.
	Src string
	// Partial is the partial template name to use, if any.
	Partial string
	// Dest is the destination file.
	Dest string
	// SortType is the sort order type.
	SortType string
	// SortName is the name to sort by.
	SortName string
	// Data is the template data.
	Data interface{}
}

// ContextKey is a context key.
type ContextKey string

// Context keys.
const (
	DriverKey    ContextKey = "driver"
	DbKey        ContextKey = "db"
	SchemaKey    ContextKey = "schema"
	OutKey       ContextKey = "out"
	SingleKey    ContextKey = "single"
	ArrayModeKey ContextKey = "array-mode"
)

// DriverDbSchema returns the driver, database connection, and schema name from
// the context.
func DriverDbSchema(ctx context.Context) (string, *sql.DB, string) {
	driver, _ := ctx.Value(DriverKey).(string)
	db, _ := ctx.Value(DbKey).(*sql.DB)
	schema, _ := ctx.Value(SchemaKey).(string)
	return driver, db, schema
}

// Out returns out option from the context.
func Out(ctx context.Context) string {
	s, _ := ctx.Value(OutKey).(string)
	return s
}

// Single returns the file to write to in single mode.
func Single(ctx context.Context) string {
	s, _ := ctx.Value(SingleKey).(string)
	return s
}

// forceLineEnd forces a \n on a string that doesn't contain one and is
// non-empty.
func forceLineEnd(s string) string {
	if strings.TrimSpace(s) != "" && !strings.Contains(s, "\n") {
		return s + "\n"
	}
	return s
}

// reflectStruct reflects a struct without its json tag.
func reflectStruct(z interface{}) (interface{}, error) {
	v := reflect.ValueOf(z)
	n, typ := v.NumField(), v.Type()
	// build fields
	var fields []reflect.StructField
	var values []reflect.Value
	for i := 0; i < n; i++ {
		f := typ.Field(i)
		// lookup json tag
		name, ok := f.Tag.Lookup("json")
		if !ok {
			return nil, fmt.Errorf("missing json tag on %T field %s (%d)", z, f.Name, i)
		}
		if j := strings.Index(name, ","); j != -1 {
			name = name[:j]
		}
		// skip
		if name == "-" {
			continue
		}
		field := f
		field.Tag = reflect.StructTag(`json:"` + name + `,omitempty"`)
		fields, values = append(fields, field), append(values, v.Field(i))
	}
	// build value
	s := reflect.New(reflect.StructOf(fields))
	for i := 0; i < len(values); i++ {
		s.Elem().Field(i).Set(values[i])
	}
	return s.Elem().Interface(), nil
}
