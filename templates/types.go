package templates

import "github.com/knq/xo/models"

// EnumTemplate is a template item for a enum.
type EnumTemplate struct {
	Type       string
	TypeNative string
	Values     []*models.Enum
}

// TableTemplate is a template item for a table.
type TableTemplate struct {
	Type            string
	TableSchema     string
	TableName       string
	PrimaryKey      string
	PrimaryKeyField string
	Fields          []*models.Column
}
