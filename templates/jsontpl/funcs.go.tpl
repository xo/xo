// Package funcs provides custom template funcs.
package funcs

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"text/template"

	"github.com/xo/xo/templates/jsontpl"
)

// Init intializes the custom template funcs.
func Init(ctx context.Context) (template.FuncMap, error) {
	return template.FuncMap{
		// json marshals v as json.
		"json": func(v interface{}) (string, error) {
			buf := new(bytes.Buffer)
			enc := json.NewEncoder(buf)
			if !jsontpl.Ugly(ctx) {
				enc.SetIndent("", jsontpl.Indent(ctx))
			}
			if err := enc.Encode(v); err != nil {
				return "", err
			}
			return strings.TrimSpace(buf.String()), nil
		},
	}, nil
}
