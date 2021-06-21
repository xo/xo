package gotpl

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"text/template"

	"github.com/kenshaw/snaker"
	"github.com/xo/xo/templates"
)

// Funcs is a set of template funcs.
type Funcs struct {
	driver    string
	schema    string
	first     *bool
	nth       func(int) string
	pkg       string
	tags      []string
	imports   []string
	conflict  string
	custom    string
	escSchema bool
	escTable  bool
	escColumn bool
	fieldtag  *template.Template
	context   string
	inject    string
	// knownTypes is the collection of known Go types.
	knownTypes map[string]bool
	// shorts is the collection of Go style short names for types, mainly
	// used for use with declaring a func receiver on a type.
	shorts map[string]string
}

// NewFuncs returns a set of template funcs.
func NewFuncs(ctx context.Context, knownTypes map[string]bool, shorts map[string]string, first *bool) (*Funcs, error) {
	// force not first
	if NotFirst(ctx) {
		b := false
		first = &b
	}
	esc := Esc(ctx)
	var tags []string
	for _, tag := range Tag(ctx) {
		if tag != "" {
			tags = append(tags, tag)
		}
	}
	var imports []string
	for _, s := range Import(ctx) {
		if s != "" {
			imports = append(imports, s)
		}
	}
	if s := UUID(ctx); s != "" {
		imports = append(imports, s)
	}
	// parse field tag template
	fieldtag, err := template.New("fieldtag").Parse(FieldTag(ctx))
	if err != nil {
		return nil, err
	}
	// load inject
	inject := Inject(ctx)
	if s := InjectFile(ctx); s != "" {
		buf, err := ioutil.ReadFile(s)
		if err != nil {
			return nil, fmt.Errorf("unable to read file: %v", err)
		}
		inject = string(buf)
	}
	return &Funcs{
		driver:     templates.Driver(ctx),
		schema:     templates.Schema(ctx),
		first:      first,
		nth:        templates.NthParam(ctx),
		pkg:        Pkg(ctx),
		tags:       tags,
		imports:    imports,
		conflict:   Conflict(ctx),
		custom:     Custom(ctx),
		escSchema:  !contains(esc, "none") && (contains(esc, "all") || contains(esc, "schema")),
		escTable:   !contains(esc, "none") && (contains(esc, "all") || contains(esc, "table")),
		escColumn:  !contains(esc, "none") && (contains(esc, "all") || contains(esc, "column")),
		fieldtag:   fieldtag,
		context:    Context(ctx),
		inject:     inject,
		knownTypes: knownTypes,
		shorts:     shorts,
	}, nil
}

// AddKnownType adds a known type.
func (f *Funcs) AddKnownType(name string) {
	f.knownTypes[name] = true
}

// FuncMap returns the func map.
func (f *Funcs) FuncMap() template.FuncMap {
	return template.FuncMap{
		// general
		"driver":  f.driverfn,
		"schema":  f.schemafn,
		"first":   f.firstfn,
		"pkg":     f.pkgfn,
		"tags":    f.tagsfn,
		"imports": f.importsfn,
		"inject":  f.injectfn,
		"eval":    f.eval,
		// context
		"context":         f.contextfn,
		"context_both":    f.context_both,
		"context_disable": f.context_disable,
		// func and query
		"func_name_context": f.func_name_context,
		"func_name":         f.func_name,
		"func_context":      f.func_context,
		"func":              f.func_none,
		"sqlstr":            f.sqlstr,
		"db":                f.db,
		// type
		"names":     f.names,
		"names_all": f.names_all,
		"zero":      f.zero,
		"type":      f.typefn,
		"field":     f.field,
		"short":     f.short,
		/*
			"nthparam":        f.nthparam,
			// general
			"colcount":           f.colcount,
			"colname":            f.colname,
			"colnames":           f.colnames,
			"colnamesmulti":      f.colnamesmulti,
			"colnamesquery":      f.colnamesquery,
			"colnamesquerymulti": f.colnamesquerymulti,
			"colprefixnames":     f.colprefixnames,
			"colvals":            f.colvals,
			"colvalsmulti":       f.colvalsmulti,
			"convext":            f.convext,
			"fieldnames":         f.fieldnames,
			"fieldnamesmulti":    f.fieldnamesmulti,
			"startcount":         f.startcount,
			"hascolumn":          f.hascolumn,
			"hasfield":           f.hasfield,
			"paramlist":          f.paramlist,
			"retype":             f.retype,
		*/
	}
}

// driverfn returns true if the driver is any of the passed drivers.
func (f *Funcs) driverfn(drivers ...string) bool {
	for _, driver := range drivers {
		if f.driver == driver {
			return true
		}
	}
	return false
}

// schemafn takes a series of names and joins them with the schema name.
func (f *Funcs) schemafn(names ...string) string {
	s := f.schema
	// escape table names
	if f.escTable {
		for i, name := range names {
			names[i] = f.esc(name, "table")
		}
	}
	n := strings.Join(names, ".")
	switch {
	case s == "" && n == "":
		return ""
	case f.driver == "sqlite3" && n == "":
		return f.schema
	case f.driver == "sqlite3":
		return n
	case s != "" && n != "":
		if f.escSchema {
			s = f.esc(s, "schema")
		}
		s += "."
	}
	return s + n
}

// firstfn returns true if it is the template was the first template generated.
func (f *Funcs) firstfn() bool {
	b := *f.first
	*f.first = false
	return b
}

// pkgfn returns the package name.
func (f *Funcs) pkgfn() string {
	return f.pkg
}

// tagsfn returns the tags.
func (f *Funcs) tagsfn() []string {
	return f.tags
}

// importsfn returns the imports.
func (f *Funcs) importsfn() []PackageImport {
	var imports []PackageImport
	for _, s := range f.imports {
		alias, pkg := "", s
		if i := strings.Index(pkg, " "); i != -1 {
			alias, pkg = pkg[:i], strings.TrimSpace(pkg[i:])
		}
		imports = append(imports, PackageImport{
			Alias: alias,
			Pkg:   pkg,
		})
	}
	return imports
}

// contextfn returns true when the context mode is both or only.
func (f *Funcs) contextfn() bool {
	return f.context == "both" || f.context == "only"
}

// context_both returns true with the context mode is both.
func (f *Funcs) context_both() bool {
	return f.context == "both"
}

// context_disable returns true with the context mode is both.
func (f *Funcs) context_disable() bool {
	return f.context == "disable"
}

// injectfn returns the injected content provided from args.
func (f *Funcs) injectfn() string {
	return f.inject
}

// eval evalutates a template s against v.
func (f *Funcs) eval(v interface{}, s string) (string, error) {
	tpl, err := template.New(fmt.Sprintf("[EVAL %q]", s)).Parse(s)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	if err := tpl.Execute(buf, v); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// func_name builds a func name.
func (f *Funcs) func_name(v interface{}) string {
	var name string
	switch x := v.(type) {
	case *templates.Query:
		name = nameContext(false, x.Name)
	default:
		return fmt.Sprintf("[[ UNSUPPORTED TYPE: %T ]]", v)
	}
	return name
}

// func_name_context generates a name for the func.
func (f *Funcs) func_name_context(v interface{}) string {
	var name string
	switch x := v.(type) {
	case *templates.Query:
		name = nameContext(f.context_both(), x.Name)
	default:
		return fmt.Sprintf("[[ UNSUPPORTED TYPE: %T ]]", v)
	}
	return name
}

// funcfn builds a func definition.
func (f *Funcs) funcfn(name string, context bool, v interface{}) string {
	var p, r []string
	switch x := v.(type) {
	case *templates.Query:
		// params
		if context {
			p = append(p, "ctx context.Context")
		}
		p = append(p, "db DB")
		for _, z := range x.Params {
			p = append(p, fmt.Sprintf("%s %s", z.Name, z.Type))
		}
		// returns
		switch {
		case x.Exec:
			r = append(r, "sql.Result")
		case x.Flat:
			for _, z := range x.Type.Fields {
				r = append(r, f.typefn(z.Type))
			}
		case x.One:
			r = append(r, "*"+x.Type.Name)
		default:
			r = append(r, "[]*"+x.Type.Name)
		}
		r = append(r, "error")
	default:
		return fmt.Sprintf("[[ UNSUPPORTED TYPE: %T ]]", v)
	}
	return fmt.Sprintf("func %s(%s) (%s)", name, strings.Join(p, ", "), strings.Join(r, ", "))
}

// func_context generates a func signature for v with context determined by the context mode.
func (f *Funcs) func_context(v interface{}) string {
	return f.funcfn(f.func_name_context(v), f.contextfn(), v)
}

// func_none genarates a func signature for v without context.
func (f *Funcs) func_none(v interface{}) string {
	return f.funcfn(f.func_name(v), false, v)
}

// sqlstr generates a sqlstr for the specified query and any accompanying
// comments.
func (f *Funcs) sqlstr(v interface{}) string {
	var interpolate bool
	var query, comments []string
	switch x := v.(type) {
	case *templates.Query:
		interpolate, query, comments = x.Interpolate, x.Query, x.Comments
	default:
		return fmt.Sprintf(`const sqlstr = [[ NOT IMPLEMENTED FOR %T ]]`, v)
	}
	typ := "const"
	if interpolate {
		typ = "var"
	}
	var lines []string
	for i := 0; i < len(query); i++ {
		line := "`" + query[i] + "`"
		if i != len(query)-1 {
			line += " + "
		}
		if s := strings.TrimSpace(comments[i]); s != "" {
			line += "// " + s
		}
		lines = append(lines, line)
	}
	sqlstr := stripRE.ReplaceAllString(strings.Join(lines, "\n"), " ")
	return fmt.Sprintf("%s sqlstr = %s", typ, sqlstr)
}

var stripRE = regexp.MustCompile(`\s+\+\s+` + "``")

// db generates a db.<name>Context(ctx, sqlstr, ...)
func (f *Funcs) db(name string, v interface{}) string {
	// params
	var p []interface{}
	if f.contextfn() {
		name, p = name+"Context", append(p, "ctx")
	}
	return fmt.Sprintf(`db.%s(%s)`, name, f.names("", append(p, "sqlstr", v)...))
}

// names generates a list of names.
func (f *Funcs) namesfn(all bool, prefix string, z ...interface{}) string {
	var names []string
	for i, v := range z {
		switch x := v.(type) {
		case string:
			names = append(names, x)
		case *templates.Query:
			for _, p := range x.Params {
				if !all && p.Interpolate {
					continue
				}
				names = append(names, prefix+p.Name)
			}
		case *templates.Type:
			for _, p := range x.Fields {
				names = append(names, prefix+p.Name)
			}
		case []*templates.Field:
			for _, p := range x {
				names = append(names, prefix+p.Name)
			}
		default:
			names = append(names, fmt.Sprintf("/* UNSUPPORTED TYPE %d %T */", i, v))
		}
	}
	return strings.Join(names, ", ")
}

// names generates a list of names (excluding certain ones such as interpolated
// names).
func (f *Funcs) names(prefix string, z ...interface{}) string {
	return f.namesfn(false, prefix, z...)
}

// names_all generates a list of all names.
func (f *Funcs) names_all(prefix string, z ...interface{}) string {
	return f.namesfn(true, prefix, z...)
}

// zero generates a zero list.
func (f *Funcs) zero(z ...interface{}) string {
	var zeroes []string
	for i, v := range z {
		switch x := v.(type) {
		case string:
			zeroes = append(zeroes, x)
		case *templates.Type:
			for _, p := range x.Fields {
				zeroes = append(zeroes, p.Zero)
			}
		case []*templates.Field:
			for _, p := range x {
				zeroes = append(zeroes, p.Zero)
			}
		default:
			zeroes = append(zeroes, fmt.Sprintf("/* UNSUPPORTED TYPE %d %T */", i, v))
		}
	}
	return strings.Join(zeroes, ", ")
}

// typefn generates the Go type, prefixing the custom package name if applicable.
func (f *Funcs) typefn(typ string) string {
	if strings.Contains(typ, ".") {
		return typ
	}
	prefix := ""
	for strings.HasPrefix(typ, "[]") {
		typ = typ[2:]
		prefix = prefix + "[]"
	}
	if _, ok := f.knownTypes[typ]; !ok {
		pkg := f.custom
		if pkg != "" {
			pkg = pkg + "."
		}
		return prefix + pkg + typ
	}
	return prefix + typ
}

// field generates a field definition for a struct.
func (f *Funcs) field(field *templates.Field) (string, error) {
	tag, comment := "", ""
	buf := new(bytes.Buffer)
	if err := f.fieldtag.Funcs(f.FuncMap()).Execute(buf, field); err != nil {
		return "", err
	}
	if s := buf.String(); s != "" {
		tag = " " + s
	}
	if field.Col != nil {
		comment = " // " + field.Col.ColumnName
	}
	return fmt.Sprintf("%s %s%s%s", field.Name, field.Type, tag, comment), nil
}

// short generates a safe Go identifier for typ. typ is first checked
// against shorts, and if not found, then the value is calculated and
// stored in the shorts for future use.
//
// A short is the concatentation of the lowercase of the first character in
// the words comprising the name. For example, "MyCustomName" will have have
// the short of "mcn".
//
// If a generated short conflicts with a Go reserved name or a name used in
// the templates, then the corresponding value in goReservedNames map will be
// used.
//
// Generated shorts that have conflicts with any scopeConflicts member will
// have nameConflictSuffix appended.
func (f *Funcs) short(v interface{}) string {
	var n string
	switch x := v.(type) {
	case *templates.Type:
		n = x.Name
	default:
		return fmt.Sprintf("[[ UNSUPPORTED TYPE %T ]]", v)
	}
	// check short name map
	name, ok := f.shorts[n]
	if !ok {
		// calc the short name
		var u []string
		for _, s := range strings.Split(strings.ToLower(snaker.CamelToSnake(n)), "_") {
			if len(s) > 0 && s != "id" {
				u = append(u, s[:1])
			}
		}
		name = strings.Join(u, "")
		// check go reserved names
		if n, ok := goReservedNames[name]; ok {
			name = n
		}
		// store back to short name map
		f.shorts[n] = name
	}
	// append suffix if conflict exists
	if _, ok := templateReservedNames[name]; ok {
		name += f.conflict
	}
	return name
}

/*

// colnames creates a list of the column names found in fields, excluding any
// Field with Name contained in ignore.
//
// Used to present a comma separated list of column names, that can be used in
// a SELECT, or UPDATE, or other SQL clause requiring an list of identifiers
// (ie, "field_1, field_2, field_3, ...").
func (f *Funcs) colnames(fields []*templates.Field, ignore ...string) string {
	m := map[string]bool{}
	for _, n := range ignore {
		m[n] = true
	}
	s, i := "", 0
	for _, field := range fields {
		if m[field.Name] {
			continue
		}
		if i != 0 {
			s = s + ", "
		}
		s += f.colname(field.Col)
		i++
	}
	return s
}

// colnamesmulti creates a list of the column names found in fields, excluding any
// Field with Name contained in ignore.
//
// Used to present a comma separated list of column names, that can be used in
// a SELECT, or UPDATE, or other SQL clause requiring an list of identifiers
// (ie, "field_1, field_2, field_3, ...").
func (f *Funcs) colnamesmulti(fields []*templates.Field, ignore []*templates.Field) string {
	m := map[string]bool{}
	for _, field := range ignore {
		m[field.Name] = true
	}
	s, i := "", 0
	for _, field := range fields {
		if m[field.Name] {
			continue
		}
		if i != 0 {
			s += ", "
		}
		s += f.colname(field.Col)
		i++
	}
	return s
}

// colnamesquery creates a list of the column names in fields as a query and
// joined by sep, excluding any Field with Name contained in ignore.
//
// Used to create a list of column names in a WHERE clause (ie, "field_1 = $1
// AND field_2 = $2 AND ...") or in an UPDATE clause (ie, "field = $1, field =
// $2, ...").
func (f *Funcs) colnamesquery(fields []*templates.Field, sep string, ignore ...string) string {
	m := map[string]bool{}
	for _, n := range ignore {
		m[n] = true
	}
	s, i := "", 0
	for _, field := range fields {
		if m[field.Name] {
			continue
		}
		if i != 0 {
			s += sep
		}
		s += f.colname(field.Col) + " = " + f.nth(i)
		i++
	}
	return s
}

// colnamesquerymulti creates a list of the column names in fields as a query and
// joined by sep, excluding any Field with Name contained in the slice of fields in ignore.
//
// Used to create a list of column names in a WHERE clause (ie, "field_1 = $1
// AND field_2 = $2 AND ...") or in an UPDATE clause (ie, "field = $1, field =
// $2, ...").
func (f *Funcs) colnamesquerymulti(fields []*templates.Field, sep string, startCount int, ignore []*templates.Field) string {
	m := map[string]bool{}
	for _, field := range ignore {
		m[field.Name] = true
	}
	s, i := "", startCount
	for _, field := range fields {
		if m[field.Name] {
			continue
		}
		if i > startCount {
			s += sep
		}
		s += f.colname(field.Col) + " = " + f.nth(i)
		i++
	}
	return s
}

// colprefixnames creates a list of the column names found in fields with the
// supplied prefix, excluding any Field with Name contained in ignore.
//
// Used to present a comma separated list of column names with a prefix. Used in
// a SELECT, or UPDATE (ie, "t.field_1, t.field_2, t.field_3, ...").
func (f *Funcs) colprefixnames(fields []*templates.Field, prefix string, ignore ...string) string {
	m := map[string]bool{}
	for _, n := range ignore {
		m[n] = true
	}
	s, i := "", 0
	for _, field := range fields {
		if m[field.Name] {
			continue
		}
		if i != 0 {
			s += ", "
		}
		s += prefix + "." + f.colname(field.Col)
		i++
	}
	return s
}

// colvals creates a list of value place holders for fields excluding any Field
// with Name contained in ignore.
//
// Used to present a comma separated list of column place holders, used in a
// SELECT or UPDATE statement (ie, "$1, $2, $3 ...").
func (f *Funcs) colvals(fields []*templates.Field, ignore ...string) string {
	m := map[string]bool{}
	for _, n := range ignore {
		m[n] = true
	}
	s, i := "", 0
	for _, field := range fields {
		if m[field.Name] {
			continue
		}
		if i != 0 {
			s += ", "
		}
		s += f.nth(i)
		i++
	}
	return s
}

// colvalsmulti creates a list of value place holders for fields excluding any Field
// with Name contained in ignore.
//
// Used to present a comma separated list of column place holders, used in a
// SELECT or UPDATE statement (ie, "$1, $2, $3 ...").
func (f *Funcs) colvalsmulti(fields []*templates.Field, ignore []*templates.Field) string {
	m := map[string]bool{}
	for _, field := range ignore {
		m[field.Name] = true
	}
	s, i := "", 0
	for _, field := range fields {
		if m[field.Name] {
			continue
		}
		if i != 0 {
			s += ", "
		}
		s += f.nth(i)
		i++
	}
	return s
}

// fieldnames creates a list of field names from fields of the adding the
// provided prefix, and excluding any Field with Name contained in ignore.
//
// Used to present a comma separated list of field names, ie in a Go statement
// (ie, "t.Field1, t.Field2, t.Field3 ...")
func (f *Funcs) fieldnames(fields []*templates.Field, prefix string, ignore ...string) string {
	m := map[string]bool{}
	for _, n := range ignore {
		m[n] = true
	}
	s, i := "", 0
	for _, field := range fields {
		if m[field.Name] {
			continue
		}
		if i != 0 {
			s += ", "
		}
		if prefix == "" {
			s += "&"
		} else {
			s += prefix + "."
		}
		s += field.Name
		i++
	}
	return s
}

// fieldnamesmulti creates a list of field names from fields of the adding the
// provided prefix, and excluding any Field with the slice contained in ignore.
//
// Used to present a comma separated list of field names, ie in a Go statement
// (ie, "t.Field1, t.Field2, t.Field3 ...")
func (f *Funcs) fieldnamesmulti(fields []*templates.Field, prefix string, ignore []*templates.Field) string {
	m := map[string]bool{}
	for _, field := range ignore {
		m[field.Name] = true
	}
	s, i := "", 0
	for _, field := range fields {
		if m[field.Name] {
			continue
		}
		if i != 0 {
			s += ", "
		}
		if prefix == "" {
			s += "&"
		} else {
			s += prefix + "."
		}
		s += field.Name
		i++
	}
	return s
}

// colcount returns the 1-based count of fields, excluding any Field with Name
// contained in ignore.
//
// Used to get the count of fields, and useful for specifying the last SQL
// parameter.
func (f *Funcs) colcount(fields []*templates.Field, ignore ...string) int {
	m := map[string]bool{}
	for _, n := range ignore {
		m[n] = true
	}
	i := 1
	for _, field := range fields {
		if m[field.Name] {
			continue
		}
		i++
	}
	return i
}

// paramlist converts a list of fields into their named Go parameters, skipping
// any Field with Name contained in ignore. addType will cause the go Type to
// be added after each variable name. addPrefix will cause the returned string
// to be prefixed with ", " if the generated string is not empty.
//
// Any field name encountered will be checked against goReservedNames, and will
// have its name substituted by its corresponding looked up value.
//
// Used to present a comma separated list of Go variable names for use with as
// either a Go func parameter list, or in a call to another Go func.
// (ie, ", a, b, c, ..." or ", a T1, b T2, c T3, ...").
func (f *Funcs) paramlist(fields []*templates.Field, addPrefix, addType bool, ignore ...string) string {
	m := map[string]bool{}
	for _, n := range ignore {
		m[n] = true
	}
	var vals []string
	i := 0
	for _, field := range fields {
		if m[field.Name] {
			continue
		}
		s := "v" + strconv.Itoa(i)
		if len(field.Name) > 0 {
			n := strings.Split(snaker.CamelToSnake(field.Name), "_")
			s = strings.ToLower(n[0]) + field.Name[len(n[0]):]
		}
		// check go reserved names
		if r, ok := goReservedNames[strings.ToLower(s)]; ok {
			s = r
		}
		// add the go type
		if addType {
			s += " " + f.retype(field.Type)
		}
		// add to vals
		vals = append(vals, s)
		i++
	}
	// concat generated values
	s := strings.Join(vals, ", ")
	if addPrefix && s != "" {
		return ", " + s
	}
	return s
}

// convext generates the Go conversion for a to be assignable to b.
func (*Funcs) convext(prefix string, a, b *templates.Field) string {
	expr := prefix + "." + a.Name
	if a.Type == b.Type {
		return expr
	}
	typ := a.Type
	if strings.HasPrefix(typ, "sql.Null") {
		expr = expr + "." + a.Type[8:]
		typ = strings.ToLower(a.Type[8:])
	}
	if b.Type != typ {
		expr = b.Type + "(" + expr + ")"
	}
	return expr
}

// colname returns the ColumnName of col, optionally escaping it if
// escapeColumnNames is toggled.
func (f *Funcs) colname(col *models.Column) string {
	if f.escColumn {
		return f.esc(col.ColumnName, "column")
	}
	return col.ColumnName
}

// hascolumn takes a list of fields and determines if field with the specified
// column name is in the list.
func (f *Funcs) hascolumn(fields []*templates.Field, name string) bool {
	for _, field := range fields {
		if field.Col.ColumnName == name {
			return true
		}
	}
	return false
}

// hasfield takes a list of fields and determines if field with the specified
// field name is in the list.
func (f *Funcs) hasfield(fields []*templates.Field, name string) bool {
	for _, field := range fields {
		if field.Name == name {
			return true
		}
	}
	return false
}

// startcount returns a starting count for numbering columns in queries.
func (f *Funcs) startcount(fields []*templates.Field, pkFields []*templates.Field) int {
	return len(fields) - len(pkFields)
}

*/

/*
// nthparam returns the nth param placeholder
func (f *Funcs) nthparam(i int) string {
	return f.nth(i)
}
*/

// esc escapes s.
func (f *Funcs) esc(s string, typ string) string {
	return `"` + s + `"`
}

// PackageImport holds information about a Go package import.
type PackageImport struct {
	Alias string
	Pkg   string
}

// String satisfies the fmt.Stringer interface.
func (v PackageImport) String() string {
	if v.Alias != "" {
		return fmt.Sprintf("%s %q", v.Alias, v.Pkg)
	}
	return fmt.Sprintf("%q", v.Pkg)
}

// templateReservedNames are the template reserved names.
var templateReservedNames = map[string]bool{
	// variables
	"ctx":  true,
	"db":   true,
	"err":  true,
	"res":  true,
	"rows": true,

	// packages
	"context": true,
	"csv":     true,
	"driver":  true,
	"errors":  true,
	"fmt":     true,
	"hstore":  true,
	"regexp":  true,
	"sql":     true,
	"strings": true,
	"time":    true,
	"uuid":    true,
}

// goReservedNames is a map of of go reserved names to "safe" names.
var goReservedNames = map[string]string{
	"break":       "brk",
	"case":        "cs",
	"chan":        "chn",
	"const":       "cnst",
	"continue":    "cnt",
	"default":     "def",
	"defer":       "dfr",
	"else":        "els",
	"fallthrough": "flthrough",
	"for":         "fr",
	"func":        "fn",
	"go":          "goVal",
	"goto":        "gt",
	"if":          "ifVal",
	"import":      "imp",
	"interface":   "iface",
	"map":         "mp",
	"package":     "pkg",
	"range":       "rnge",
	"return":      "ret",
	"select":      "slct",
	"struct":      "strct",
	"switch":      "swtch",
	"type":        "typ",
	"var":         "vr",
	// go types
	"error":      "e",
	"bool":       "b",
	"string":     "str",
	"byte":       "byt",
	"rune":       "r",
	"uintptr":    "uptr",
	"int":        "i",
	"int8":       "i8",
	"int16":      "i16",
	"int32":      "i32",
	"int64":      "i64",
	"uint":       "u",
	"uint8":      "u8",
	"uint16":     "u16",
	"uint32":     "u32",
	"uint64":     "u64",
	"float32":    "z",
	"float64":    "f",
	"complex64":  "c",
	"complex128": "c128",
}

// nameContext adds suffix Context to name.
func nameContext(context bool, name string) string {
	if context {
		return name + "Context"
	}
	return name
}
