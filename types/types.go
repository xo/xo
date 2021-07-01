package types

// ContextKey is a context key.
type ContextKey string

// Context key values.
const (
	// database and loader keys
	DriverKey   ContextKey = "driver"
	SchemaKey   ContextKey = "schema"
	DbKey       ContextKey = "db"
	EmitterKey  ContextKey = "emitter"
	LoaderKey   ContextKey = "loader"
	NthParamKey ContextKey = "nth-param"
	// type keys
	Int32Key  ContextKey = "int32"
	Uint32Key ContextKey = "uint32"
)

// XO is the data from introspection.
type XO struct {
	Queries []Query  `json:"queries,omitempty"`
	Schemas []Schema `json:"schemas,omitempty"`
}

// Emit adds v.
func (xo *XO) Emit(v interface{}) {
	switch x := v.(type) {
	case Query:
		xo.Queries = append(xo.Queries, x)
	case Schema:
		xo.Schemas = append(xo.Schemas, x)
	}
}

// Query is a query.
type Query struct {
	Driver       string   `json:"driver,omitempty"`
	Name         string   `json:"name,omitempty"`
	Comment      string   `json:"comment,omitempty"`
	Exec         bool     `json:"exec,omitempty"`
	Flat         bool     `json:"flat,omitempty"`
	One          bool     `json:"one,omitempty"`
	Interpolate  bool     `json:"interpolate,omitempty"`
	Type         string   `json:"type,omitempty"`
	TypeComment  string   `json:"type_comment,omitempty"`
	Fields       []Field  `json:"fields,omitempty"`
	ManualFields bool     `json:"manual_fields,omitempty"` // fields generated or provided by user
	Params       []Field  `json:"params,omitempty"`
	Query        []string `json:"query,omitempty"`
	Comments     []string `json:"comments,omitempty"`
}

// Schema is a SQL schema.
type Schema struct {
	Driver string  `json:"type,omitempty"`
	Name   string  `json:"name,omitempty"`
	Enums  []Enum  `json:"enums,omitempty"`
	Procs  []Proc  `json:"procs,omitempty"`
	Tables []Table `json:"tables,omitempty"`
	Views  []Table `json:"views,omitempty"`
}

// Enum is a enum type.
type Enum struct {
	Name    string  `json:"name,omitempty"`
	Values  []Field `json:"values,omitempty"`
	Comment string  `json:"comment,omitempty"`
}

// Proc is a stored procedure.
type Proc struct {
	Name    string  `json:"name,omitempty"`
	Params  []Field `json:"params,omitempty"`
	Return  Field   `json:"return,omitempty"`
	Comment string  `json:"comment,omitempty"`
}

// Table is a table or view.
type Table struct {
	Type        string       `json:"type,omitempty"` // 'table' or 'view'
	Name        string       `json:"name,omitempty"`
	Columns     []Field      `json:"columns,omitempty"`
	PrimaryKeys []Field      `json:"primary_keys,omitempty"`
	Indexes     []Index      `json:"indexes,omitempty"`
	ForeignKeys []ForeignKey `json:"foreign_keys,omitempty"`
	Manual      bool         `json:"manual,omitempty"`
	Comment     string       `json:"comment,omitempty"`
}

// Index is a index.
type Index struct {
	Name      string  `json:"name,omitempty"`
	FuncName  string  `json:"func_name,omitempty"`
	Fields    []Field `json:"fields,omitempty"`
	IsUnique  bool    `json:"is_unique,omitempty"`
	IsPrimary bool    `json:"is_primary,omitempty"`
}

// ForeignKey is a foreign key.
type ForeignKey struct {
	Name         string `json:"name,omitempty"`          // constraint name
	ResolvedName string `json:"resolved_name,omitempty"` // foreign key name (based on fkey mode)
	Field        Field  `json:"column,omitempty"`        // column that has the key on it
	RefIndex     string `json:"ref_index,omitempty"`     // name of the index on the foreign table
	RefTable     string `json:"ref_table,omitempty"`     // table the foreign key refers to
	RefField     Field  `json:"ref_column,omitempty"`    // column in ref table the index refers to
	RefFuncName  string `json:"ref_func_name"`           // func name from ref index

}

// Field is a column, index, enum value, or stored procedure parameter.
type Field struct {
	Name        string   `json:"name,omitempty"`
	Datatype    Datatype `json:"datatype,omitempty"`
	Default     *string  `json:"default,omitempty"`
	Comment     string   `json:"comment,omitempty"`
	IsPrimary   bool     `json:"is_primary,omitempty"`
	IsSequence  bool     `json:"is_sequence,omitempty"`
	ConstValue  *int     `json:"const_value,omitempty"`
	Interpolate bool     `json:"interpolate,omitempty"`
	Join        bool     `json:"join,omitempty"`
}

// Datatype is a datatype.
type Datatype struct {
	Type     string `json:"type,omitempty"`
	Prec     int    `json:"prec,omitempty"`
	Scale    int    `json:"scale,omitempty"`
	Nullable bool   `json:"nullable,omitempty"`
	Array    bool   `json:"array,omitempty"`
}
