package models

import "database/sql"

// ForeignKey represents a ForeignKey.
type ForeignKey struct {
	Name          string
	TableName     string
	ColumnName    string
	RefIndexName  string
	RefTableName  string
	RefColumnName string
}

// ForeignKeysBySchema returns the foreign keys from the database for the
// specified schema.
func ForeignKeysBySchema(db *sql.DB, schema string) ([]*ForeignKey, error) {
	// sql query
	const sqlstr = `SELECT r.conname, a.relname, b.attname, ` +
		`i.relname, c.relname, d.attname ` +
		`FROM pg_constraint r ` +
		`JOIN ONLY pg_class a ON a.oid = r.conrelid ` +
		`JOIN ONLY pg_attribute b ON b.attnum = ANY(r.conkey) AND b.attrelid = r.conrelid ` +
		`JOIN ONLY pg_class i on i.oid = r.conindid ` +
		`JOIN ONLY pg_class c on c.oid = r.confrelid ` +
		`JOIN ONLY pg_attribute d ON d.attnum = ANY(r.confkey) AND d.attrelid = r.confrelid ` +
		`JOIN ONLY pg_namespace n ON n.oid = r.connamespace ` +
		`WHERE r.contype = 'f' AND n.nspname = $1`

	// run query
	q, err := db.Query(sqlstr, schema)
	if err != nil {
		return nil, err
	}
	defer q.Close()

	// load results
	res := []*ForeignKey{}
	for q.Next() {
		fk := ForeignKey{}

		// scan
		err = q.Scan(
			&fk.Name, &fk.TableName, &fk.ColumnName,
			&fk.RefIndexName, &fk.RefTableName, &fk.RefColumnName,
		)
		if err != nil {
			return nil, err
		}

		res = append(res, &fk)
	}

	return res, nil
}
