package loaders

import (
	"bytes"
	"database/sql"

	"github.com/knq/xo/internal"
)

// GetBuf is a utility func to retrieve a previously defined byte.Buffer with
// name from m, creating a new byte.Buffer if necessary.
func GetBuf(m map[string]*bytes.Buffer, name string) *bytes.Buffer {
	buf, ok := m[name]
	if !ok {
		m[name] = new(bytes.Buffer)
		return m[name]
	}

	return buf
}

// Driver is the common interface for database drivers that can generate code
// from a database schema.
type Driver interface {
	ParseQuery(*internal.ArgType, *sql.DB, map[string]*bytes.Buffer) error
	LoadSchemaTypes(*internal.ArgType, *sql.DB, map[string]*bytes.Buffer) error
}

// LoadFunc is the func signature for loading types from a database schema.
type LoadFunc func(*internal.ArgType, *sql.DB, map[string]*bytes.Buffer) error

// Loader is a handle
type Loader struct {
	QueryFunc      LoadFunc
	LoadSchemaFunc LoadFunc
}

// ParseQuery satisfies Driver's ParseQuery.
func (l Loader) ParseQuery(args *internal.ArgType, db *sql.DB, typeMap map[string]*bytes.Buffer) error {
	return l.QueryFunc(args, db, typeMap)
}

// LoadSchemaTypes satisfies Driver's LoadSchemaTypes.
func (l Loader) LoadSchemaTypes(args *internal.ArgType, db *sql.DB, typeMap map[string]*bytes.Buffer) error {
	return l.LoadSchemaFunc(args, db, typeMap)
}
