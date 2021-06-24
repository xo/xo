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
		"func_name_context":   f.func_name_context,
		"func_name":           f.func_name,
		"func_context":        f.func_context,
		"func":                f.func_none,
		"recv_context":        f.recv_context,
		"recv":                f.recv_none,
		"foreign_key_context": f.foreign_key_context,
		"foreign_key":         f.foreign_key,
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
			names[i] = f.escfn(name)
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
			s = f.escfn(s)
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
	switch x := v.(type) {
	case string:
		return x
	case *templates.Query:
		return x.Name
	case *templates.Type:
		return x.Name
	case *templates.ForeignKey:
		return x.Name
	case *templates.Proc:
		return x.Name
	case *templates.Index:
		return x.FuncName
	default:
		return fmt.Sprintf("[[ UNSUPPORTED TYPE: %T ]]", v)
	}
}

// func_name_context generates a name for the func.
func (f *Funcs) func_name_context(v interface{}) string {
	var name string
	switch x := v.(type) {
	case string:
		return nameContext(f.context_both(), x)
	case *templates.Query:
		name = nameContext(f.context_both(), x.Name)
	case *templates.Type:
		name = nameContext(f.context_both(), x.Name)
	case *templates.ForeignKey:
		name = nameContext(f.context_both(), x.Name)
	case *templates.Proc:
		name = nameContext(f.context_both(), x.Name)
	case *templates.Index:
		name = nameContext(f.context_both(), x.FuncName)
	default:
		return fmt.Sprintf("[[ UNSUPPORTED TYPE: %T ]]", v)
	}
	return name
}

// funcfn builds a func definition.
func (f *Funcs) funcfn(name string, context bool, v interface{}) string {
	var p, r []string
	if context {
		p = append(p, "ctx context.Context")
	}
	p = append(p, "db DB")
	switch x := v.(type) {
	case *templates.Query:
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
			r = append(r, "*"+x.Type.Name)
		default:
			r = append(r, "[]*"+x.Type.Name)
		}
	case *templates.Proc:
		// params
		p = append(p, f.params(x.Params, true))
		// returns
		if x.Return.Type != "void" {
			r = append(r, f.typefn(x.Return.Type))
		}
	case *templates.Index:
		// params
		p = append(p, f.params(x.Fields, true))
		// returns
		rt := "*" + x.Type.Name
		if !x.Index.IsUnique {
			rt = "[]" + rt
		}
		r = append(r, rt)
	default:
		return fmt.Sprintf("[[ UNSUPPORTED TYPE: %T ]]", v)
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
	return f.funcfn(f.func_name(v), false, v)
}

// recv builds a receiver func definition.
func (f *Funcs) recv(name string, context bool, typ interface{}, v interface{}) string {
	short := f.short(typ)
	var typeName string
	switch x := typ.(type) {
	case *templates.Type:
		typeName = x.Name
	default:
		return fmt.Sprintf("[[ UNSUPPORTED RECEIVER TYPE: %T ]]", typ)
	}
	var p, r []string
	// determine params and return type
	if context {
		p = append(p, "ctx context.Context")
	}
	p = append(p, "db DB")
	switch x := v.(type) {
	case *templates.ForeignKey:
		r = append(r, "*"+x.RefType.Name)
	}
	r = append(r, "error")
	return fmt.Sprintf("func (%s *%s) %s(%s) (%s)", short, typeName, name, strings.Join(p, ", "), strings.Join(r, ", "))
}

// recv_context builds a receiver func definition with context determined by
// the context mode.
func (f *Funcs) recv_context(typ interface{}, v interface{}) string {
	return f.recv(f.func_name_context(v), f.contextfn(), typ, v)
}

// recv_none builds a receiver func definition without context.
func (f *Funcs) recv_none(typ interface{}, v interface{}) string {
	return f.recv(f.func_name(v), false, typ, v)
}

func (f *Funcs) foreign_key_context(v interface{}) string {
	var name string
	var p []string
	if f.contextfn() {
		p = append(p, "ctx")
	}
	switch x := v.(type) {
	case *templates.ForeignKey:
		var ctx string
		if f.context_both() {
			ctx = "Context"
		}
		name = fmt.Sprintf("%sBy%s%s", x.RefType.Name, x.RefField.Name, ctx)
		// add params
		p = append(p, "db", f.convext(x.Type, x.Field, x.RefField))
	default:
		return fmt.Sprintf("[[ UNSUPPORTED TYPE %T ]]", v)
	}
	return fmt.Sprintf("%s(%s)", name, strings.Join(p, ", "))
}

func (f *Funcs) foreign_key(v interface{}) string {
	var name string
	var p []string
	switch x := v.(type) {
	case *templates.ForeignKey:
		name = fmt.Sprintf("%sBy%sContext", x.RefType.Name, x.RefField.Name)
		p = append(p, "context.Background()", "db", f.convext(x.Type, x.Field, x.RefField))
	default:
		return fmt.Sprintf("[[ UNSUPPORTED TYPE %T ]]", v)
	}
	return fmt.Sprintf("%s(%s)", name, strings.Join(p, ", "))
}

// db generates a db.<name>Context(ctx, sqlstr, ...)
func (f *Funcs) db(name string, v ...interface{}) string {
	// params
	var p []interface{}
	if f.contextfn() {
		name, p = name+"Context", append(p, "ctx")
	}
	p = append(p, "sqlstr")
	return fmt.Sprintf(`db.%s(%s)`, name, f.names("", append(p, v...)...))
}

// db_prefix generates a db.<name>Context(ctx, sqlstr, <prefix>.param, ...).
//
// Will skip the specific parameters based
// on the type provided.
func (f *Funcs) db_prefix(name string, skip bool, vs ...interface{}) string {
	var prefix string
	var params []interface{}
	for i, v := range vs {
		var ignore []string
		switch x := v.(type) {
		case string:
			params = append(params, x)
		case *templates.Type:
			prefix = f.short(x.Name) + "."
			// skip primary keys
			if skip {
				ignore = append(ignore, x.PrimaryKey.Name)
			}
			params = append(params, f.names_ignore(prefix, v, ignore...))
		default:
			return fmt.Sprintf("[[ UNSUPPORTED TYPE %d: %T ]]", i, v)
		}
	}
	return f.db(name, params...)
}

// db_update generates a db.<name>Context(ctx, sqlstr, regularparams,
// primaryparams)
func (f *Funcs) db_update(name string, v interface{}) string {
	var p []string
	switch x := v.(type) {
	case *templates.Type:
		prefix := f.short(x.Name) + "."
		p = append(p, f.names_ignore(prefix, x, x.PrimaryKey.Name), f.names(prefix, x.PrimaryKeyFields))
	default:
		return fmt.Sprintf("[[ UNSUPPORTED TYPE: %T ]]", v)
	}
	return f.db(name, strings.Join(p, ", "))
}

// db_named generates a db.<name>Context(ctx, sql.Named(name, res)...)
func (f *Funcs) db_named(name string, v interface{}) string {
	var p []string
	switch x := v.(type) {
	case *templates.Proc:
		for _, z := range x.Params {
			p = append(p, f.named(z.Name, f.param(z, false), false))
		}
		p = append(p, f.named(x.Proc.ReturnName, "&"+f.short(x.Return.Type), true))
	default:
		return fmt.Sprintf("[[ UNSUPPORTED TYPE: %T ]]", v)
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
	case *templates.Type:
		p = append(p, f.names(f.short(x.Name)+".", x.PrimaryKeyFields))
	}
	return fmt.Sprintf("logf(%s)", strings.Join(p, ", "))
}

func (f *Funcs) logf(v interface{}, ignore ...string) string {
	p := []string{"sqlstr"}
	switch x := v.(type) {
	case *templates.Type:
		p = append(p, f.names_ignore(f.short(x.Name)+".", x, ignore...))
	default:
		return fmt.Sprintf("[[ UNSUPPORTED TYPE: %T ]]", v)
	}
	return fmt.Sprintf("logf(%s)", strings.Join(p, ", "))
}

func (f *Funcs) logf_update(v interface{}) string {
	p := []string{"sqlstr"}
	switch x := v.(type) {
	case *templates.Type:
		prefix := f.short(x.Name) + "."
		p = append(p, f.names_ignore(prefix, x, x.PrimaryKey.Name), f.names(prefix, x.PrimaryKeyFields))
	default:
		return fmt.Sprintf("[[ UNSUPPORTED TYPE: %T ]]", v)
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
		case *templates.Proc:
			names = append(names, f.params(x.Params, false))
		case *templates.Index:
			names = append(names, f.params(x.Fields, false))
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

// names_all generates a list of all names, ignoring fields that match the value in ignore.
func (f *Funcs) names_ignore(prefix string, v interface{}, ignore ...string) string {
	m := map[string]bool{}
	for _, n := range ignore {
		m[n] = true
	}
	var vals []*templates.Field
	switch x := v.(type) {
	case *templates.Type:
		for _, p := range x.Fields {
			if m[p.Name] {
				continue
			}
			vals = append(vals, p)
		}
	case []*templates.Field:
		for _, p := range x {
			if m[p.Name] {
				continue
			}
			vals = append(vals, p)
		}
	default:
		return fmt.Sprintf("[[ UNSUPPORTED TYPE %T ]]", v)
	}
	return f.namesfn(true, prefix, vals)
}

// querystr generates a querystr for the specified query and any accompanying
// comments.
func (f *Funcs) querystr(v interface{}) string {
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

func (f *Funcs) sqlstr(typ string, v interface{}) string {
	var sql string
	switch typ {
	case "insert_manual":
		sql = f.insert_manual(v)
	case "insert":
		sql = f.insert(v)
	case "update":
		sql = f.update(v)
	case "upsert":
		sql = f.upsert(v)
	case "delete":
		sql = f.delete(v)
	case "proc":
		sql = f.proc(v)
	case "index":
		sql = f.index(v)
	default:
		return fmt.Sprintf("const sqlstr =  `UNKNOWN sqlstr: %q`", typ)
	}
	split := strings.Split(sql, "|")
	return fmt.Sprintf("const sqlstr = `%s`", strings.Join(split, "` +\n`"))
}

func (f *Funcs) insert_manual(v interface{}) string {
	var table string
	var fields, vals []string
	switch x := v.(type) {
	case *templates.Type:
		table = f.schemafn(x.Table.TableName)
		// names and values
		for i, z := range x.Fields {
			fields = append(fields, f.colname(z))
			vals = append(vals, f.nth(i))
		}
	default:
		return fmt.Sprintf("[[ UNSUPPORTED TYPE: %T ]]", v)
	}
	return fmt.Sprintf("INSERT INTO %s (|%s|) VALUES (|%s|)",
		table, strings.Join(fields, ", "), strings.Join(vals, ", "))
}

func (f *Funcs) insert(v interface{}) string {
	var table, end string
	var fields, vals []string
	switch x := v.(type) {
	case *templates.Type:
		table = f.schemafn(x.Table.TableName)
		// names and values
		var i int
		for _, z := range x.Fields {
			if z.Col.IsPrimaryKey {
				continue
			}
			fields = append(fields, f.colname(z))
			vals = append(vals, f.nth(i))
			i++
		}
		// end
		switch f.driver {
		case "postgres":
			end = fmt.Sprintf(" RETURNING %s", f.colname(x.PrimaryKey))
		case "oracle":
			end = fmt.Sprintf(" RETURNING %s /*LASTINSERTID*/ INTO :pk", f.colname(x.PrimaryKey))
		case "sqlserver":
			end = "; select ID = convert(bigint, SCOPE_IDENTITY())"
		}
	default:
		return fmt.Sprintf("[[ UNSUPPORTED TYPE: %T ]]", v)
	}
	return fmt.Sprintf("INSERT INTO %s (|%s|) VALUES (|%s|)%s",
		table, strings.Join(fields, ", "), strings.Join(vals, ", "), end)
}

func (f *Funcs) update(v interface{}) string {
	var table string
	var fields, vals, pvals []string
	switch x := v.(type) {
	case *templates.Type:
		table = f.schemafn(x.Table.TableName)
		// names and values
		var i int
		for _, z := range x.Fields {
			if z.Col.IsPrimaryKey {
				continue
			}
			fields = append(fields, f.colname(z))
			vals = append(vals, f.nth(i))
			i++
		}
		// add values for pkey fields
		for j, z := range x.PrimaryKeyFields {
			pvals = append(pvals, fmt.Sprintf("%s = %s", f.colname(z), f.nth(i+j)))
		}
	default:
		return fmt.Sprintf("[[ UNSUPPORTED TYPE: %T ]]", v)
	}
	if f.driver == "postgres" {
		return fmt.Sprintf("UPDATE %s SET (|%s|) = (|%s|) WHERE %s",
			table, strings.Join(fields, ", "), strings.Join(vals, ", "), strings.Join(pvals, " AND "))
	}
	// col1 = $1, col2 = $2...
	for i, s := range fields {
		fields[i] = fmt.Sprintf("%s = %s", s, vals[i])
	}
	return fmt.Sprintf("UPDATE %s SET |%s| WHERE %s",
		table, strings.Join(fields, ", "), strings.Join(pvals, " AND "))
}

func (f *Funcs) upsert(v interface{}) string {
	var table string
	var fields, vals, pkeys, excluded []string
	switch x := v.(type) {
	case *templates.Type:
		table = f.schemafn(x.Table.TableName)
		// names and values
		for i, z := range x.Fields {
			if z.Col.IsPrimaryKey {
				pkeys = append(pkeys, f.colname(z))
			}
			fields = append(fields, f.colname(z))
			vals = append(vals, f.nth(i))
			excluded = append(excluded, "EXCLUDED."+f.colname(z))
		}
	default:
		return fmt.Sprintf("[[ UNSUPPORTED TYPE: %T ]]", v)
	}
	return fmt.Sprintf("INSERT INTO %s (|%s|) VALUES (|%s|) ON CONFLICT (%s) DO UPDATE SET (|%s|) = (|%s|)",
		table, strings.Join(fields, ", "), strings.Join(vals, ", "), strings.Join(pkeys, ", "),
		strings.Join(fields, ", "), strings.Join(excluded, ", "))
}

func (f *Funcs) delete(v interface{}) string {
	var table string
	var fields []string
	switch x := v.(type) {
	case *templates.Type:
		table = f.schemafn(x.Table.TableName)
		// names and values
		for i, z := range x.PrimaryKeyFields {
			fields = append(fields, fmt.Sprintf("%s = %s", f.colname(z), f.nth(i)))
		}
	default:
		return fmt.Sprintf("[[ UNSUPPORTED TYPE: %T ]]", v)
	}

	return fmt.Sprintf("DELETE FROM %s WHERE %s",
		table, strings.Join(fields, " AND "))
}

func (f *Funcs) proc(v interface{}) string {
	name, mask := "", "SELECT %s(%s)"
	var params []string
	switch x := v.(type) {
	case *templates.Proc:
		name = f.schemafn(x.Proc.ProcName)
		for i := range x.Params {
			params = append(params, f.nth(i))
		}
	default:
		return fmt.Sprintf("[[ UNSUPPORTED TYPE: %T ]]", v)
	}
	switch f.driver {
	case "sqlserver":
		return name
	}
	return fmt.Sprintf(mask, name, strings.Join(params, ", "))
}

func (f *Funcs) index(v interface{}) string {
	var table string
	var fields, vals []string
	switch x := v.(type) {
	case *templates.Index:
		table = f.schemafn(x.Type.Table.TableName)
		// table fieldnames
		for _, z := range x.Type.Fields {
			fields = append(fields, f.colname(z))
		}
		// index fields
		for i, z := range x.Fields {
			vals = append(vals, fmt.Sprintf("%s = %s", f.colname(z), f.nth(i)))
		}
	default:
		return fmt.Sprintf("[[ UNSUPPORTED TYPE: %T ]]", v)
	}
	return fmt.Sprintf("SELECT |%s |FROM %s |WHERE %s",
		strings.Join(fields, ", "), table, strings.Join(vals, " AND "))
}

// convext generates the Go conversion for a to be assignable to b.
func (f *Funcs) convext(v interface{}, a, b *templates.Field) string {
	expr := a.Name
	switch x := v.(type) {
	case *templates.Type:
		expr = f.short(x) + "." + expr
	default:
		return fmt.Sprintf("[[ UNSUPPORTED TYPE: %T ]]", v)
	}
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
func (f *Funcs) params(fields []*templates.Field, addType bool, ignore ...string) string {
	m := map[string]bool{}
	for _, n := range ignore {
		m[n] = true
	}
	var vals []string
	for _, field := range fields {
		if m[field.Name] {
			continue
		}
		// add to vals
		vals = append(vals, f.param(field, addType))
	}
	return strings.Join(vals, ", ")
}

func (f *Funcs) param(field *templates.Field, addType bool) string {
	n := strings.Split(snaker.CamelToSnake(field.Name), "_")
	s := strings.ToLower(n[0]) + field.Name[len(n[0]):]
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

// column returns the ColumnName of a field escaped if needed.
func (f *Funcs) colname(z *templates.Field) string {
	if f.escColumn {
		return f.escfn(z.Col.ColumnName)
	}
	return z.Col.ColumnName
}

// escfn escapes s.
func (f *Funcs) escfn(s string) string {
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
