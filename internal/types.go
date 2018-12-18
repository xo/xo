package internal

import "github.com/JLightning/xo/models"

// TemplateType represents a template type.
type TemplateType uint

// the order here will be the alter the output order per file.
const (
	EnumTemplate TemplateType = iota
	ProcTemplate
	TypeTemplate
	RepositoryTemplate
	ForeignKeyTemplate
	IndexTemplate
	QueryTypeTemplate
	QueryTemplate

	// always last
	XOTemplate
	PaginationTemplate
	PaginationSchemaTemplate
	ScalarTemplate
	SchemaGraphQLTemplate
	SchemaGraphQLEnumTemplate
	SchemaGraphQLScalarTemplate
	GqlgenModelTemplate
	WireTemplate
)

// String returns the name for the associated template type.
func (tt TemplateType) String() string {
	var s string
	switch tt {
	case XOTemplate:
		s = "xo_db"
	case EnumTemplate:
		s = "enum"
	case ProcTemplate:
		s = "proc"
	case TypeTemplate:
		s = "type"
	case RepositoryTemplate:
		s = "repository"
	case ForeignKeyTemplate:
		s = "foreignkey"
	case IndexTemplate:
		s = "index"
	case QueryTypeTemplate:
		s = "querytype"
	case QueryTemplate:
		s = "query"
	case PaginationTemplate:
		s = "pagination"
	case PaginationSchemaTemplate:
		s = "pagination.schema"
	case ScalarTemplate:
		s = "scalar"
	case SchemaGraphQLTemplate:
		s = "schema.graphql"
	case SchemaGraphQLEnumTemplate:
		s = "schema.graphql.enum"
	case SchemaGraphQLScalarTemplate:
		s = "schema.graphql.scalar"
	case GqlgenModelTemplate:
		s = "gqlgen.yml.model"
	case WireTemplate:
		s = "wire"
	default:
		panic("unknown TemplateType")
	}
	return s
}

// RelType represents the different types of relational storage (table/view).
type RelType uint

const (
	// Table reltype
	Table RelType = iota

	// View reltype
	View
)

// EscType represents the different escape types.
type EscType uint

const (
	SchemaEsc = iota
	TableEsc
	ColumnEsc
)

// String provides the string representation of RelType.
func (rt RelType) String() string {
	var s string
	switch rt {
	case Table:
		s = "TABLE"
	case View:
		s = "VIEW"
	default:
		panic("unknown RelType")
	}
	return s
}

// EnumValue holds data for a single enum value.
type EnumValue struct {
	Name    string
	Val     *models.EnumValue
	Comment string
}

// Enum is a template item for a enum.
type Enum struct {
	Name              string
	Schema            string
	Values            []*EnumValue
	Enum              *models.Enum
	Comment           string
	ReverseConstNames bool
}

// Proc is a template item for a stored procedure.
type Proc struct {
	Name       string
	Schema     string
	ProcParams string
	Params     []*Field
	Return     *Field
	Proc       *models.Proc
	Comment    string
}

// Field contains field information.
type Field struct {
	Name    string
	Type    string
	NilType string
	Len     int
	Col     *models.Column
	Comment string
}

// Type is a template item for a type (ie, table/view/custom query).
type Type struct {
	Name             string
	RepoName         string
	Schema           string
	RelType          RelType
	PrimaryKey       *Field
	PrimaryKeyFields []*Field
	Fields           []*Field
	Table            *models.Table
	Comment          string
	Indexes          []*Index
}

// ForeignKey is a template item for a foreign relationship on a table.
type ForeignKey struct {
	Name       string
	Schema     string
	Type       *Type
	Field      *Field
	RefType    *Type
	RefField   *Field
	ForeignKey *models.ForeignKey
	Comment    string
}

// Index is a template item for a index into a table.
type Index struct {
	FuncName string
	Schema   string
	Type     *Type
	Fields   []*Field
	Index    *models.Index
	Comment  string
}

// QueryParam is a query parameter for a custom query.
type QueryParam struct {
	Name        string
	Type        string
	Interpolate bool
}

// Query is a template item for a custom query.
type Query struct {
	Schema        string
	Name          string
	Query         []string
	QueryComments []string
	QueryParams   []*QueryParam
	OnlyOne       bool
	Interpolate   bool
	Type          *Type
	Comment       string
}
