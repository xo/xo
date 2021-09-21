// Package funcs provides custom template funcs.
package funcs

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"text/template"

	"github.com/kenshaw/snaker"
	"github.com/xo/xo/loader"
	"github.com/xo/xo/templates/gotpl"
	xo "github.com/xo/xo/types"
)

// Init intializes the custom template funcs.
func Init(ctx context.Context) (template.FuncMap, error) {
	// force not first
	first := gotpl.First(ctx)
	if gotpl.NotFirst(ctx) {
		b := false
		first = &b
	}
	// parse field tag template
	fieldtag, err := template.New("fieldtag").Parse(gotpl.FieldTag(ctx))
	if err != nil {
		return nil, err
	}
	// load inject
	inject := gotpl.Inject(ctx)
	if s := gotpl.InjectFile(ctx); s != "" {
		buf, err := ioutil.ReadFile(s)
		if err != nil {
			return nil, fmt.Errorf("unable to read file: %v", err)
		}
		inject = string(buf)
	}
	driver, _, schema := xo.DriverDbSchema(ctx)
	nth, err := loader.NthParam(ctx)
	if err != nil {
		return nil, err
	}
	funcs := &Funcs{
		driver:     driver,
		schema:     schema,
		nth:        nth,
		first:      first,
		pkg:        gotpl.Pkg(ctx),
		tags:       gotpl.Tags(ctx),
		imports:    gotpl.Imports(ctx),
		conflict:   gotpl.Conflict(ctx),
		custom:     gotpl.Custom(ctx),
		escSchema:  gotpl.Esc(ctx, "schema"),
		escTable:   gotpl.Esc(ctx, "table"),
		escColumn:  gotpl.Esc(ctx, "column"),
		fieldtag:   fieldtag,
		context:    gotpl.Context(ctx),
		inject:     inject,
		knownTypes: gotpl.KnownTypes(ctx),
		shorts:     gotpl.Shorts(ctx),
	}
	return funcs.FuncMap(), nil
}

// Funcs is a set of template funcs.
type Funcs struct {
	driver    string
	schema    string
	nth       func(int) string
	first     *bool
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
		// context
		"context":         f.contextfn,
		"context_both":    f.context_both,
		"context_disable": f.context_disable,
		// func and query
		"func_name_context":   f.func_name_context,
		"func_name":           f.func_name_none,
		"func_context":        f.func_context,
		"func":                f.func_none,
		"recv_context":        f.recv_context,
		"recv":                f.recv_none,
		"foreign_key_context": f.foreign_key_context,
		"foreign_key":         f.foreign_key_none,
		"db":                  f.db,
		"db_prefix":           f.db_prefix,
		"db_update":           f.db_update,
		"db_named":            f.db_named,
		"named":               f.named,
		"logf":                f.logf,
		"logf_pkeys":          f.logf_pkeys,
		"logf_update":         f.logf_update,
		// type
		"names":        f.names,
		"names_all":    f.names_all,
		"names_ignore": f.names_ignore,
		"params":       f.params,
		"zero":         f.zero,
		"type":         f.typefn,
		"field":        f.field,
		"short":        f.short,
		// sqlstr funcs
		"querystr": f.querystr,
		"sqlstr":   f.sqlstr,
		// helpers
		"check_name": checkName,
		"eval":       eval,
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
			names[i] = escfn(name)
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
			s = escfn(s)
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
func (f *Funcs) importsfn() []gotpl.PackageImport {
	var imports []gotpl.PackageImport
	for _, s := range f.imports {
		alias, pkg := "", s
		if i := strings.Index(pkg, " "); i != -1 {
			alias, pkg = pkg[:i], strings.TrimSpace(pkg[i:])
		}
		imports = append(imports, gotpl.PackageImport{
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

// func_name_none builds a func name.
func (f *Funcs) func_name_none(v interface{}) string {
	switch x := v.(type) {
	case string:
		return x
	case gotpl.Query:
		return x.Name
	case gotpl.Table:
		return x.GoName
	case gotpl.ForeignKey:
		return x.GoName
	case gotpl.Proc:
		n := x.GoName
		if x.Overloaded {
			n = x.OverloadedName
		}
		return n
	case gotpl.Index:
		return x.Func
	}
	return fmt.Sprintf("[[ UNSUPPORTED TYPE 1: %T ]]", v)
}

// func_name_context generates a name for the func.
func (f *Funcs) func_name_context(v interface{}) string {
	switch x := v.(type) {
	case string:
		return nameContext(f.context_both(), x)
	case gotpl.Query:
		return nameContext(f.context_both(), x.Name)
	case gotpl.Table:
		return nameContext(f.context_both(), x.GoName)
	case gotpl.ForeignKey:
		return nameContext(f.context_both(), x.GoName)
	case gotpl.Proc:
		n := x.GoName
		if x.Overloaded {
			n = x.OverloadedName
		}
		return nameContext(f.context_both(), n)
	case gotpl.Index:
		return nameContext(f.context_both(), x.Func)
	}
	return fmt.Sprintf("[[ UNSUPPORTED TYPE 2: %T ]]", v)
}

// funcfn builds a func definition.
func (f *Funcs) funcfn(name string, context bool, v interface{}) string {
	var p, r []string
	if context {
		p = append(p, "ctx context.Context")
	}
	p = append(p, "db DB")
	switch x := v.(type) {
	case gotpl.Query:
		// params
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
			r = append(r, "*"+x.Type.GoName)
		default:
			r = append(r, "[]*"+x.Type.GoName)
		}
	case gotpl.Proc:
		// params
		p = append(p, f.params(x.Params, true))
		// returns
		if !x.Void {
			for _, ret := range x.Returns {
				r = append(r, f.typefn(ret.Type))
			}
		}
	case gotpl.Index:
		// params
		p = append(p, f.params(x.Fields, true))
		// returns
		rt := "*" + x.Table.GoName
		if !x.IsUnique {
			rt = "[]" + rt
		}
		r = append(r, rt)
	default:
		return fmt.Sprintf("[[ UNSUPPORTED TYPE 3: %T ]]", v)
	}
	r = append(r, "error")
	return fmt.Sprintf("func %s(%s) (%s)", name, strings.Join(p, ", "), strings.Join(r, ", "))
}

// func_context generates a func signature for v with context determined by the
// context mode.
func (f *Funcs) func_context(v interface{}) string {
	return f.funcfn(f.func_name_context(v), f.contextfn(), v)
}

// func_none genarates a func signature for v without context.
func (f *Funcs) func_none(v interface{}) string {
	return f.funcfn(f.func_name_none(v), false, v)
}

// recv builds a receiver func definition.
func (f *Funcs) recv(name string, context bool, t gotpl.Table, v interface{}) string {
	short := f.short(t)
	var p, r []string
	// determine params and return type
	if context {
		p = append(p, "ctx context.Context")
	}
	p = append(p, "db DB")
	switch x := v.(type) {
	case gotpl.ForeignKey:
		r = append(r, "*"+x.RefTable)
	}
	r = append(r, "error")
	return fmt.Sprintf("func (%s *%s) %s(%s) (%s)", short, t.GoName, name, strings.Join(p, ", "), strings.Join(r, ", "))
}

// recv_context builds a receiver func definition with context determined by
// the context mode.
func (f *Funcs) recv_context(typ interface{}, v interface{}) string {
	switch x := typ.(type) {
	case gotpl.Table:
		return f.recv(f.func_name_context(v), f.contextfn(), x, v)
	}
	return fmt.Sprintf("[[ UNSUPPORTED TYPE 4: %T ]]", typ)
}

// recv_none builds a receiver func definition without context.
func (f *Funcs) recv_none(typ interface{}, v interface{}) string {
	switch x := typ.(type) {
	case gotpl.Table:
		return f.recv(f.func_name_none(v), false, x, v)
	}
	return fmt.Sprintf("[[ UNSUPPORTED TYPE 5: %T ]]", typ)
}

func (f *Funcs) foreign_key_context(v interface{}) string {
	var name string
	var p []string
	if f.contextfn() {
		p = append(p, "ctx")
	}
	switch x := v.(type) {
	case gotpl.ForeignKey:
		name = x.RefFunc
		if f.context_both() {
			name += "Context"
		}
		// add params
		p = append(p, "db", f.convertTypes(x))
	default:
		return fmt.Sprintf("[[ UNSUPPORTED TYPE 6: %T ]]", v)
	}
	return fmt.Sprintf("%s(%s)", name, strings.Join(p, ", "))
}

func (f *Funcs) foreign_key_none(v interface{}) string {
	var name string
	var p []string
	switch x := v.(type) {
	case gotpl.ForeignKey:
		name = x.RefFunc
		p = append(p, "context.Background()", "db", f.convertTypes(x))
	default:
		return fmt.Sprintf("[[ UNSUPPORTED TYPE 7: %T ]]", v)
	}
	return fmt.Sprintf("%s(%s)", name, strings.Join(p, ", "))
}

// db generates a db.<name>Context(ctx, sqlstr, ...)
func (f *Funcs) db(name string, v ...interface{}) string {
	// params
	var p []interface{}
	if f.contextfn() {
		name += "Context"
		p = append(p, "ctx")
	}
	p = append(p, "sqlstr")
	return fmt.Sprintf("db.%s(%s)", name, f.names("", append(p, v...)...))
}

// db_prefix generates a db.<name>Context(ctx, sqlstr, <prefix>.param, ...).
//
// Will skip the specific parameters based on the type provided.
func (f *Funcs) db_prefix(name string, skip bool, vs ...interface{}) string {
	var prefix string
	var params []interface{}
	for i, v := range vs {
		var ignore []string
		switch x := v.(type) {
		case string:
			params = append(params, x)
		case gotpl.Table:
			prefix = f.short(x.GoName) + "."
			// skip primary keys
			if skip {
				for _, field := range x.Fields {
					if field.IsSequence {
						ignore = append(ignore, field.GoName)
					}
				}
			}
			p := f.names_ignore(prefix, v, ignore...)
			// p is "" when no columns are present except for primary key
			// params
			if p != "" {
				params = append(params, p)
			}
		default:
			return fmt.Sprintf("[[ UNSUPPORTED TYPE 8 (%d): %T ]]", i, v)
		}
	}
	return f.db(name, params...)
}

// db_update generates a db.<name>Context(ctx, sqlstr, regularparams,
// primaryparams)
func (f *Funcs) db_update(name string, v interface{}) string {
	var ignore, p []string
	switch x := v.(type) {
	case gotpl.Table:
		prefix := f.short(x.GoName) + "."
		for _, pk := range x.PrimaryKeys {
			ignore = append(ignore, pk.GoName)
		}
		p = append(p, f.names_ignore(prefix, x, ignore...), f.names(prefix, x.PrimaryKeys))
	default:
		return fmt.Sprintf("[[ UNSUPPORTED TYPE 9: %T ]]", v)
	}
	return f.db(name, strings.Join(p, ", "))
}

// db_named generates a db.<name>Context(ctx, sql.Named(name, res)...)
func (f *Funcs) db_named(name string, v interface{}) string {
	var p []string
	switch x := v.(type) {
	case gotpl.Proc:
		for _, z := range x.Params {
			p = append(p, f.named(z.SQLName, z.GoName, false))
		}
		for _, z := range x.Returns {
			p = append(p, f.named(z.SQLName, "&"+z.GoName, true))
		}
	default:
		return fmt.Sprintf("[[ UNSUPPORTED TYPE 10: %T ]]", v)
	}
	return f.db(name, strings.Join(p, ", "))
}

func (f *Funcs) named(name, value string, out bool) string {
	if out {
		return fmt.Sprintf("sql.Named(%q, sql.Out{Dest: %s})", name, value)
	}
	return fmt.Sprintf("sql.Named(%q, %s)", name, value)
}

func (f *Funcs) logf_pkeys(v interface{}) string {
	p := []string{"sqlstr"}
	switch x := v.(type) {
	case gotpl.Table:
		p = append(p, f.names(f.short(x.GoName)+".", x.PrimaryKeys))
	}
	return fmt.Sprintf("logf(%s)", strings.Join(p, ", "))
}

func (f *Funcs) logf(v interface{}, ignore ...interface{}) string {
	var ignoreNames []string
	p := []string{"sqlstr"}
	// build ignore list
	for i, x := range ignore {
		switch z := x.(type) {
		case string:
			ignoreNames = append(ignoreNames, z)
		case gotpl.Field:
			ignoreNames = append(ignoreNames, z.GoName)
		case []gotpl.Field:
			for _, f := range z {
				ignoreNames = append(ignoreNames, f.GoName)
			}
		default:
			return fmt.Sprintf("[[ UNSUPPORTED TYPE 11 (%d): %T ]]", i, x)
		}
	}
	// add fields
	switch x := v.(type) {
	case gotpl.Table:
		p = append(p, f.names_ignore(f.short(x.GoName)+".", x, ignoreNames...))
	default:
		return fmt.Sprintf("[[ UNSUPPORTED TYPE 12: %T ]]", v)
	}
	return fmt.Sprintf("logf(%s)", strings.Join(p, ", "))
}

func (f *Funcs) logf_update(v interface{}) string {
	var ignore []string
	p := []string{"sqlstr"}
	switch x := v.(type) {
	case gotpl.Table:
		prefix := f.short(x.GoName) + "."
		for _, pk := range x.PrimaryKeys {
			ignore = append(ignore, pk.GoName)
		}
		p = append(p, f.names_ignore(prefix, x, ignore...), f.names(prefix, x.PrimaryKeys))
	default:
		return fmt.Sprintf("[[ UNSUPPORTED TYPE 13: %T ]]", v)
	}
	return fmt.Sprintf("logf(%s)", strings.Join(p, ", "))
}

// names generates a list of names.
func (f *Funcs) namesfn(all bool, prefix string, z ...interface{}) string {
	var names []string
	for i, v := range z {
		switch x := v.(type) {
		case string:
			names = append(names, x)
		case gotpl.Query:
			for _, p := range x.Params {
				if !all && p.Interpolate {
					continue
				}
				names = append(names, prefix+p.Name)
			}
		case gotpl.Table:
			for _, p := range x.Fields {
				names = append(names, prefix+checkName(p.GoName))
			}
		case []gotpl.Field:
			for _, p := range x {
				names = append(names, prefix+checkName(p.GoName))
			}
		case gotpl.Proc:
			if params := f.params(x.Params, false); params != "" {
				names = append(names, params)
			}
		case gotpl.Index:
			names = append(names, f.params(x.Fields, false))
		default:
			names = append(names, fmt.Sprintf("/* UNSUPPORTED TYPE 14 (%d): %T */", i, v))
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

// names_all generates a list of all names, ignoring fields that match the value in ignore.
func (f *Funcs) names_ignore(prefix string, v interface{}, ignore ...string) string {
	m := map[string]bool{}
	for _, n := range ignore {
		m[n] = true
	}
	var vals []gotpl.Field
	switch x := v.(type) {
	case gotpl.Table:
		for _, p := range x.Fields {
			if m[p.GoName] {
				continue
			}
			vals = append(vals, p)
		}
	case []gotpl.Field:
		for _, p := range x {
			if m[p.GoName] {
				continue
			}
			vals = append(vals, p)
		}
	default:
		return fmt.Sprintf("[[ UNSUPPORTED TYPE 15: %T ]]", v)
	}
	return f.namesfn(true, prefix, vals)
}

// querystr generates a querystr for the specified query and any accompanying
// comments.
func (f *Funcs) querystr(v interface{}) string {
	var interpolate bool
	var query, comments []string
	switch x := v.(type) {
	case gotpl.Query:
		interpolate, query, comments = x.Interpolate, x.Query, x.Comments
	default:
		return fmt.Sprintf("const sqlstr = [[ UNSUPPORTED TYPE 16: %T ]]", v)
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

func (f *Funcs) sqlstr(typ string, v interface{}) string {
	var lines []string
	switch typ {
	case "insert_manual":
		lines = f.sqlstr_insert_manual(v)
	case "insert":
		lines = f.sqlstr_insert(v)
	case "update":
		lines = f.sqlstr_update(v)
	case "upsert":
		lines = f.sqlstr_upsert(v)
	case "delete":
		lines = f.sqlstr_delete(v)
	case "proc":
		lines = f.sqlstr_proc(v)
	case "index":
		lines = f.sqlstr_index(v)
	default:
		return fmt.Sprintf("const sqlstr = `UNKNOWN QUERY TYPE: %s`", typ)
	}
	return fmt.Sprintf("const sqlstr = `%s`", strings.Join(lines, "` +\n\t`"))
}

// sqlstr_insert_manual builds an INSERT query
// If not all, sequence columns are skipped.
func (f *Funcs) sqlstr_insert_base(all bool, v interface{}) []string {
	switch x := v.(type) {
	case gotpl.Table:
		// build names and values
		var fields, vals []string
		var i int
		for _, z := range x.Fields {
			if z.IsSequence && !all {
				continue
			}
			fields, vals = append(fields, f.colname(z)), append(vals, f.nth(i))
			i++
		}
		return []string{
			"INSERT INTO " + f.schemafn(x.SQLName) + " (",
			strings.Join(fields, ", "),
			") VALUES (",
			strings.Join(vals, ", "),
			")",
		}
	}
	return []string{fmt.Sprintf("[[ UNSUPPORTED TYPE 17: %T ]]", v)}
}

// sqlstr_insert_manual builds an INSERT query that inserts all fields.
func (f *Funcs) sqlstr_insert_manual(v interface{}) []string {
	return f.sqlstr_insert_base(true, v)
}

// sqlstr_insert builds an INSERT query, skipping the sequence field with
// applicable RETURNING clause for generated primary key fields.
func (f *Funcs) sqlstr_insert(v interface{}) []string {
	switch x := v.(type) {
	case gotpl.Table:
		var seq gotpl.Field
		for _, field := range x.Fields {
			if field.IsSequence {
				seq = field
			}
		}
		lines := f.sqlstr_insert_base(false, v)
		// add return clause
		switch f.driver {
		case "oracle":
			lines[len(lines)-1] += ` RETURNING ` + f.colname(seq) + ` /*LASTINSERTID*/ INTO :pk`
		case "postgres":
			lines[len(lines)-1] += ` RETURNING ` + f.colname(seq)
		case "sqlserver":
			lines[len(lines)-1] += "; SELECT ID = CONVERT(BIGINT, SCOPE_IDENTITY())"
		}
		return lines
	}
	return []string{fmt.Sprintf("[[ UNSUPPORTED TYPE 18: %T ]]", v)}
}

// sqlstr_update_base builds an UPDATE query, using primary key fields as the WHERE
// clause, adding prefix.
//
// When prefix is empty, the WHERE clause will be in the form of name = $1.
// When prefix is non-empty, the WHERE clause will be in the form of name = <PREFIX>name.
//
// Similarly, when prefix is empty, the table's name is added after UPDATE,
// otherwise it is omitted.
func (f *Funcs) sqlstr_update_base(prefix string, v interface{}) (int, []string) {
	switch x := v.(type) {
	case gotpl.Table:
		// build names and values
		var i int
		var list []string
		for _, z := range x.Fields {
			if z.IsPrimary {
				continue
			}
			name, param := f.colname(z), f.nth(i)
			if prefix != "" {
				param = prefix + name
			}
			list = append(list, fmt.Sprintf("%s = %s", name, param))
			i++
		}
		name := ""
		if prefix == "" {
			name = f.schemafn(x.SQLName) + " "
		}
		return i, []string{
			"UPDATE " + name + "SET ",
			strings.Join(list, ", ") + " ",
		}
	}
	return 0, []string{fmt.Sprintf("[[ UNSUPPORTED TYPE 19: %T ]]", v)}
}

// sqlstr_update builds an UPDATE query, using primary key fields as the WHERE
// clause.
func (f *Funcs) sqlstr_update(v interface{}) []string {
	// build pkey vals
	switch x := v.(type) {
	case gotpl.Table:
		var list []string
		n, lines := f.sqlstr_update_base("", v)
		for i, z := range x.PrimaryKeys {
			list = append(list, fmt.Sprintf("%s = %s", f.colname(z), f.nth(n+i)))
		}
		return append(lines, "WHERE "+strings.Join(list, ", "))
	}
	return []string{fmt.Sprintf("[[ UNSUPPORTED TYPE 20: %T ]]", v)}
}

func (f *Funcs) sqlstr_upsert(v interface{}) []string {
	switch x := v.(type) {
	case gotpl.Table:
		// build insert
		lines := f.sqlstr_insert_base(true, x)
		switch f.driver {
		case "postgres", "sqlite3":
			return append(lines, f.sqlstr_upsert_postgres_sqlite(x)...)
		case "mysql":
			return append(lines, f.sqlstr_upsert_mysql(x)...)
		case "sqlserver", "oracle":
			return f.sqlstr_upsert_sqlserver_oracle(x)
		}
	}
	return []string{fmt.Sprintf("[[ UNSUPPORTED TYPE 21 %s: %T ]]", f.driver, v)}
}

// sqlstr_upsert_postgres_sqlite builds an uspert query for postgres and sqlite
//
// INSERT (..) VALUES (..) ON CONFLICT DO UPDATE SET ...
func (f *Funcs) sqlstr_upsert_postgres_sqlite(v interface{}) []string {
	switch x := v.(type) {
	case gotpl.Table:
		// add conflict and update
		var conflicts []string
		for _, f := range x.PrimaryKeys {
			conflicts = append(conflicts, f.SQLName)
		}
		lines := []string{" ON CONFLICT (" + strings.Join(conflicts, ", ") + ") DO "}
		_, update := f.sqlstr_update_base("EXCLUDED.", v)
		return append(lines, update...)
	}
	return []string{fmt.Sprintf("[[ UNSUPPORTED TYPE 22: %T ]]", v)}
}

// sqlstr_upsert_mysql builds an uspert query for mysql
//
// INSERT (..) VALUES (..) ON DUPLICATE KEY UPDATE SET ...
func (f *Funcs) sqlstr_upsert_mysql(v interface{}) []string {
	switch x := v.(type) {
	case gotpl.Table:
		lines := []string{" ON DUPLICATE KEY UPDATE "}
		var list []string
		i := len(x.Fields)
		for _, z := range x.Fields {
			if z.IsSequence {
				continue
			}
			name := f.colname(z)
			list = append(list, fmt.Sprintf("%s = VALUES(%s)", name, name))
			i++
		}
		return append(lines, strings.Join(list, ", "))
	}
	return []string{fmt.Sprintf("[[ UNSUPPORTED TYPE 23: %T ]]", v)}
}

// sqlstr_upsert_sqlserver_oracle builds an upsert query for sqlserver
//
// MERGE [table] AS target USING (SELECT [pkeys]) AS source ...
func (f *Funcs) sqlstr_upsert_sqlserver_oracle(v interface{}) []string {
	switch x := v.(type) {
	case gotpl.Table:
		var lines []string
		// merge [table]...
		switch f.driver {
		case "sqlserver":
			lines = []string{"MERGE " + f.schemafn(x.SQLName) + " AS t "}
		case "oracle":
			lines = []string{"MERGE " + f.schemafn(x.SQLName) + "t "}
		}
		// using (select ..)
		var fields, predicate []string
		for i, field := range x.Fields {
			fields = append(fields, fmt.Sprintf("%s %s", f.nth(i), field.SQLName))
		}
		for _, field := range x.PrimaryKeys {
			predicate = append(predicate, fmt.Sprintf("s.%s = t.%s", field.SQLName, field.SQLName))
		}
		// closing part for select
		var closing string
		switch f.driver {
		case "sqlserver":
			closing = `) AS s `
		case "oracle":
			closing = `FROM DUAL ) s `
		}
		lines = append(lines, `USING (`,
			`SELECT `+strings.Join(fields, ", ")+" ",
			closing,
			`ON `+strings.Join(predicate, " AND ")+" ")
		// build param lists
		var updateParams, insertParams, insertVals []string
		for _, field := range x.Fields {
			// sequences are always managed by db
			if field.IsSequence {
				continue
			}
			// primary keys
			if !field.IsPrimary {
				updateParams = append(updateParams, fmt.Sprintf("t.%s = s.%s", field.SQLName, field.SQLName))
			}
			insertParams = append(insertParams, field.SQLName)
			insertVals = append(insertVals, "s."+field.SQLName)
		}
		// when matched then update...
		lines = append(lines,
			`WHEN MATCHED THEN `, `UPDATE SET `,
			strings.Join(updateParams, ", ")+" ",
			`WHEN NOT MATCHED THEN `,
			`INSERT (`,
			strings.Join(insertParams, ", "),
			`) VALUES (`,
			strings.Join(insertVals, ", "),
			`);`,
		)
		return lines
	}
	return []string{fmt.Sprintf("[[ UNSUPPORTED TYPE 24: %T ]]", v)}
}

// sqlstr_delete builds a DELETE query for the primary keys.
func (f *Funcs) sqlstr_delete(v interface{}) []string {
	switch x := v.(type) {
	case gotpl.Table:
		// names and values
		var list []string
		for i, z := range x.PrimaryKeys {
			list = append(list, fmt.Sprintf("%s = %s", f.colname(z), f.nth(i)))
		}
		return []string{
			"DELETE FROM " + f.schemafn(x.SQLName) + " ",
			"WHERE " + strings.Join(list, " AND "),
		}
	}
	return []string{fmt.Sprintf("[[ UNSUPPORTED TYPE 25: %T ]]", v)}
}

// sqlstr_index builds a
func (f *Funcs) sqlstr_index(v interface{}) []string {
	switch x := v.(type) {
	case gotpl.Index:
		// build table fieldnames
		var fields []string
		for _, z := range x.Table.Fields {
			fields = append(fields, f.colname(z))
		}
		// index fields
		var list []string
		for i, z := range x.Fields {
			list = append(list, fmt.Sprintf("%s = %s", f.colname(z), f.nth(i)))
		}
		return []string{
			"SELECT ",
			strings.Join(fields, ", ") + " ",
			"FROM " + f.schemafn(x.Table.SQLName) + " ",
			"WHERE " + strings.Join(list, " AND "),
		}
	}
	return []string{fmt.Sprintf("[[ UNSUPPORTED TYPE 26: %T ]]", v)}
}

// sqlstr_proc builds a stored procedure call.
func (f *Funcs) sqlstr_proc(v interface{}) []string {
	switch x := v.(type) {
	case gotpl.Proc:
		if x.Type == "function" {
			return f.sqlstr_func(v)
		}
		// sql string format
		var format string
		switch f.driver {
		case "postgres", "mysql":
			format = "CALL %s(%s)"
		case "sqlserver":
			format = "%[1]s"
		case "oracle":
			format = "BEGIN %s(%s); END;"
		}
		// build params list; add return fields for orcle
		l := x.Params
		if f.driver == "oracle" {
			l = append(l, x.Returns...)
		}
		var list []string
		for i, field := range l {
			s := f.nth(i)
			if f.driver == "oracle" {
				s = ":" + field.SQLName
			}
			list = append(list, s)
		}
		// dont prefix with schema for oracle
		name := f.schemafn(x.SQLName)
		if f.driver == "oracle" {
			name = x.SQLName
		}
		return []string{
			fmt.Sprintf(format, name, strings.Join(list, ", ")),
		}
	}
	return []string{fmt.Sprintf("[[ UNSUPPORTED TYPE 27: %T ]]", v)}
}

func (f *Funcs) sqlstr_func(v interface{}) []string {
	switch x := v.(type) {
	case gotpl.Proc:
		var format string
		switch f.driver {
		case "postgres":
			format = "SELECT * FROM %s(%s)"
		case "mysql":
			format = "SELECT %s(%s)"
		case "sqlserver":
			format = "SELECT %s(%s) AS OUT"
		case "oracle":
			format = "SELECT %s(%s) FROM dual"
		}
		var list []string
		l := x.Params
		for i := range l {
			list = append(list, f.nth(i))
		}
		return []string{
			fmt.Sprintf(format, f.schemafn(x.SQLName), strings.Join(list, ", ")),
		}
	}
	return []string{fmt.Sprintf("[[ UNSUPPORTED TYPE 28: %T ]]", v)}
}

// convertTypes generates the conversions to convert the foreign key field
// types to their respective referenced field types.
func (f *Funcs) convertTypes(fkey gotpl.ForeignKey) string {
	var p []string
	for i := range fkey.Fields {
		field := fkey.Fields[i]
		refField := fkey.RefFields[i]
		expr := f.short(fkey.Table) + "." + field.GoName
		// types match, can match
		if field.Type == refField.Type {
			p = append(p, expr)
			continue
		}
		// convert types
		typ, refType := field.Type, refField.Type
		if strings.HasPrefix(typ, "sql.Null") {
			expr = expr + "." + typ[8:]
			typ = strings.ToLower(typ[8:])
		}
		if strings.ToLower(refType) != typ {
			expr = refType + "(" + expr + ")"
		}
		p = append(p, expr)
	}
	return strings.Join(p, ", ")
}

// params converts a list of fields into their named Go parameters, skipping
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
func (f *Funcs) params(fields []gotpl.Field, addType bool) string {
	var vals []string
	for _, field := range fields {
		vals = append(vals, f.param(field, addType))
	}
	return strings.Join(vals, ", ")
}

func (f *Funcs) param(field gotpl.Field, addType bool) string {
	n := strings.Split(snaker.CamelToSnake(field.GoName), "_")
	s := strings.ToLower(n[0]) + field.GoName[len(n[0]):]
	// check go reserved names
	if r, ok := goReservedNames[strings.ToLower(s)]; ok {
		s = r
	}
	// add the go type
	if addType {
		s += " " + f.typefn(field.Type)
	}
	// add to vals
	return s
}

// zero generates a zero list.
func (f *Funcs) zero(z ...interface{}) string {
	var zeroes []string
	for i, v := range z {
		switch x := v.(type) {
		case string:
			zeroes = append(zeroes, x)
		case gotpl.Table:
			for _, p := range x.Fields {
				zeroes = append(zeroes, f.zero(p))
			}
		case []gotpl.Field:
			for _, p := range x {
				zeroes = append(zeroes, f.zero(p))
			}
		case gotpl.Field:
			if _, ok := f.knownTypes[x.Type]; ok || x.Zero == "nil" {
				zeroes = append(zeroes, x.Zero)
				break
			}
			zeroes = append(zeroes, f.typefn(x.Type)+"{}")
		default:
			zeroes = append(zeroes, fmt.Sprintf("/* UNSUPPORTED TYPE 29 (%d): %T */", i, v))
		}
	}
	return strings.Join(zeroes, ", ")
}

// typefn generates the Go type, prefixing the custom package name if applicable.
func (f *Funcs) typefn(typ string) string {
	if strings.Contains(typ, ".") {
		return typ
	}
	var prefix string
	for strings.HasPrefix(typ, "[]") {
		typ = typ[2:]
		prefix += "[]"
	}
	if _, ok := f.knownTypes[typ]; ok || f.custom == "" {
		return prefix + typ
	}
	return prefix + f.custom + "." + typ
}

// field generates a field definition for a struct.
func (f *Funcs) field(field gotpl.Field) (string, error) {
	tag := ""
	buf := new(bytes.Buffer)
	if err := f.fieldtag.Funcs(f.FuncMap()).Execute(buf, field); err != nil {
		return "", err
	}
	if s := buf.String(); s != "" {
		tag = " " + s
	}
	s := fmt.Sprintf("\t%s %s%s %s", field.GoName, f.typefn(field.Type), tag, "// "+field.SQLName)
	return s, nil
}

// short generates a safe Go identifier for typ. typ is first checked
// against shorts, and if not found, then the value is calculated and
// stored in the shorts for future use.
//
// A short is the concatenation of the lowercase of the first character in
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
	case string:
		n = x
	case gotpl.Table:
		n = x.GoName
	default:
		return fmt.Sprintf("[[ UNSUPPORTED TYPE 30: %T ]]", v)
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
		// ensure no name conflict
		name = checkName(strings.Join(u, ""))
		// store back to short name map
		f.shorts[n] = name
	}
	// append suffix if conflict exists
	if _, ok := templateReservedNames[name]; ok {
		name += f.conflict
	}
	return name
}

// colname returns the ColumnName of a field escaped if needed.
func (f *Funcs) colname(z gotpl.Field) string {
	if f.escColumn {
		return escfn(z.SQLName)
	}
	return z.SQLName
}

func checkName(name string) string {
	if n, ok := goReservedNames[name]; ok {
		return n
	}
	return name
}

// escfn escapes s.
func escfn(s string) string {
	return `"` + s + `"`
}

// eval evalutates a template s against v.
func eval(v interface{}, s string) (string, error) {
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

// templateReservedNames are the template reserved names.
var templateReservedNames = map[string]bool{
	// variables
	"ctx":  true,
	"db":   true,
	"err":  true,
	"log":  true,
	"logf": true,
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
