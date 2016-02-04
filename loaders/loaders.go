package loaders

import (
	"bytes"
	"database/sql"

	"github.com/knq/xo/internal"
)

// Loader is the interface for loading templated database stuff.
type Loader func(*internal.ArgType, *sql.DB, map[string]*bytes.Buffer) error

// GetBuf retrieves previously defined byte.Buffer with name from m, creating
// it a new byte.Buffer if necessary.
func GetBuf(m map[string]*bytes.Buffer, name string) *bytes.Buffer {
	buf, ok := m[name]
	if !ok {
		m[name] = new(bytes.Buffer)
		return m[name]
	}

	return buf
}
