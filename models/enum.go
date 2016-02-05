package models

import "database/sql"

// Enum represents an Enum.
type Enum struct {
	Type       string
	Value      string
	ConstValue uint16

	// extras
	EnumType string
}

// EnumsBySchema returns enums from the database for the specified schema.
func EnumsBySchema(db *sql.DB, schema string) ([]*Enum, error) {
	const sqlstr = `SELECT t.typname, e.enumlabel, e.enumsortorder ` +
		`FROM pg_type t ` +
		`LEFT JOIN pg_namespace n ON n.oid = t.typnamespace ` +
		`JOIN pg_enum e ON t.oid = e.enumtypid ` +
		`WHERE n.nspname = $1`

	// run query
	q, err := db.Query(sqlstr, schema)
	if err != nil {
		return nil, err
	}
	defer q.Close()

	// load results
	res := []*Enum{}
	for q.Next() {
		e := Enum{}

		// scan
		err = q.Scan(
			&e.Type, &e.Value, &e.ConstValue,
		)
		if err != nil {
			return nil, err
		}

		res = append(res, &e)
	}

	return res, nil
}
