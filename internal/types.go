package internal

import "github.com/knq/xo/models"

// TemplateType represents a template type
type TemplateType uint

// the order here will be the alter the output order per file.
const (
	XO TemplateType = iota
	Enum
	Model
	Proc
	Index
	ForeignKey
	QueryModel
	Query
)

// String returns the name for the associated template type.
func (tt TemplateType) String() string {
	var s string
	switch tt {
	case XO:
		s = "xo_db"
	case Enum:
		s = "enum"
	case Proc:
		s = "proc"
	case Model:
		s = "model"
	case Index:
		s = "idx"
	case ForeignKey:
		s = "fkey"
	case QueryModel:
		s = "model"
	case Query:
		s = "query"

	default:
		panic("unknown TemplateType")
	}
	return s
}

// EnumTemplate is a template item for a enum.
type EnumTemplate struct {
	Type     string
	EnumType string
	Comment  string
	Values   []*models.Enum
}

// TableTemplate is a template item for a table.
type TableTemplate struct {
	Type            string
	TableSchema     string
	TableName       string
	PrimaryKey      string
	PrimaryKeyField string
	PrimaryKeyType  string
	Comment         string
	Fields          []*models.Column
}

// ProcTemplate is a template item for a stored procedure.
type ProcTemplate struct {
	Name               string
	ReturnType         string
	NilReturnType      string
	TableSchema        string
	ProcName           string
	ProcParameterNames string
	ProcParameterTypes string
	ProcReturnType     string
	Comment            string
	Parameters         []*models.Column
}

// ForeignKeyTemplate is a template item for a foreign relationship on a table.
type ForeignKeyTemplate struct {
	Type           string
	ForeignKeyName string
	ColumnName     string
	Field          string
	FieldType      string
	RefType        string
	RefField       string
	RefFieldType   string
}

// IndexTemplate is a template item for a index into a table.
type IndexTemplate struct {
	Type        string
	Name        string
	TableSchema string
	TableName   string
	IndexName   string
	IsUnique    bool
	Plural      string
	Comment     string
	Fields      []*models.Column
	Table       *TableTemplate
}

// FuncTemplate is a template item for a custom query.
type QueryTemplate struct {
	Name          string
	Type          string
	Query         []string
	QueryComments []string
	Parameters    []QueryParameter
	OnlyOne       bool
	Interpolate   bool
	Comment       string
	Table         *TableTemplate
}
