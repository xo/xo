// Package types provides schema types.
package types

// Index is an index template.
type Index struct {
	SQLName   string
	FuncName  string
	Table     Table
	Fields    []Field
	IsUnique  bool
	IsPrimary bool
	Comment   string
}

func (i Index) GetName() string {
	return i.FuncName
}
