package templates

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"text/template"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
	"github.com/xo/xo/internal"
	xo "github.com/xo/xo/types"
)

// Set holds a set of templates and handles generating files for a target
// files.
//
// Set should not be used more than once to output its contents.
type Set struct {
	symbols  map[string]map[string]reflect.Value
	initfunc string
	tags     []string
	target   string
	targets  map[string]*Target
	files    map[string]*EmittedTemplate
	post     map[string][]byte
	err      error
	goTpl    *template.Template
}

// NewTemplateSet creates a template set.
func NewTemplateSet(symbols map[string]map[string]reflect.Value, initfunc string, tags ...string) *Set {
	return &Set{
		symbols:  symbols,
		initfunc: initfunc,
		tags:     tags,
		targets:  make(map[string]*Target),
		files:    make(map[string]*EmittedTemplate),
		post:     make(map[string][]byte),
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

// LoadDefault loads a single default target.
func (ts *Set) LoadDefault(ctx context.Context, name string) error {
	dir, err := files.ReadDir(".")
	if err != nil {
		return err
	}
	for _, d := range dir {
		if d.Name() != name {
			continue
		}
		sub, err := fs.Sub(files, name)
		if err != nil {
			return err
		}
		_, err = ts.Add(ctx, name, sub, true)
		return err
	}
	return fmt.Errorf("unknown target %q", ts.target)
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

// Add adds a target from src to the template set.
func (ts *Set) Add(ctx context.Context, name string, src fs.FS, unrestricted bool) (string, error) {
	// create template
	target, err := ts.NewTemplate(ctx, name, src, unrestricted)
	if err != nil {
		return "", err
	}
	// check target not already defined
	if ts.Has(target.Name) {
		return "", fmt.Errorf("cannot redefine template target %q", target.Name)
	}
	ts.targets[target.Name] = target
	return target.Name, nil
}

// Use sets the target being used.
func (ts *Set) Use(name string) {
	ts.target = name
}

// Target returns the target being used.
func (ts *Set) Target() string {
	return ts.target
}

// Has determines if a template target has been defined.
func (ts *Set) Has(name string) bool {
	_, ok := ts.targets[name]
	return ok
}

// Targets returns the available template targets.
func (ts *Set) Targets() []string {
	var targets []string
	for target := range ts.targets {
		targets = append(targets, target)
	}
	sort.Strings(targets)
	return targets
}

// Flags returns flag options for a template target.
func (ts *Set) Flags(name string) []xo.FlagSet {
	target, ok := ts.targets[name]
	if ok {
		return target.Flags()
	}
	return nil
}

// For determines if the the template target supports the mode.
func (ts *Set) For(mode string) error {
	if target, ok := ts.targets[ts.target]; ok && contains(target.Type.Modes, mode) {
		return nil
	}
	return fmt.Errorf("template %s does not support %s", ts.target, mode)
}

// Src returns template target file source.
func (ts *Set) Src() (fs.FS, error) {
	target, ok := ts.targets[ts.target]
	if !ok {
		return nil, fmt.Errorf("unknown target %q", ts.target)
	}
	return target.Src, nil
}

// NewContext creates a new context for the template target.
func (ts *Set) NewContext(ctx context.Context, mode string) context.Context {
	target, ok := ts.targets[ts.target]
	if !ok {
		ts.err = fmt.Errorf("unknown target %q", ts.target)
		return nil
	}
	if target.Type.NewContext != nil {
		return target.Type.NewContext(ctx, mode)
	}
	return ctx
}

// addFile returns a function that handles adding templates.
func (ts *Set) addFile(ctx context.Context) func(xo.Template) {
	return func(t xo.Template) {
		singleFile := xo.Single(ctx)
		if singleFile != "" {
			// Force all templates to be outputted in the specified file if xo is in single mode.
			t.Dest = singleFile
		}
		if _, ok := ts.files[t.Dest]; !ok {
			ts.files[t.Dest] = &EmittedTemplate{}
		}
		ts.files[t.Dest].Template = append(ts.files[t.Dest].Template, t)
	}
}

// Pre performs pre processing of the template target.
func (ts *Set) Pre(ctx context.Context, outDir string, mode string, set *xo.Set) {
	target, ok := ts.targets[ts.target]
	switch {
	case !ok:
		ts.err = fmt.Errorf("unknown target %q", ts.target)
		return
	case target.Type.Pre == nil:
		return
	}
	if target.Type.Pre == nil {
	}
	out := os.DirFS(outDir)
	ts.err = target.Type.Pre(ctx, mode, set, out, ts.addFile(ctx))
	if ts.err != nil {
		return
	}
}

// Process processes the template target.
func (ts *Set) Process(ctx context.Context, outDir string, mode string, set *xo.Set) {
	target, ok := ts.targets[ts.target]
	switch {
	case !ok:
		ts.err = fmt.Errorf("unknown target %q", ts.target)
		return
	case target.Type.Process == nil:
		return
	}
	ts.err = target.Type.Process(ctx, mode, set, ts.addFile(ctx))
	if ts.err != nil {
		return
	}
	// Determine template order.
	order := make(map[string]int, 0)
	if target.Type.Order != nil {
		for i, o := range target.Type.Order(ctx, mode) {
			order[o] = i
		}
	}
	fs, err := ts.Src()
	if err != nil {
		ts.err = err
		return
	}
	// Parse templates and provide functions if applicable.
	ts.goTpl = template.New("")
	var funcs template.FuncMap
	if target.Type.Funcs != nil {
		var err error
		funcs, err = target.Type.Funcs(ctx, mode)
		if err != nil {
			ts.err = err
			return
		}
		ts.goTpl = ts.goTpl.Funcs(funcs)
	}
	ts.goTpl, err = ts.goTpl.ParseFS(fs, "*.tpl")
	if err != nil {
		ts.err = err
		return
	}
	// sort file output order
	filenames := make([]string, 0, len(ts.files))
	for k := range ts.files {
		filenames = append(filenames, k)
	}
	sort.Slice(filenames, func(i, j int) bool {
		return filenames[i] < filenames[j]
	})
	// Generate all files with the constructed template.
	for _, file := range filenames {
		emitted := ts.files[file]
		sort.Slice(emitted.Template, func(i int, j int) bool {
			if emitted.Template[i].Partial != emitted.Template[j].Partial {
				return order[emitted.Template[i].Partial] < order[emitted.Template[j].Partial]
			}
			if emitted.Template[i].SortType != emitted.Template[j].SortType {
				return emitted.Template[i].SortType < emitted.Template[j].SortType
			}
			return emitted.Template[i].SortName < emitted.Template[j].SortName
		})
		for _, tpl := range emitted.Template {
			if tpl.Src == "" {
				err := ts.goTpl.ExecuteTemplate(&emitted.Buf, tpl.Partial, tpl)
				if err != nil {
					ts.files[file].Err = append(ts.files[file].Err, err)
				}
				continue
			}
			gotpl, err := template.New("").Parse(tpl.Src)
			if err != nil {
				ts.files[file].Err = append(ts.files[file].Err, err)
				continue
			}
			err = gotpl.Execute(&emitted.Buf, tpl)
			if err != nil {
				ts.files[file].Err = append(ts.files[file].Err, err)
				continue
			}
		}
	}
}

// Post performs post processing of the template target.
func (ts *Set) Post(ctx context.Context, mode string) {
	target, ok := ts.targets[ts.target]
	switch {
	case !ok:
		ts.err = fmt.Errorf("unknown target %q", ts.target)
		return
	case target.Type.Post == nil:
		return
	}
	files := make(map[string][]byte, len(ts.files))
	for fileName, emitted := range ts.files {
		files[fileName] = emitted.Buf.Bytes()
	}
	err := target.Type.Post(ctx, mode, files, func(fileName string, content []byte) {
		// Reset the buffer and fill it with the provided content.
		ts.files[fileName].Buf.Reset()
		ts.files[fileName].Buf.Write(content)
	})
	if err != nil {
		ts.err = err
		return
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
		buf := ts.files[file].Buf.Bytes()
		if err := os.WriteFile(filepath.Join(out, file), buf, 0o644); err != nil {
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

// Target is set of files defining a template.
type Target struct {
	Name   string
	Type   xo.TemplateType
	Interp *interp.Interpreter
	Src    fs.FS
}

// NewTemplate creates a new template from the provided fs. Creates a
// github.com/traefik/yaegi interpreter and evaluates the template. See
// existing templates for implementation examples.
//
// Uses the template set's symbols, init func name, and declared tags.
func (ts *Set) NewTemplate(ctx context.Context, target string, src fs.FS, unrestricted bool) (*Target, error) {
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
	tplInit, ok := v.Interface().(func(context.Context, func(xo.TemplateType)) error)
	if !ok {
		return nil, fmt.Errorf("%s: %s has signature `%T` (must be `func(context.Context, func(github.com/xo/types.TemplateType)) error`)", target, ts.initfunc, v.Interface())
	}
	// init
	var typ xo.TemplateType
	err = tplInit(ctx, func(tplType xo.TemplateType) {
		typ = tplType
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %s error: %w", target, ts.initfunc, err)
	}
	if typ.Name != "" {
		target = typ.Name
	}
	return &Target{
		Name:   target,
		Type:   typ,
		Interp: i,
		Src:    src,
	}, nil
}

// Flags returns the dynamic flags for the template.
func (target *Target) Flags() []xo.FlagSet {
	var flags []xo.FlagSet
	for _, flag := range target.Type.Flags {
		flags = append(flags, xo.FlagSet{
			Type: target.Name,
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
	buf, err := io.ReadAll(f)
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
		return os.ReadFile(name)
}
*/

// EmittedTemplate wraps a template with its content and file name.
type EmittedTemplate struct {
	Template []xo.Template
	Buf      bytes.Buffer
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
	// Ensure that Windows paths have forward slash
	name = filepath.ToSlash(name)
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
