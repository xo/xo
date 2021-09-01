// Package funcs provides custom template funcs.
package funcs

import (
	"context"
	"text/template"

	"github.com/goccy/go-yaml"
)

// Init intializes the custom template funcs.
func Init(ctx context.Context) (template.FuncMap, error) {
	return template.FuncMap{
		"yaml": yamlfn,
	}, nil
}

// yamlfn marshals v as yaml.
func yamlfn(v interface{}) (string, error) {
	buf, err := yaml.MarshalWithOptions(v, yaml.Indent(2), yaml.IndentSequence(true))
	if err != nil {
		return "", err
	}
	return string(buf), nil
}
