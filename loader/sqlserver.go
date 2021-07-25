package loader

import (
	"context"
	"regexp"

	"github.com/xo/xo/models"
	xo "github.com/xo/xo/types"
)

func init() {
	Register(&Loader{
		Driver:           "sqlserver",
		Mask:             "@p%d",
		GoType:           SqlserverGoType,
		Schema:           models.SqlserverSchema,
		Procs:            models.SqlserverProcs,
		ProcParams:       models.SqlserverProcParams,
		Tables:           models.SqlserverTables,
		TableColumns:     models.SqlserverTableColumns,
		TableSequences:   models.SqlserverTableSequences,
		TableForeignKeys: models.SqlserverTableForeignKeys,
		TableIndexes:     models.SqlserverTableIndexes,
		IndexColumns:     models.SqlserverIndexColumns,
		ViewStrip:        SqlserverViewStrip,
		ViewCreate:       models.SqlserverViewCreate,
		ViewDrop:         models.SqlserverViewDrop,
	})
}

// SqlserverGoType parse a mssql type into a Go type based on the column
// definition.
func SqlserverGoType(ctx context.Context, d xo.Datatype) (string, string, error) {
	typ, nullable := d.Type, d.Nullable
	var goType, zero string
	switch typ {
	case "tinyint", "bit":
		goType, zero = "bool", "false"
		if nullable {
			goType, zero = "sql.NullBool", "sql.NullBool{}"
		}
	case "char", "money", "nchar", "ntext", "nvarchar", "smallmoney", "text", "varchar":
		goType, zero = "string", `""`
		if nullable {
			goType, zero = "sql.NullString", "sql.NullString{}"
		}
	case "smallint":
		goType, zero = "int16", "0"
		if nullable {
			goType, zero = "sql.NullInt64", "sql.NullInt64{}"
		}
	case "int":
		goType, zero = Int32(ctx), "0"
		if nullable {
			goType, zero = "sql.NullInt64", "sql.NullInt64{}"
		}
	case "bigint":
		goType, zero = "int64", "0"
		if nullable {
			goType, zero = "sql.NullInt64", "sql.NullInt64{}"
		}
	case "real":
		goType, zero = "float32", "0.0"
		if nullable {
			goType, zero = "sql.NullFloat64", "sql.NullFloat64{}"
		}
	case "numeric", "decimal", "float":
		goType, zero = "float64", "0.0"
		if nullable {
			goType, zero = "sql.NullFloat64", "sql.NullFloat64{}"
		}
	case "binary", "image", "varbinary", "xml":
		goType, zero = "[]byte", "nil"
	case "date", "time", "smalldatetime", "datetime", "datetime2", "datetimeoffset":
		goType, zero = "time.Time", "time.Time{}"
		if nullable {
			goType, zero = "sql.NullTime", "sql.NullTime{}"
		}
	default:
		goType, zero = SchemaGoType(ctx, typ, nullable)
	}
	return goType, zero, nil
}

// SqlserverViewStrip strips ORDER BY clauses from the passed query.
func SqlserverViewStrip(query []string) ([]string, []string) {
	// sqlserver cannot have an 'ORDER BY' clause in a CREATE VIEW
	var inspect []string
	for _, line := range query {
		if orderByRE.MatchString(line) {
			continue
		}
		inspect = append(inspect, line)
	}
	return inspect, make([]string, len(inspect))
}

// orderByRE is a regexp matching ORDER by clauses in sqlserver queries.
var orderByRE = regexp.MustCompile(`(?i)^\s*ORDER\s+BY\s+`)
