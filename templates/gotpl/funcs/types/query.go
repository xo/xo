// Package types provides schema types.
package types

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

// QueryParam is a custom query parameter template.
type QueryParam struct {
	Name        string
	Type        string
	Interpolate bool
	Join        bool
}

func (q Query) GetName() string {
	return q.Name
}
