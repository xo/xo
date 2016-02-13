// Package templates contains the various Go code templates used by xo.
package internal

import (
	"bytes"
	"errors"
	"io"
	"text/template"
)

// ExecuteTemplate loads and parses the supplied template with name and
// executes it with obj as the context.
func (a *ArgType) ExecuteTemplate(tt TemplateType, name string, obj interface{}) error {
	var err error

	// setup generated
	if a.Generated == nil {
		a.Generated = []TBuf{}
	}

	// create store
	v := TBuf{
		Type: tt,
		Name: name,
		Buf:  new(bytes.Buffer),
	}

	// build template name
	templateName := ""
	if tt != XO {
		// grab tl
		tl, ok := a.Loader.(TypeLoader)
		if !ok {
			return errors.New("internal error")
		}

		templateName = tl.Schemes[0] + "."
	}
	templateName = templateName + tt.String() + ".go.tpl"

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
