package jsontpl

import (
	"context"
	"encoding/json"
	"fmt"
	"text/template"
)

// Funcs is a set of template funcs.
type Funcs struct {
	Indent string
}

// NewFuncs creates a new Funcs
func NewFuncs(ctx context.Context) *Funcs {
	return &Funcs{
		Indent: Indent(ctx),
	}
}

// FuncMap returns the func map.
func (f *Funcs) FuncMap() template.FuncMap {
	return template.FuncMap{
		"json": f.jsonfn,
	}
}

func (f *Funcs) jsonfn(v interface{}) string {
	buf, err := json.MarshalIndent(v, "", f.Indent)
	if err != nil {
		return fmt.Sprintf("[[ COULD NOT MARSHAL: %v ]]", err)
	}
	return string(buf)
}
