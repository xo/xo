// Package funcs provides custom template funcs.
package funcs

import (
	"context"
	"encoding/json"
	"text/template"

	"github.com/xo/xo/templates/jsontpl"
)

// Init intializes the custom template funcs.
func Init(ctx context.Context) (template.FuncMap, error) {
	funcs := &Funcs{
		indent: jsontpl.Indent(ctx),
		ugly:   jsontpl.Ugly(ctx),
	}
	return template.FuncMap{
		"json": funcs.jsonfn,
	}, nil
}

// Funcs is a set of template funcs.
type Funcs struct {
	indent string
	ugly   bool
}

// jsonfn marshals v as json.
func (f *Funcs) jsonfn(v interface{}) (string, error) {
	z := json.MarshalIndent
	if f.ugly {
		z = uglyMarshal
	}
	// marshal
	buf, err := z(v, "", f.indent)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

// uglyMarshal marshals v without indentation.
func uglyMarshal(v interface{}, _, _ string) ([]byte, error) {
	return json.Marshal(v)
}
