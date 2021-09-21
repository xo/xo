// Package types provides schema types.
package types

// EnumValue is a enum value template.
type EnumValue struct {
	GoName     string
	SQLName    string
	ConstValue int
}

// Enum is a enum type template.
type Enum struct {
	GoName  string
	SQLName string
	Values  []EnumValue
	Comment string
}

func (e Enum) GetName() string {
	return e.GoName
}
