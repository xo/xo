package jsontpl

import (
	"context"
	"encoding/json"
	"text/template"
)

// Funcs is a set of template funcs.
type Funcs struct {
	indent string
	ugly   bool
}

// NewFuncs creates a new Funcs.
func NewFuncs(ctx context.Context) *Funcs {
	return &Funcs{
		indent: Indent(ctx),
		ugly:   Ugly(ctx),
	}
}

// FuncMap returns the func map.
func (f *Funcs) FuncMap() template.FuncMap {
	return template.FuncMap{
		"json": f.jsonfn,
	}
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
