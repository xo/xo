package models

import "database/sql"

// A Column is a single information_schema.columns row.
type Column struct {
	TableName        string
	ColumnName       string
	DataType         string
	FieldOrdinal     uint16
	IsNullable       bool
	IsPrimaryKey     bool
	IsIndex          bool
	IsUnique         bool
	HasDefault       bool
	DefaultValue     string
	IndexName        string
	ForeignIndexName string

	// extras
	Field     string
	GoType    string
	GoNilType string
	Tag       string
	Len       int
}

// ColumnsByTableSchema retrieves all the Column entries having the specified
// tableSchema.
func ColumnsByTableSchema(db *sql.DB, tableSchema string) ([]*Column, error) {
	var err error

	// sql query
	const sqlstr = `SELECT c.relname, f.attname, format_type(f.atttypid, f.atttypmod), ` + // table name, column name, data type
		`f.attnum, NOT f.attnotnull, ` + // field ordinal, is nullable
		`CASE WHEN p.contype = 'p' THEN true ELSE false END, ` + // is primary key
		`CASE WHEN i.oid <> 0 THEN true ELSE false END, ` + // is index
		`CASE WHEN p.contype = 'u' THEN true WHEN p.contype = 'p' THEN true ELSE false END, ` + // is unique
		`CASE WHEN f.atthasdef = 't' THEN true ELSE false END, ` + // has default
		`substring(COALESCE(pg_get_expr(d.adbin, d.adrelid), '') FOR 128), ` + // default value
		`COALESCE(i.relname, ''), ` + // index name
		`CASE WHEN p.contype = 'f' THEN p.conname ELSE '' END ` + // foreign index name
		`FROM pg_class c ` +
		`JOIN pg_attribute f ON c.oid = f.attrelid ` +
		`JOIN pg_type t ON f.atttypid = t.oid ` +
		`LEFT JOIN pg_attrdef d ON d.adrelid = c.oid AND d.adnum = f.attnum ` +
		`LEFT JOIN pg_namespace n ON n.oid = c.relnamespace ` +
		`LEFT JOIN pg_constraint p ON p.conrelid = c.oid AND f.attnum = ANY (p.conkey) ` +
		`LEFT JOIN pg_class g ON p.confrelid = g.oid ` +
		`LEFT JOIN pg_index ix ON f.attnum = ANY(ix.indkey) AND c.oid = f.attrelid AND c.oid = ix.indrelid ` +
		`LEFT JOIN pg_class i ON ix.indexrelid = i.oid ` +
		`WHERE c.relkind = 'r' AND f.attnum > 0 AND n.nspname = $1 ` +
		`ORDER BY c.relname, f.attnum`

	// run query
	q, err := db.Query(sqlstr, tableSchema)
	if err != nil {
		return nil, err
	}
	defer q.Close()

	// load results
	res := []*Column{}
	for q.Next() {
		c := Column{}

		// scan
		err = q.Scan(
			&c.TableName, &c.ColumnName, &c.DataType,
			&c.FieldOrdinal, &c.IsNullable, &c.IsPrimaryKey,
			&c.IsIndex, &c.IsUnique, &c.HasDefault,
			&c.DefaultValue, &c.IndexName, &c.ForeignIndexName,
		)

		// check err
		if err != nil {
			return nil, err
		}

		res = append(res, &c)
	}

	return res, nil
}
