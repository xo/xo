package jsontpl

import (
	"context"
	"text/template"

	"github.com/goccy/go-yaml"
)

// Funcs is a set of template funcs.
type Funcs struct{}

// NewFuncs creates a new Funcs.
func NewFuncs(ctx context.Context) *Funcs {
	return &Funcs{}
}

// FuncMap returns the func map.
func (f *Funcs) FuncMap() template.FuncMap {
	return template.FuncMap{
		"yaml": f.yaml,
	}
}

// yaml marshals v as yaml.
func (f *Funcs) yaml(v interface{}) (string, error) {
	buf, err := yaml.MarshalWithOptions(v, yaml.Indent(2), yaml.IndentSequence(true))
	if err != nil {
		return "", err
	}
	return string(buf), nil
}
