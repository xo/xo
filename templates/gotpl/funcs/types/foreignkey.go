// Package types provides schema types.
package types

// ForeignKey is a foreign key template.
type ForeignKey struct {
	GoName      string
	SQLName     string
	Table       Table
	Fields      []Field
	RefTable    string
	RefFields   []Field
	RefFuncName string
	Comment     string
}
