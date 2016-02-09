// XODB is the common interface for database operations that can be used with
// types from {{ .Schema }}.
//
// This should work with database/sql.DB and database/sql.Tx.
type XODB interface {
	Exec(string, ...interface{}) (sql.Result, error)
	Query(string, ...interface{}) (*sql.Rows, error)
	QueryRow(string, ...interface{}) *sql.Row
}

// ScannerValuer is a type that implements both the sql.Scanner and
// driver.Valuer interfaces.
type ScannerValuer {
    sql.Scanner
    driver.Valuer
}

// Slice is a slice of ScannerValuers.
type Slice []ScannerValuer

type QueryOpt interface {

}

