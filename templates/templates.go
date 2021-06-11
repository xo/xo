package templates

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"text/template"
)

// templates are registered template sets.
var templates = map[string]*TemplateSet{}

// Register registers a template set.
func Register(typ string, set *TemplateSet) {
	templates[typ] = set
}

// Types returns the registered type names of template sets.
func Types() []string {
	var types []string
	for k := range templates {
		types = append(types, k)
	}
	sort.Strings(types)
	return types
}

// Flags returns flag options and context for the template sets.
//
// These should be added to the invocation context for any call to a template
// set func.
func Flags() []FlagSet {
	var flags []FlagSet
	for _, typ := range Types() {
		set := templates[typ]
		for _, flag := range set.Flags {
			flags = append(flags, FlagSet{
				Type: typ,
				Name: string(flag.ContextKey),
				Flag: flag,
			})
		}
	}
	return flags
}

// FlagSet wraps a option flag with context information.
type FlagSet struct {
	Type string
	Name string
	Flag Flag
}

// Flag is a option flag.
type Flag struct {
	ContextKey  ContextKey
	Desc        string
	PlaceHolder string
	Default     string
	Short       rune
	Value       interface{}
	Enums       []string
}

// AddKnownType adds a known type name to a template set.
func AddKnownType(ctx context.Context, name string) error {
	typ := TemplateType(ctx)
	set, ok := templates[typ]
	if !ok {
		return fmt.Errorf("unknown template %q", typ)
	}
	if set.AddType != nil {
		set.AddType(name)
	}
	return nil
}

// Emit emits a template to a template set.
func Emit(ctx context.Context, tpl *Template) error {
	typ := TemplateType(ctx)
	set, ok := templates[typ]
	if !ok {
		return fmt.Errorf("unknown template %q", typ)
	}
	return set.Emit(ctx, tpl)
}

// Process processes emitted templates for a template set.
func Process(ctx context.Context, doAppend bool, single string, order ...string) error {
	typ := TemplateType(ctx)
	set, ok := templates[typ]
	if !ok {
		return fmt.Errorf("unknown template %q", typ)
	}
	sortEmitted(set.emitted)
	// add package templates
	if !doAppend {
		var additional []string
		for _, tpl := range set.PackageTemplates(ctx) {
			if err := set.Emit(ctx, tpl); err != nil {
				return err
			}
			additional = append(additional, tpl.Template)
		}
		order = removeMatching(order, additional)
		order = append(additional, order...)
	}
	set.files = make(map[string]*EmittedTemplate)
	for _, n := range order {
		for _, tpl := range set.emitted {
			if tpl.Template.Template != n {
				continue
			}
			fileExt := set.FileExt
			if s := Suffix(ctx); s != "" {
				fileExt = s
			}
			// determine filename
			if single != "" {
				tpl.File = single
			} else {
				tpl.File = set.FileName(ctx, tpl.Template) + fileExt
			}
			// load
			file, ok := set.files[tpl.File]
			if !ok {
				buf, err := set.LoadFile(ctx, tpl.File, doAppend)
				if err != nil {
					return err
				}
				file = &EmittedTemplate{
					Buf:  buf,
					File: tpl.File,
				}
				set.files[tpl.File] = file
			}
			file.Buf = append(file.Buf, tpl.Buf...)
		}
	}
	return nil
}

// Write performs post processing of emitted templates to a template set,
// writing to disk the processed content.
func Write(ctx context.Context) error {
	typ := TemplateType(ctx)
	set, ok := templates[typ]
	if !ok {
		return fmt.Errorf("unknown template %q", typ)
	}
	var files []string
	for file := range set.files {
		files = append(files, file)
	}
	sort.Strings(files)
	for _, file := range files {
		buf, err := set.Post(ctx, set.files[file].Buf)
		switch {
		case err != nil:
			set.files[file].Err = append(set.files[file].Err, &ErrPostFailed{file, err})
		case err == nil:
			set.files[file].Buf = buf
		}
	}
	return WriteFiles(ctx)
}

// WriteFiles writes the generated files to disk.
func WriteFiles(ctx context.Context) error {
	typ := TemplateType(ctx)
	set, ok := templates[typ]
	if !ok {
		return fmt.Errorf("unknown template %q", typ)
	}
	out := Out(ctx)
	var files []string
	for file := range set.files {
		files = append(files, file)
	}
	sort.Strings(files)
	for _, file := range files {
		if err := ioutil.WriteFile(filepath.Join(out, file), set.files[file].Buf, 0644); err != nil {
			set.files[file].Err = append(set.files[file].Err, err)
		}
	}
	return nil
}

// WriteRaw writes the raw templates for a template set.
func WriteRaw(ctx context.Context) error {
	typ := TemplateType(ctx)
	set, ok := templates[typ]
	if !ok {
		return fmt.Errorf("unknown template %q", typ)
	}
	out := Out(ctx)
	return fs.WalkDir(set.Files, ".", func(n string, d fs.DirEntry, err error) error {
		switch {
		case err != nil:
			return err
		case d.IsDir():
			return os.MkdirAll(filepath.Join(out, n), 0755)
		}
		buf, err := set.Files.ReadFile(n)
		if err != nil {
			return err
		}
		return ioutil.WriteFile(filepath.Join(out, n), buf, 0644)
	})
}

// Errors returns errors collected during file generation.
func Errors(ctx context.Context) ([]error, error) {
	typ := TemplateType(ctx)
	set, ok := templates[typ]
	if !ok {
		return nil, fmt.Errorf("unknown template %q", typ)
	}
	var files []string
	for file := range set.files {
		files = append(files, file)
	}
	sort.Strings(files)
	var errors []error
	for _, file := range files {
		errors = append(errors, set.files[file].Err...)
	}
	return errors, nil
}

// TemplateSet is a template set.
type TemplateSet struct {
	// Files are the embedded templates.
	Files embed.FS
	// FileExt is the file extension added to out files.
	FileExt string
	// AddType will be called when a new type is encountered.
	AddType func(string)
	// Flags are additional template flags.
	Flags []Flag
	// HeaderTemplate is the name of the header template.
	HeaderTemplate func(context.Context) *Template
	// PackageTemplates returns package templates.
	PackageTemplates func(context.Context) []*Template
	// Funcs provides template funcs for use by templates.
	Funcs func(context.Context) (template.FuncMap, error)
	// FileName determines the out file name templates.
	FileName func(context.Context, *Template) string
	// Post performs post processing of generated content.
	Post func(context.Context, []byte) ([]byte, error)
	// emitted holds emitted templates.
	emitted []*EmittedTemplate
	// files holds the generated files.
	files map[string]*EmittedTemplate
}

// Emit emits a template to the template set.
func (set *TemplateSet) Emit(ctx context.Context, tpl *Template) error {
	buf, err := set.Exec(ctx, tpl)
	if err != nil {
		return err
	}
	set.emitted = append(set.emitted, &EmittedTemplate{Template: tpl, Buf: buf})
	return nil
}

// Exec loads and executes a template.
func (set *TemplateSet) Exec(ctx context.Context, tpl *Template) ([]byte, error) {
	t, err := set.Load(ctx, tpl)
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
func (set *TemplateSet) Load(ctx context.Context, tpl *Template) (*template.Template, error) {
	// template source
	src := Src(ctx)
	if src == nil {
		src = set.Files
	}
	// load template content
	name := tpl.File() + set.FileExt + ".tpl"
	f, err := src.Open(name)
	if err != nil {
		return nil, fmt.Errorf("unable to open template %s: %w", name, err)
	}
	defer f.Close()
	// read template
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("unable to read template %s: %w", name, err)
	}
	// build funcs
	funcs, err := set.Funcs(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to build funcs %s: %v", name, err)
	}
	// parse content
	t, err := template.New(name).Funcs(funcs).Parse(string(buf))
	if err != nil {
		return nil, fmt.Errorf("unable to parse template %s: %w", name, err)
	}
	return t, nil
}

// LoadFile loads a file.
func (set *TemplateSet) LoadFile(ctx context.Context, file string, doAppend bool) ([]byte, error) {
	name := filepath.Join(Out(ctx), file)
	fi, err := os.Stat(name)
	switch {
	case (err != nil && os.IsNotExist(err)) || !doAppend:
		return set.Exec(ctx, set.HeaderTemplate(ctx))
	case err != nil:
		return nil, err
	case fi.IsDir():
		return nil, fmt.Errorf("%s is a directory: cannot emit template", name)
	}
	return ioutil.ReadFile(name)
}

// EmittedTemplate wraps a template with its content and file name.
type EmittedTemplate struct {
	Template *Template
	Buf      []byte
	File     string
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

// ContextKey is a context key.
type ContextKey string

// Context keys.
const (
	GenTypeKey      ContextKey = "gen-type"
	TemplateTypeKey ContextKey = "template-type"
	SuffixKey       ContextKey = "suffix"
	DriverKey       ContextKey = "driver"
	SchemaKey       ContextKey = "schema"
	SrcKey          ContextKey = "src"
	NthParamKey     ContextKey = "nth-param"
	OutKey          ContextKey = "out"
)

// GenType returns the the gen-type option from the context.
func GenType(ctx context.Context) string {
	s, _ := ctx.Value(GenTypeKey).(string)
	return s
}

// TemplateType returns type option from the context.
func TemplateType(ctx context.Context) string {
	s, _ := ctx.Value(TemplateTypeKey).(string)
	return s
}

// Suffix returns suffix option from the context.
func Suffix(ctx context.Context) string {
	s, _ := ctx.Value(SuffixKey).(string)
	return s
}

// Driver returns driver option from the context.
func Driver(ctx context.Context) string {
	s, _ := ctx.Value(DriverKey).(string)
	return s
}

// Schema returns schema option from the context.
func Schema(ctx context.Context) string {
	s, _ := ctx.Value(SchemaKey).(string)
	return s
}

// Src returns src option from the context.
func Src(ctx context.Context) fs.FS {
	v, _ := ctx.Value(SrcKey).(fs.FS)
	return v
}

// NthParam returns the nth-param option from the context.
func NthParam(ctx context.Context) func(int) string {
	f, _ := ctx.Value(NthParamKey).(func(int) string)
	return f
}

// Out returns out option from the context.
func Out(ctx context.Context) string {
	s, _ := ctx.Value(OutKey).(string)
	return s
}

// sortEmitted sorts the emitted templates.
func sortEmitted(tpl []*EmittedTemplate) {
	sort.Slice(tpl, func(i, j int) bool {
		if tpl[i].Template.Template != tpl[j].Template.Template {
			return tpl[i].Template.Template < tpl[j].Template.Template
		}
		if tpl[i].Template.Type != tpl[j].Template.Type {
			return tpl[i].Template.Type < tpl[j].Template.Type
		}
		return tpl[i].Template.Name < tpl[j].Template.Name
	})
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
