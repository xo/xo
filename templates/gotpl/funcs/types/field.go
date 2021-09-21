// Package types provides schema types.
package types

// Field is a field template.
type Field struct {
	GoName     string
	SQLName    string
	Type       string
	Zero       string
	IsPrimary  bool
	IsSequence bool
	Comment    string
}

func (f Field) GetName() string {
	return f.GoName
}
