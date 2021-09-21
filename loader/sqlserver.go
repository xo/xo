package loader

import (
	"regexp"

	"github.com/xo/xo/models"
	xo "github.com/xo/xo/types"
)

func init() {
	Register("sqlserver", Loader{
		Mask:             "@p%d",
		Schema:           models.SqlserverSchema,
		Procs:            models.SqlserverProcs,
		ProcParams:       models.SqlserverProcParams,
		Tables:           models.SqlserverTables,
		TableColumns:     models.SqlserverTableColumns,
		TableSequences:   models.SqlserverTableSequences,
		TableForeignKeys: models.SqlserverTableForeignKeys,
		TableIndexes:     models.SqlserverTableIndexes,
		IndexColumns:     models.SqlserverIndexColumns,
		ViewCreate:       models.SqlserverViewCreate,
		ViewDrop:         models.SqlserverViewDrop,
		ViewStrip:        SqlserverViewStrip,
	})
}

// SqlserverGoType parse a mssql type into a Go type based on the column
// definition.
func SqlserverGoType(d xo.Type, schema, itype, utype string) (string, string, error) {
	var goType, zero string
	switch d.Type {
	case "tinyint", "bit":
		goType, zero = "bool", "false"
		if d.Nullable {
			goType, zero = "sql.NullBool", "sql.NullBool{}"
		}
	case "char", "money", "nchar", "ntext", "nvarchar", "smallmoney", "text", "varchar":
		goType, zero = "string", `""`
		if d.Nullable {
			goType, zero = "sql.NullString", "sql.NullString{}"
		}
	case "smallint":
		goType, zero = "int16", "0"
		if d.Nullable {
			goType, zero = "sql.NullInt64", "sql.NullInt64{}"
		}
	case "int":
		goType, zero = itype, "0"
		if d.Nullable {
			goType, zero = "sql.NullInt64", "sql.NullInt64{}"
		}
	case "bigint":
		goType, zero = "int64", "0"
		if d.Nullable {
			goType, zero = "sql.NullInt64", "sql.NullInt64{}"
		}
	case "real":
		goType, zero = "float32", "0.0"
		if d.Nullable {
			goType, zero = "sql.NullFloat64", "sql.NullFloat64{}"
		}
	case "numeric", "decimal", "float":
		goType, zero = "float64", "0.0"
		if d.Nullable {
			goType, zero = "sql.NullFloat64", "sql.NullFloat64{}"
		}
	case "binary", "image", "varbinary", "xml":
		goType, zero = "[]byte", "nil"
	case "date", "time", "smalldatetime", "datetime", "datetime2", "datetimeoffset":
		goType, zero = "time.Time", "time.Time{}"
		if d.Nullable {
			goType, zero = "sql.NullTime", "sql.NullTime{}"
		}
	default:
		goType, zero = schemaType(d.Type, d.Nullable, schema)
	}
	return goType, zero, nil
}

// SqlserverViewStrip strips ORDER BY clauses from the passed query.
func SqlserverViewStrip(query, inspect []string) ([]string, []string, []string, error) {
	// sqlserver cannot have an 'ORDER BY' clause in a CREATE VIEW
	var res []string
	for _, line := range inspect {
		if orderByRE.MatchString(line) {
			continue
		}
		res = append(res, line)
	}
	return query, res, make([]string, len(query)), nil
}

// orderByRE is a regexp matching ORDER by clauses in sqlserver queries.
var orderByRE = regexp.MustCompile(`(?i)^\s*ORDER\s+BY\s+`)
