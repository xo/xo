// DB is the common interface for database operations that can be used with
// types from {{ .Schema }}.
//
// This should work with database/sql.DB and database/sql.Tx.
type DB interface {
	Exec(string, ...interface{}) (sql.Result, error)
	Query(string, ...interface{}) (*sql.Rows, error)
	QueryRow(string, ...interface{}) *sql.Row
}

