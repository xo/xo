package gotpl

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

// Proc is a stored procedure template.
type Proc struct {
	GoName     string
	SQLName    string
	Signature  string
	Params     []Field
	Return     Field
	ReturnType string
	Comment    string
}

// Field is a field template.
type Field struct {
	GoName    string
	SQLName   string
	Type      string
	Zero      string
	Prec      int
	Array     bool
	IsPrimary bool
	Comment   string
}

// Table is a type (ie, table/view/custom query) template.
type Table struct {
	GoName      string
	SQLName     string
	Kind        string
	PrimaryKeys []Field
	Fields      []Field
	Manual      bool
	Comment     string
}

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

// Index is an index template.
type Index struct {
	SQLName   string
	FuncName  string
	Table     Table
	Fields    []Field
	IsUnique  bool
	IsPrimary bool
	Comment   string
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
	Params      []QueryParam
	One         bool
	Flat        bool
	Exec        bool
	Interpolate bool
	Type        Table
	Comment     string
}
