package internal

import (
	"database/sql"
	"net/url"
	"strconv"
	"strings"
)

// Loader is the common interface for database drivers that can generate code
// from a database schema.
type Loader interface {
	IsSupported(*url.URL) (string, bool)
	ParseQuery(*ArgType, *sql.DB) error
	LoadSchemaTypes(*ArgType, *sql.DB) error
	NthParam(i int) string
}

// TypeLoader provides a Loader.
type TypeLoader struct {
	Schemes        []string
	ProcessDSN     func(*url.URL, string) string
	QueryFunc      func(*ArgType, *sql.DB) error
	LoadSchemaFunc func(*ArgType, *sql.DB) error
	ParamN         func(int) string
}

// IsSupported returns whether or not the url is supported, and modifies the
// URL to have the correct scheme and returns a correct DSN string for
// sql.Open.
func (tl TypeLoader) IsSupported(u *url.URL) (string, bool) {
	uscheme := strings.ToLower(u.Scheme)
	protocol := "tcp"

	// check if +unix or whatever is in the scheme
	if strings.Contains(uscheme, "+") {
		p := strings.SplitN(uscheme, "+", 2)
		uscheme = p[0]
		protocol = p[1]
	}

	var found bool
	for _, s := range tl.Schemes {
		if uscheme == strings.ToLower(s) {
			found = true
			break
		}
	}

	if !found {
		return "", false
	}

	// fix scheme
	u.Scheme = tl.Schemes[0]

	// process dsn if func is non-nil
	if tl.ProcessDSN != nil {
		return tl.ProcessDSN(u, protocol), true
	}

	return u.String(), true
}

// ParseQuery satisfies Loader's ParseQuery.
func (tl TypeLoader) ParseQuery(args *ArgType, db *sql.DB) error {
	return tl.QueryFunc(args, db)
}

// LoadSchemaTypes satisfies Loader's LoadSchemaTypes.
func (tl TypeLoader) LoadSchemaTypes(args *ArgType, db *sql.DB) error {
	return tl.LoadSchemaFunc(args, db)
}

// NthParam satisifies Loader's NthParam.
func (tl TypeLoader) NthParam(i int) string {
	if tl.ParamN != nil {
		return tl.ParamN(i)
	}

	return "$" + strconv.Itoa(i+1)
}

// SchemaLoaders are the available schema loaders.
var SchemaLoaders = map[string]Loader{}
