package internal

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"text/template"

	templates "github.com/knq/xo/tplbin"
)

// TemplateLoader loads templates from the specified name.
func (a *ArgType) TemplateLoader(name string) ([]byte, error) {
	// no template path specified
	if a.TemplatePath == "" {
		return templates.Asset(name)
	}

	return ioutil.ReadFile(path.Join(a.TemplatePath, name))
}

// TemplateSet retrieves the created template set.
func (a *ArgType) TemplateSet() *TemplateSet {
	if a.templateSet == nil {
		a.templateSet = &TemplateSet{
			funcs: a.NewTemplateFuncs(),
			l:     a.TemplateLoader,
			tpls:  map[string]*template.Template{},
		}
	}

	return a.templateSet
}

// ExecuteTemplate loads and parses the supplied template with name and
// executes it with obj as the context.
func (a *ArgType) ExecuteTemplate(tt TemplateType, name string, sub string, obj interface{}) error {
	var err error

	// setup generated
	if a.Generated == nil {
		a.Generated = []TBuf{}
	}

	// create store
	v := TBuf{
		TemplateType: tt,
		Name:         name,
		Subname:      sub,
		Buf:          new(bytes.Buffer),
	}

	// build template name
	loaderType := ""
	if tt != XOTemplate {
		if a.LoaderType == "oci8" || a.LoaderType == "ora" {
			// force oracle for oci8 since the oracle driver doesn't recognize
			// 'oracle' as valid protocol
			loaderType = "oracle."
		} else {
			loaderType = a.LoaderType + "."
		}
	}
	templateName := fmt.Sprintf("%s%s.go.tpl", loaderType, tt)

	// execute template
	err = a.TemplateSet().Execute(v.Buf, templateName, obj)
	if err != nil {
		return err
	}

	a.Generated = append(a.Generated, v)
	return nil
}

// TemplateSet is a set of templates.
type TemplateSet struct {
	funcs template.FuncMap
	l     func(string) ([]byte, error)
	tpls  map[string]*template.Template
}

// Execute executes a specified template in the template set using the supplied
// obj as its parameters and writing the output to w.
func (ts *TemplateSet) Execute(w io.Writer, name string, obj interface{}) error {
	tpl, ok := ts.tpls[name]
	if !ok {
		// attempt to load and parse the template
		buf, err := ts.l(name)
		if err != nil {
			return err
		}

		// parse template
		tpl, err = template.New(name).Funcs(ts.funcs).Parse(string(buf))
		if err != nil {
			return err
		}
	}

	return tpl.Execute(w, obj)
}
