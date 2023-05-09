{{ define "db" -}}
var (
	// logf is used by generated code to log SQL queries.
	logf = func(string, ...interface{}) {}
	// errf is used by generated code to log SQL errors.
	errf = func(string, ...interface{}) {}
)

// logerror logs the error and returns it.
func logerror(err error) error {
	errf("ERROR: %v", err)
	return err
}

// Logf logs a message using the package logger.
func Logf(s string, v ...interface{}) {
	logf(s, v...)
}

// SetLogger sets the package logger. Valid logger types:
//
//     io.Writer
//     func(string, ...interface{}) (int, error) // fmt.Printf
//     func(string, ...interface{}) // log.Printf
//
func SetLogger(logger interface{}) {
	logf = convLogger(logger)
}

// Errorf logs an error message using the package error logger.
func Errorf(s string, v ...interface{}) {
	errf(s, v...)
}

// SetErrorLogger sets the package error logger. Valid logger types:
//
//     io.Writer
//     func(string, ...interface{}) (int, error) // fmt.Printf
//     func(string, ...interface{}) // log.Printf
//
func SetErrorLogger(logger interface{}) {
	errf = convLogger(logger)
}

// convLogger converts logger to the standard logger interface.
func convLogger(logger interface{}) func(string, ...interface{}) {
	switch z := logger.(type) {
	case io.Writer:
		return func(s string, v ...interface{}) {
			fmt.Fprintf(z, s, v...)
		}
	case func(string, ...interface{}) (int, error): // fmt.Printf
		return func(s string, v ...interface{}) {
			_, _ = z(s, v...)
		}
	case func(string, ...interface{}): // log.Printf
		return z
	}
	panic(fmt.Sprintf("unsupported logger type %T", logger))
}

// DB is the common interface for database operations that can be used with
// types from schema '{{ schema }}'.
//
// This works with both [database/sql.DB] and [database/sql.Tx].
type DB interface {
{{ if context -}}
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
{{- end -}}{{- if or context_both context_disable }}
	Exec(string, ...interface{}) (sql.Result, error)
	Query(string, ...interface{}) (*sql.Rows, error)
	QueryRow(string, ...interface{}) *sql.Row
{{- end }}
}

// Error is an error.
type Error string

// Error satisfies the error interface.
func (err Error) Error() string {
	return string(err)
}

// Error values.
const (
	// ErrAlreadyExists is the already exists error.
	ErrAlreadyExists Error = "already exists"
	// ErrDoesNotExist is the does not exist error.
	ErrDoesNotExist Error = "does not exist"
	// ErrMarkedForDeletion is the marked for deletion error.
	ErrMarkedForDeletion Error = "marked for deletion"
)

// ErrInsertFailed is the insert failed error.
type ErrInsertFailed struct {
	Err error
}

// Error satisfies the error interface.
func (err *ErrInsertFailed) Error() string {
	return fmt.Sprintf("insert failed: %v", err.Err)
}

// Unwrap satisfies the unwrap interface.
func (err *ErrInsertFailed) Unwrap() error {
	return err.Err
}

// ErrUpdateFailed is the update failed error.
type ErrUpdateFailed struct {
	Err error
}

// Error satisfies the error interface.
func (err *ErrUpdateFailed) Error() string {
	return fmt.Sprintf("update failed: %v", err.Err)
}

// Unwrap satisfies the unwrap interface.
func (err *ErrUpdateFailed) Unwrap() error {
	return err.Err
}

// ErrUpsertFailed is the upsert failed error.
type ErrUpsertFailed struct {
	Err error
}

// Error satisfies the error interface.
func (err *ErrUpsertFailed) Error() string {
	return fmt.Sprintf("upsert failed: %v", err.Err)
}

// Unwrap satisfies the unwrap interface.
func (err *ErrUpsertFailed) Unwrap() error {
	return err.Err
}

{{ if driver "sqlite3" -}}
// ErrInvalidTime is the invalid Time error.
type ErrInvalidTime string

// Error satisfies the error interface.
func (err ErrInvalidTime) Error() string {
	return fmt.Sprintf("invalid Time (%s)", string(err))
}

// Time is a SQLite3 Time that scans for the various timestamps values used by
// SQLite3 database drivers to store time.Time values.
type Time struct {
	time time.Time
}

// NewTime creates a time.
func NewTime(t time.Time) Time {
	return Time{time: t}
}

// String satisfies the fmt.Stringer interface.
func (t Time) String() string {
	return t.time.String()
}

// Format formats the time.
func (t Time) Format(layout string) string {
	return t.time.Format(layout)
}

// Time returns a time.Time.
func (t Time) Time() time.Time {
	return t.time
}

// Value satisfies the sql/driver.Valuer interface.
func (t Time) Value() (driver.Value, error) {
	return t.time, nil
}

// Scan satisfies the sql.Scanner interface.
func (t *Time) Scan(v interface{}) error {
	switch x := v.(type) {
	case time.Time:
		t.time = x
		return nil
	case []byte:
		return t.Parse(string(x))
	case string:
		return t.Parse(x)
	}
	return ErrInvalidTime(fmt.Sprintf("%T", v))
}

// Parse attempts to Parse string s to t.
func (t *Time) Parse(s string) error {
	if s == "" {
		return nil
	}
	for _, f := range TimestampFormats {
		if z, err := time.Parse(f, s); err == nil {
			t.time = z
			return nil
		}
	}
	return ErrInvalidTime(s)
}

// MarshalJSON satisfies the [json.Marshaler] interface.
func (t Time) MarshalJSON() ([]byte, error) {
	return t.time.MarshalJSON()
}

// UnmarshalJSON satisfies the [json.Unmarshaler] interface.
func (t *Time) UnmarshalJSON(data []byte) error {
	return t.time.UnmarshalJSON(data)
}

// TimestampFormats are the timestamp formats used by SQLite3 database drivers
// to store a time.Time in SQLite3.
//
// The first format in the slice will be used when saving time values into the
// database.  When parsing a string from a timestamp or datetime column, the
// formats are tried in order.
var TimestampFormats = []string{
	// By default, use timestamps with the timezone they have. When parsed,
	// they will be returned with the same timezone.
	"2006-01-02 15:04:05.999999999-07:00",
	"2006-01-02T15:04:05.999999999-07:00",
	"2006-01-02 15:04:05.999999999",
	"2006-01-02T15:04:05.999999999",
	"2006-01-02 15:04:05",
	"2006-01-02T15:04:05",
	"2006-01-02 15:04",
	"2006-01-02T15:04",
	"2006-01-02",
}
{{- end }}
{{- end }}
