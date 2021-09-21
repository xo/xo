// Package types provides schema types.
package types

// Table is a type (ie, table/view/custom query) template.
type Table struct {
	Type        string
	GoName      string
	SQLName     string
	PrimaryKeys []Field
	Fields      []Field
	Manual      bool
	Comment     string
}

func (t Table) GetName() string {
	return t.GoName
}
