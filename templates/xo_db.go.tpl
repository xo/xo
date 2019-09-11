// XODB is the common interface for database operations that can be used with
// types from schema '{{ schema .Schema }}'.
//
// This should work with database/sql.DB and database/sql.Tx.
type XODB interface {
	Exec(string, ...interface{}) (sql.Result, error)
	Query(string, ...interface{}) (*sql.Rows, error)
	QueryRow(string, ...interface{}) *sql.Row
}

type XOTX interface {
  XODB
  Commit() error
  Rollback() error
}

// XOLog provides the log func used by generated queries.
var XOLog = func(s string, a ...interface{}) { }

// Helper function for doing transactions
func DoTransaction(tx XOTX, txFunc func(XOTX) error) (err error) {
  defer func() {
    if p := recover(); p != nil {
      tx.Rollback()
      panic(p) // re-throw panic after Rollback
    } else if err != nil {
      tx.Rollback() // err is non-nil; don't change it
    } else {
      err = tx.Commit() // err is nil; if Commit returns error update err
    }
  }()

  err = txFunc(tx)
  return err
}

// ScannerValuer is the common interface for types that implement both the
// database/sql.Scanner and sql/driver.Valuer interfaces.
type ScannerValuer interface {
	sql.Scanner
	driver.Valuer
}

// helper function taken directly from sql/convert.go

func asString(src interface{}) string {
  switch v := src.(type) {
  case string:
    return v
  case []byte:
    return string(v)
  }
  rv := reflect.ValueOf(src)
  switch rv.Kind() {
  case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
    return strconv.FormatInt(rv.Int(), 10)
  case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
    return strconv.FormatUint(rv.Uint(), 10)
  case reflect.Float64:
    return strconv.FormatFloat(rv.Float(), 'g', -1, 64)
  case reflect.Float32:
    return strconv.FormatFloat(rv.Float(), 'g', -1, 32)
  case reflect.Bool:
    return strconv.FormatBool(rv.Bool())
  }
  return fmt.Sprintf("%v", src)
}

type XoBigFloat struct {
  big.Float
}

func (xf *XoBigFloat) Scan(src interface{}) error {
  if src == nil {
    *xf = XoBigFloat{}
    return nil
  }
  str := asString(src)
  newxf, err := NewXoBigFloat(str)
  *xf = newxf
  return err
}

func (ss XoBigFloat) Value() (driver.Value, error) {
  return ss.String(), nil
}

type NullableXoBigFloat struct {
  XoBigFloat
  Valid bool
}

func (xf *NullableXoBigFloat) Scan(src interface{}) error {
  if src == nil {
    xf.XoBigFloat = XoBigFloat{}
    xf.Valid = false
    return nil
  }

  return (&xf.XoBigFloat).Scan(src)
}

func (ss NullableXoBigFloat) Value() (driver.Value, error) {
  if ss.Valid == false {
    return nil, nil
  }

  return ss.XoBigFloat.String(), nil
}

func NewXoBigFloat(strval string) (XoBigFloat, error) {
  newbf, _, err := big.ParseFloat(strval, 10, 53, big.ToNearestEven)
  return XoBigFloat{*newbf}, err
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

