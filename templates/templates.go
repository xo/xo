package templates

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
	"github.com/xo/xo/internal"
	xo "github.com/xo/xo/types"
)

// Set holds a set of templates and handles generating files for a target
// files.
type Set struct {
	symbols   map[string]map[string]reflect.Value
	initfunc  string
	tags      []string
	target    string
	templates map[string]*Template
	files     map[string]*EmittedTemplate
	post      map[string][]byte
	err       error
}

// NewTemplateSet creates a template set.
func NewTemplateSet(symbols map[string]map[string]reflect.Value, initfunc string, tags ...string) *Set {
	return &Set{
		symbols:   symbols,
		initfunc:  initfunc,
		tags:      tags,
		templates: make(map[string]*Template),
		files:     make(map[string]*EmittedTemplate),
		post:      make(map[string][]byte),
	}
}

// NewDefaultTemplateSet creates a template set using the default symbols, init
// func, tags, and embedded templates.
func NewDefaultTemplateSet(ctx context.Context) *Set {
	return NewTemplateSet(DefaultSymbols(), DefaultInitFunc, DefaultTags()...)
}

// LoadDefaults loads the default templates. Sets the default template target
// to "go" if available in embedded templates, or to the first available
// target.
func (ts *Set) LoadDefaults(ctx context.Context) error {
	if err := ts.AddTemplates(ctx, files, true); err != nil {
		return err
	}
	// determine default target
	switch targets := ts.Targets(); {
	case contains(targets, "go"):
		ts.Use("go")
	case len(targets) != 0:
		ts.Use(targets[0])
	}
	return nil
}

// LoadDefault loads a single default template.
func (ts *Set) LoadDefault(ctx context.Context, target string) error {
	dir, err := files.ReadDir(".")
	if err != nil {
		return err
	}
	for _, d := range dir {
		if d.Name() != target {
			continue
		}
		sub, err := fs.Sub(files, target)
		if err != nil {
			return err
		}
		ts.Add(ctx, target, sub, true)
		return nil
	}
	return fmt.Errorf("default template not found: %s", target)
}

// AddTemplates adds templates to the template set from src, adding a template
// for each subdirectory.
func (ts *Set) AddTemplates(ctx context.Context, src fs.FS, unrestricted bool) error {
	// get target dir names
	var targets []string
	if err := fs.WalkDir(src, ".", func(n string, d fs.DirEntry, err error) error {
		switch {
		case err != nil:
			return err
		case d.IsDir() && n != ".":
			targets = append(targets, n)
		}
		return nil
	}); err != nil {
		return err
	}
	sort.Strings(targets)
	// add templates
	for _, target := range targets {
		src, err := fs.Sub(src, target)
		if err != nil {
			return err
		}
		if _, err := ts.Add(ctx, target, src, unrestricted); err != nil {
			return err
		}
	}
	return nil
}

func (ts *Set) Clear(ctx context.Context) {
}

// Add adds a template target from src to the template set.
func (ts *Set) Add(ctx context.Context, target string, src fs.FS, unrestricted bool) (string, error) {
	// create template
	tpl, err := ts.NewTemplate(ctx, target, src, unrestricted)
	if err != nil {
		return "", err
	}
	// check target not already defined
	if ts.Has(tpl.Target) {
		return "", fmt.Errorf("cannot redefine template target %q", tpl.Target)
	}
	ts.templates[tpl.Target] = tpl
	return tpl.Target, nil
}

// Use sets the template target.
func (ts *Set) Use(target string) {
	ts.target = target
}

// Target returns the template target.
func (ts *Set) Target() string {
	return ts.target
}

// Has determines if a template target has been defined.
func (ts *Set) Has(target string) bool {
	_, ok := ts.templates[target]
	return ok
}

// Targets returns the available template targets.
func (ts *Set) Targets() []string {
	var targets []string
	for target := range ts.templates {
		targets = append(targets, target)
	}
	sort.Strings(targets)
	return targets
}

// Flags returns flag options for a template target.
func (ts *Set) Flags(target string) []xo.FlagSet {
	tpl, ok := ts.templates[target]
	if ok {
		return tpl.Flags()
	}
	return nil
}

// For determines if the the template target supports the mode.
func (ts *Set) For(mode string) error {
	if tpl, ok := ts.templates[ts.target]; ok && contains(tpl.Type.Modes, mode) {
		return nil
	}
	return fmt.Errorf("template %s does not support %s", ts.target, mode)
}

// Src returns template target file source.
func (ts *Set) Src() (fs.FS, error) {
	tpl, ok := ts.templates[ts.target]
	if !ok {
		return nil, fmt.Errorf("unknown target %q", ts.target)
	}
	return tpl.Src, nil
}

// NewContext creates a new context for the template target.
func (ts *Set) NewContext(ctx context.Context, mode string) context.Context {
	tpl, ok := ts.templates[ts.target]
	if !ok {
		ts.err = fmt.Errorf("unknown target %q", ts.target)
		return nil
	}
	if tpl.Type.NewContext != nil {
		return tpl.Type.NewContext(ctx, mode)
	}
	return ctx
}

// Pre performs pre processing of the template target.
func (ts *Set) Pre(ctx context.Context, mode string, src fs.FS) {
	tpl, ok := ts.templates[ts.target]
	if !ok {
		ts.err = fmt.Errorf("unknown target %q", ts.target)
		return
	}
	/*
		set.err = template.Type.Pre(ctx, src, func(file string, buf []byte) {
		})
	*/
	tpl = tpl
	if ts.err != nil {
		return
	}
}

// Process processes the template target.
func (ts *Set) Process(ctx context.Context, mode string, set *xo.Set) {
	tpl, ok := ts.templates[ts.target]
	if !ok {
		ts.err = fmt.Errorf("unknown target %q", ts.target)
		return
	}
	// capture emitted templates
	var emitted []*EmittedTemplate
	tpl, emitted = tpl, emitted
	/*
		set.err = template.Type.Process(ctx, v, func(tpl xo.Template) {
			emitted = append(emitted, &EmittedTemplate{
				Template: tpl,
			})
		})
	*/
	if ts.err != nil {
		return
	}
	/*
		// sort the emitted templates
		sort.Slice(emitted, func(i, j int) bool {
			if emitted[i].Template.Template != emitted[j].Template.Template {
				return emitted[i].Template.Template < emitted[j].Template.Template
			}
			if emitted[i].Template.Type != emitted[j].Template.Type {
				return emitted[i].Template.Type < emitted[j].Template.Type
			}
			return emitted[i].Template.Name < emitted[j].Template.Name
		})
		order := template.Type.Order
		// add package templates
		if !doAppend && template.Type.PackageTemplates != nil {
			var additional []string
			for _, tpl := range template.Type.PackageTemplates(ctx) {
				if err := emit(tpl); err != nil {
					return err
				}
				additional = append(additional, tpl.Type)
			}
			order = removeMatching(template.Type.Order, additional)
			order = append(additional, order...)
		}
		files := make(map[string]*EmittedTemplate)
		for _, n := range order {
			for _, tpl := range emitted {
				if tpl.Template.Type != n {
					continue
				}
				file, ok := files[tpl.Dest]
				if !ok {
					buf, err := template.LoadFile(ctx, tpl.File, doAppend)
					if err != nil {
						return err
					}
					file = &EmittedTemplate{
						Buf:  buf,
						File: tpl.File,
					}
					files[tpl.File] = file
				}
				file.Buf = append(file.Buf, tpl.Buf...)
			}
		}
	*/
}

// Post performs post processing of the template target.
func (ts *Set) Post(ctx context.Context, mode string) {
	tpl, ok := ts.templates[ts.target]
	switch {
	case !ok:
		ts.err = fmt.Errorf("unknown target %q", ts.target)
		return
	case tpl.Type.Post == nil:
		return
	}
	var files []string
	for file := range ts.files {
		files = append(files, file)
	}
	sort.Strings(files)
	for _, file := range files {
		/*
			err := template.Type.Post(ctx, mode, file)
			switch {
			case err != nil:
				set.files[file].Err = append(set.files[file].Err, &ErrPostFailed{file, err})
			case err == nil:
				set.files[file].Buf = buf
			}
		*/
		file = file
	}
}

// Dump dumps generated files to disk.
func (ts *Set) Dump(out string) {
	var files []string
	for file := range ts.files {
		files = append(files, file)
	}
	sort.Strings(files)
	for _, file := range files {
		if err := ioutil.WriteFile(filepath.Join(out, file), ts.files[file].Buf, 0644); err != nil {
			ts.files[file].Err = append(ts.files[file].Err, err)
		}
	}
}

// Errors returns any collected errors.
func (set *Set) Errors() []error {
	var files []string
	for file := range set.files {
		files = append(files, file)
	}
	sort.Strings(files)
	var errors []error
	if set.err != nil {
		errors = append(errors, set.err)
	}
	for _, file := range files {
		errors = append(errors, set.files[file].Err...)
	}
	return errors
}

// Template wraps a template.
type Template struct {
	Target string
	Type   xo.TemplateType
	Interp *interp.Interpreter
	Src    fs.FS
}

// NewTemplate creates a new template from the provided fs. Creates a
// github.com/traefik/yaegi interpreter and evaluates the template. See
// existing templates for implementation examples.
//
// Uses the template set's symbols, init func name, and declared tags.
func (ts *Set) NewTemplate(ctx context.Context, target string, src fs.FS, unrestricted bool) (*Template, error) {
	// build interpreter for custom funcs
	i := interp.New(interp.Options{
		GoPath:               ".",
		BuildTags:            ts.tags,
		SourcecodeFilesystem: sourceFS{path: "src/main/vendor/" + target, fs: src},
		Unrestricted:         unrestricted,
	})
	// add symbols
	if ts.symbols != nil {
		if err := i.Use(ts.symbols); err != nil {
			return nil, fmt.Errorf("%s: could not add xo internal symbols to yaegi: %w", target, err)
		}
	}
	// import
	if _, err := i.Eval(fmt.Sprintf("import (xotpl %q)", target)); err != nil {
		return nil, fmt.Errorf("%s: unable to import package: %w", target, err)
	}
	// eval init
	v, err := i.Eval("xotpl." + ts.initfunc)
	if err != nil {
		return nil, fmt.Errorf("%s: unable to eval %q: %w", target, ts.initfunc, err)
	}
	// convert init
	f, ok := v.Interface().(func(context.Context, func(xo.TemplateType)) error)
	if !ok {
		return nil, fmt.Errorf("%s: %s has signature `%T` (must be `func(context.Context, func(github.com/xo/types.TemplateType)) error`)", target, ts.initfunc, v.Interface())
	}
	// init
	var typ xo.TemplateType
	if err := f(ctx, func(tplType xo.TemplateType) {
		typ = tplType
	}); err != nil {
		return nil, fmt.Errorf("%s: %s error: %w", target, ts.initfunc, err)
	}
	if typ.Name != "" {
		target = typ.Name
	}
	return &Template{
		Target: target,
		Type:   typ,
		Interp: i,
		Src:    src,
	}, nil
}

// Flags returns the dynamic flags for the template.
func (tpl *Template) Flags() []xo.FlagSet {
	var flags []xo.FlagSet
	for _, flag := range tpl.Type.Flags {
		flags = append(flags, xo.FlagSet{
			Type: tpl.Target,
			Name: string(flag.ContextKey),
			Flag: flag,
		})
	}
	return flags
}

/*
// Emit emits a template to the template.
func (typ TemplateType) Emit(ctx context.Context, tpl Template) error {
	buf, err := typ.Exec(ctx, tpl)
	if err != nil {
		return err
	}
	typ.emitted = append(typ.emitted, &EmittedTemplate{Template: tpl, Buf: buf})
	return nil
}

// Exec loads and executes a template.
func (typ TemplateType) Exec(ctx context.Context, fs fs.FS, tpl Template) ([]byte, error) {
	return nil, fmt.Errorf("TemplateType.Exec")

		t, err := typ.Load(ctx, fs, tpl)
		if err != nil {
			return nil, err
		}
		buf := new(bytes.Buffer)
		if err := t.Execute(buf, tpl); err != nil {
			return nil, fmt.Errorf("unable to exec template %s: %w", tpl.File(), err)
		}
		return buf.Bytes(), nil
}

// Load loads a template.
func (typ TemplateType) Load(ctx context.Context, fs fs.FS, tpl Template) (*template.Template, error) {
	// template source
	// load template content
	name := tpl.File() + ".tpl"
	f, err := fs.Open(name)
	if err != nil {
		return nil, fmt.Errorf("unable to open template %s: %w", name, err)
	}
	defer f.Close()
	// read template
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("unable to read template %s: %w", name, err)
	}
	// create template and add funcs
	t := template.New(name)
	if typ.Funcs != nil {
		funcs, err := typ.Funcs(ctx)
		if err != nil {
			return nil, err
		}
		t = t.Funcs(funcs)
	}
	// parse content
	if t, err = t.Parse(string(buf)); err != nil {
		return nil, fmt.Errorf("unable to parse template %s: %w", name, err)
	}
	return t, nil
}

// LoadFile loads a file.
func (typ TemplateType) LoadFile(ctx context.Context, file string, doAppend bool) ([]byte, error) {
	return nil, errors.New("TemplateType.LoadFile")
		name := filepath.Join(Out(ctx), file)
		fi, err := os.Stat(name)
		switch {
		case (err != nil && errors.Is(err, os.ErrNotExist)) || !doAppend:
			if typ.HeaderTemplate == nil {
				return nil, nil
			}
			return typ.Exec(ctx, fs, typ.HeaderTemplate(ctx))
		case err != nil:
			return nil, err
		case fi.IsDir():
			return nil, fmt.Errorf("%s is a directory: cannot emit template", name)
		}
		return ioutil.ReadFile(name)
}
*/

// EmittedTemplate wraps a template with its content and file name.
type EmittedTemplate struct {
	Template xo.Template
	Buf      []byte
	Err      []error
}

// ErrPostFailed is the post failed error.
type ErrPostFailed struct {
	File string
	Err  error
}

// Error satisfies the error interface.
func (err *ErrPostFailed) Error() string {
	return fmt.Sprintf("post failed %s: %v", err.File, err.Err)
}

// Unwrap satisfies the unwrap interface.
func (err *ErrPostFailed) Unwrap() error {
	return err.Err
}

// DefaultSymbols returns the default set of yaegi and internal symbols.
func DefaultSymbols() map[string]map[string]reflect.Value {
	symbols := make(map[string]map[string]reflect.Value)
	for _, syms := range []map[string]map[string]reflect.Value{
		stdlib.Symbols,
		internal.Symbols,
	} {
		for kk, m := range syms {
			z := make(map[string]reflect.Value)
			for k, v := range m {
				z[k] = v
			}
			symbols[kk] = z
		}
	}
	return symbols
}

// DefaultInitFunc is the template init symbol.
const DefaultInitFunc = "Init"

// DefaultTags returns the default template tags.
func DefaultTags() []string {
	return []string{"xotpl"}
}

// removeMatching builds a new slice from v containing the strings not
// contained in s.
func removeMatching(v []string, s []string) []string {
	var res []string
	for _, z := range v {
		if contains(s, z) {
			continue
		}
		res = append(res, z)
	}
	return res
}

// contains returns true when s is in v.
func contains(v []string, s string) bool {
	for _, z := range v {
		if z == s {
			return true
		}
	}
	return false
}

// sourceFS handles source file mapping in a file system.
type sourceFS struct {
	fs   fs.FS
	path string
}

// Open satisfies the fs.FS interface.
func (src sourceFS) Open(name string) (fs.File, error) {
	if name == src.path {
		return src.fs.Open(".")
	}
	if n := src.path + "/"; strings.HasPrefix(name, n) {
		return src.fs.Open(strings.TrimPrefix(name, n))
	}
	return nil, os.ErrNotExist
}

// files are embedded template files.
//
//go:embed createdb
//go:embed dot
//go:embed go
//go:embed json
//go:embed yaml
var files embed.FS
