package templates

import (
	"github.com/xo/xo/models"
)

// Template wraps other templates.
type Template struct {
	Set      string
	Template string
	Type     string
	Name     string
	Data     interface{}
	Extra    map[string]interface{}
}

// File returns the file name for the template.
func (tpl *Template) File() string {
	if tpl.Set != "" {
		return tpl.Set + "/" + tpl.Template
	}
	return tpl.Template
}

// EnumValue is a enum value template.
type EnumValue struct {
	Name    string
	Val     *models.EnumValue
	Comment string
}

// Enum is a enum type template.
type Enum struct {
	Name    string
	Values  []*EnumValue
	Enum    *models.Enum
	Comment string
}

// Proc is a stored procedure template.
type Proc struct {
	Name       string
	ProcParams string
	Params     []*Field
	Return     *Field
	Proc       *models.Proc
	Comment    string
}

// Field is a field template.
type Field struct {
	Name    string
	Type    string
	Zero    string
	Prec    int
	Col     *models.Column
	Comment string
}

// Type is a type (ie, table/view/custom query) template.
type Type struct {
	Name             string
	Kind             string
	PrimaryKey       *Field
	PrimaryKeyFields []*Field
	Fields           []*Field
	Table            *models.Table
	Comment          string
}

// ForeignKey is a foreign key template.
type ForeignKey struct {
	Name       string
	Type       *Type
	Field      *Field
	RefType    *Type
	RefField   *Field
	ForeignKey *models.ForeignKey
	Comment    string
}

// Index is an index template.
type Index struct {
	FuncName string
	Type     *Type
	Fields   []*Field
	Index    *models.Index
	Comment  string
}

// QueryParam is a custom query parameter template.
type QueryParam struct {
	Name        string
	Type        string
	Interpolate bool
	Join        bool
}

// Query is a custom query template.
type Query struct {
	Name        string
	Query       []string
	Comments    []string
	Params      []*QueryParam
	One         bool
	Flat        bool
	Exec        bool
	Interpolate bool
	Type        *Type
	Comment     string
}
