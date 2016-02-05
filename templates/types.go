package templates

import "github.com/knq/xo/models"

// EnumTemplate is a template item for a enum.
type EnumTemplate struct {
	Type     string
	EnumType string
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
	Parameters         []*models.Column
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
	Fields      []*models.Column
	Table       *TableTemplate
}
