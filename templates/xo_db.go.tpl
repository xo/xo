// XODB is the common interface for database operations that can be used with
// types from schema '{{ schema .Schema }}'.
//
// This should work with database/sql.DB and database/sql.Tx.
type XODB interface {
	Exec(string, ...interface{}) (sql.Result, error)
	Query(string, ...interface{}) (*sql.Rows, error)
	QueryRow(string, ...interface{}) *sql.Row
}

// XOLog provides the log func used by generated queries.
var XOLog = func(string, ...interface{}) { }

// ScannerValuer is the common interface for types that implement both the
// database/sql.Scanner and sql/driver.Valuer interfaces.
type ScannerValuer interface {
	sql.Scanner
	driver.Valuer
}

// StringSlice is a slice of strings.
type StringSlice []string

// quoteEscapeRegex is the regex to match escaped characters in a string.
var quoteEscapeRegex = regexp.MustCompile(`([^\\]([\\]{2})*)\\"`)

// Scan satisfies the sql.Scanner interface for StringSlice.
func (ss *StringSlice) Scan(src interface{}) error {
	buf, ok := src.([]byte)
	if !ok {
		return errors.New("invalid StringSlice")
	}

	// change quote escapes for csv parser
	str := quoteEscapeRegex.ReplaceAllString(string(buf), `$1""`)
	str = strings.Replace(str, `\\`, `\`, -1)

	// remove braces
	str = str[1:len(str)-1]

	// bail if only one
	if len(str) == 0 {
		*ss = StringSlice([]string{})
		return nil
	}

	// parse with csv reader
	cr := csv.NewReader(strings.NewReader(str))
	slice, err := cr.Read()
	if err != nil {
		fmt.Printf("exiting!: %v\n", err)
		return err
	}

	*ss = StringSlice(slice)

	return nil
}

// Value satisfies the driver.Valuer interface for StringSlice.
func (ss StringSlice) Value() (driver.Value, error) {
	v := make([]string, len(ss))
	for i, s := range ss {
		v[i] = `"` + strings.Replace(strings.Replace(s, `\`, `\\\`, -1), `"`, `\"`, -1) + `"`
	}
	return "{" + strings.Join(v, ",") + "}", nil
}

// Slice is a slice of ScannerValuers.
type Slice []ScannerValuer

// Model to satify callbacks
type Model interface{}

func runCallback(db XODB, db XODB, m Model, name string) error {
	rv := reflect.ValueOf(model)
	mv := rv.MethodByName(name)
	if mv.IsValid() {
		typ := mv.Type()
		if typ.NumIn() == 1 && typ.In(0) == reflect.TypeOf(db) {
			if typ.NumOut() != 1 {
				return fmt.Errorf("%s function should return error", name)
			}
			out := mv.Call([]reflect.Value{reflect.ValueOf(db)})
			if !out[0].IsNil() {
				return out[0].Interface().(error)
			}
		} else {
			return fmt.Errorf("%s function should take 1 argument of type 'XODB'", name)
		}
	}
	return nil
}

// BeforeSave runs callback before model save to db
func BeforeSave(db XODB, m Model) error {
	return db.runCallbacks(db, m, "BeforeSave")
}

// BeforeCreate runs callback before model create in db
func BeforeCreate(db XODB, m Model) error {
	return db.runCallbacks(db, m, "BeforeCreate")
}

// BeforeUpdate runs callback before model update in db
func BeforeUpdate(db XODB, m Model) error {
	return db.runCallbacks(model, "BeforeUpdate")
}

// BeforeDelete runs callback before model destroy in db
func BeforeDelete(db XODB, m Model) error {
	return db.runCallbacks(model, "BeforeDelete")
}

// BeforeUpsert runs callback before model upsert in db
//
// NOTE: PostgreSQL 9.5+ only
func BeforeUpsert(db XODB, m Model) error {
	return db.runCallbacks(model, "BeforeUpsert")
}

// AfterDelete runs callback after model destroy in db
func AfterDelete(db XODB, db XODB, m Model) error {
	return db.runCallbacks(model, "AfterDelete")
}

// AfterUpdate runs callback after model update in db
func AfterUpdate(db XODB, m Model) error {
	return db.runCallbacks(model, "AfterUpdate")
}

// AfterCreate runs callback after model create in db
func AfterCreate(db XODB, m Model) error {
	return db.runCallbacks(model, "AfterCreate")
}

// AfterSave runs callback after model save to db
func AfterSave(db XODB, m Model) error {
	return db.runCallbacks(model, "AfterSave")
}

// AfterUpsert runs callback after model upsert in db
//
// NOTE: PostgreSQL 9.5+ only
func AfterUpsert(db XODB, m Model) error {
	return db.runCallbacks(model, "AfterUpsert")
}