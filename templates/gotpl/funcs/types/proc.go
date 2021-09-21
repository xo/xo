// Package types provides schema types.
package types

// Proc is a stored procedure template.
type Proc struct {
	Type           string
	GoName         string
	OverloadedName string
	SQLName        string
	Signature      string
	Params         []Field
	Returns        []Field
	Void           bool
	Overloaded     bool
	Comment        string
}

func (p Proc) GetName() string {
	n := p.GoName
	if p.Overloaded {
		n = p.OverloadedName
	}
	return n
}
