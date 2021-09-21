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
		// yaml marshals v as yaml.
		"yaml": func(v interface{}) (string, error) {
			buf, err := yaml.MarshalWithOptions(v)
			if err != nil {
				return "", err
			}
			return string(buf), nil
		},
	}, nil
}
