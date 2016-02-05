package models

import "database/sql"

// A Column is a single information_schema.columns row.
type Column struct {
	ColumnName       string
	TableName        string
	DataType         string
	FieldOrdinal     uint16
	IsNullable       bool
	IsIndex          bool
	IsUnique         bool
	IsPrimaryKey     bool
	IsForeignKey     bool
	IndexName        string
	ForeignIndexName string
	HasDefault       bool
	DefaultValue     string

	// extras
	Field   string
	Type    string
	NilType string
	Tag     string
	Len     int
}

// ColumnsByTableSchema retrieves all the Column entries having the specified
// tableSchema.
func ColumnsByTableSchema(db *sql.DB, tableSchema string) ([]*Column, error) {
	var err error

	// sql query
	const sqlstr = `SELECT a.attname, ` + // column name
		`c.relname, ` + // table name
		`format_type(a.atttypid, a.atttypmod), ` + // data type
		`a.attnum, ` + // field ordinal
		`NOT a.attnotnull, ` + // is nullable
		`COALESCE(i.oid <> 0, false), ` + // is index
		`COALESCE(ct.contype = 'u' OR ct.contype = 'p', false), ` + // is unique
		`COALESCE(ct.contype = 'p', false), ` + // is primary key
		`COALESCE(cf.contype = 'f', false), ` + // is foreign key
		`COALESCE(i.relname, ''), ` + // index name
		`COALESCE(cf.conname, ''), ` + // foreign index name
		`a.atthasdef, ` + // has default
		`COALESCE(pg_get_expr(ad.adbin, ad.adrelid), '') ` + // default value
		`FROM pg_attribute a ` +
		`JOIN ONLY pg_class c ON c.oid = a.attrelid ` +
		`JOIN ONLY pg_namespace n ON n.oid = c.relnamespace ` +
		`LEFT JOIN pg_constraint ct ON ct.conrelid = c.oid AND a.attnum = ANY(ct.conkey) AND ct.contype IN('p', 'u') ` +
		`LEFT JOIN pg_constraint cf ON cf.conrelid = c.oid AND a.attnum = ANY(cf.conkey) AND cf.contype IN('f') ` +
		`LEFT JOIN pg_attrdef ad ON ad.adrelid = c.oid AND ad.adnum = a.attnum ` +
		`LEFT JOIN pg_index ix ON a.attnum = ANY(ix.indkey) AND c.oid = a.attrelid AND c.oid = ix.indrelid ` +
		`LEFT JOIN pg_class i ON i.oid = ix.indexrelid ` +
		`WHERE c.relkind = 'r' AND a.attnum > 0 AND n.nspname = $1 ` +
		`ORDER BY c.relname, a.attnum `

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
			&c.ColumnName, &c.TableName, &c.DataType, &c.FieldOrdinal, &c.IsNullable,
			&c.IsIndex, &c.IsUnique, &c.IsPrimaryKey, &c.IsForeignKey,
			&c.IndexName, &c.ForeignIndexName,
			&c.HasDefault, &c.DefaultValue,
		)
		if err != nil {
			return nil, err
		}

		res = append(res, &c)
	}

	return res, nil
}
