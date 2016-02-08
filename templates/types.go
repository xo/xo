package templates

import "github.com/knq/xo/models"

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

// FkTemplate is a template item for a foreign relationship on a table.
type FkTemplate struct {
	Type       string
	ColumnName string
	Field      string
	RefType    string
	RefField   string
}

// IdxTemplate is a template item for a index into a table.
type IdxTemplate struct {
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
type FuncTemplate struct {
	Name          string
	Type          string
	Query         []string
	QueryComments []string
	Parameters    [][]string
	OnlyOne       bool
	Comment       string
	Table         *TableTemplate
}
