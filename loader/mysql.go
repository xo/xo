package loader

import (
	"context"
	"regexp"
	"strings"

	"github.com/xo/xo/models"
	xo "github.com/xo/xo/types"
)

func init() {
	Register(&Loader{
		Driver:           "mysql",
		Mask:             "?",
		GoType:           MysqlGoType,
		Schema:           models.MysqlSchema,
		Enums:            models.MysqlEnums,
		EnumValues:       MysqlEnumValues,
		Procs:            models.MysqlProcs,
		ProcParams:       models.MysqlProcParams,
		Tables:           models.MysqlTables,
		TableColumns:     models.MysqlTableColumns,
		TableSequences:   models.MysqlTableSequences,
		TableForeignKeys: models.MysqlTableForeignKeys,
		TableIndexes:     models.MysqlTableIndexes,
		IndexColumns:     models.MysqlIndexColumns,
		ViewCreate:       models.MysqlViewCreate,
		ViewDrop:         models.MysqlViewDrop,
	})
}

// MysqlGoType parse a mysql type into a Go type based on the column
// definition.
func MysqlGoType(ctx context.Context, d xo.Datatype) (string, string, error) {
	typ, nullable, prec := d.Type, d.Nullable, d.Prec
	// extract unsigned
	unsigned := false
	if strings.HasSuffix(typ, " unsigned") {
		typ, unsigned = typ[:len(typ)-len(" unsigned")], true
	}
	var goType, zero string
	switch typ {
	case "bit":
		switch {
		case prec == 1 && !nullable:
			goType, zero = "bool", "false"
		case prec == 1 && nullable:
			goType, zero = "sql.NullBool", "sql.NullBool{}"
		case prec <= 8 && !nullable:
			goType, zero = "uint8", "0"
		case prec <= 16 && !nullable:
			goType, zero = "uint16", "0"
		case prec <= 32 && !nullable:
			goType, zero = "uint32", "0"
		case nullable:
			goType, zero = "sql.NullInt64", "sql.NullInt64{}"
		default:
			goType, zero = "uint64", "0"
		}
	case "bool", "boolean":
		goType, zero = "bool", "false"
		if nullable {
			goType, zero = "sql.NullBool", "sql.NullBool{}"
		}
	case "char", "varchar", "tinytext", "text", "mediumtext", "longtext":
		goType, zero = "string", `""`
		if nullable {
			goType, zero = "sql.NullString", "sql.NullString{}"
		}
	case "tinyint":
		switch {
		case prec == 1 && !nullable: // force tinyint(1) as bool
			goType, zero = "bool", "false"
		case prec == 1 && nullable:
			goType, zero = "sql.NullBool", "sql.NullBool{}"
		case nullable:
			goType, zero = "sql.NullInt64", "sql.NullInt64{}"
		default:
			goType, zero = "int8", "0"
		}
	case "smallint", "year":
		goType, zero = "int16", "0"
		if nullable {
			goType, zero = "sql.NullInt64", "sql.NullInt64{}"
		}
	case "mediumint", "int", "integer":
		goType, zero = Int32(ctx), "0"
		if nullable {
			goType, zero = "sql.NullInt64", "sql.NullInt64{}"
		}
	case "bigint":
		goType, zero = "int64", "0"
		if nullable {
			goType, zero = "sql.NullInt64", "sql.NullInt64{}"
		}
	case "float":
		goType, zero = "float32", "0.0"
		if nullable {
			goType, zero = "sql.NullFloat64", "sql.NullFloat64{}"
		}
	case "decimal", "double":
		goType, zero = "float64", "0.0"
		if nullable {
			goType, zero = "sql.NullFloat64", "sql.NullFloat64{}"
		}
	case "binary", "blob", "longblob", "mediumblob", "tinyblob", "varbinary":
		goType, zero = "[]byte", "nil"
	case "timestamp", "datetime", "date":
		goType, zero = "time.Time", "time.Time{}"
		if nullable {
			goType, zero = "sql.NullTime", "sql.NullTime{}"
		}
	case "time":
		// time is not supported by the MySQL driver. Can parse the string to time.Time in the user code.
		goType, zero = "string", `""`
		if nullable {
			goType, zero = "sql.NullString", "sql.NullString{}"
		}
	default:
		goType, zero = schemaGoType(ctx, typ, nullable)
	}
	// force []byte for SET('a',...)
	if setRE.MatchString(typ) {
		goType, zero = "[]byte", "nil"
	}
	// if unsigned ...
	if intRE.MatchString(goType) && unsigned && goType == Int32(ctx) {
		goType, zero = Uint32(ctx), "0"
	}
	return goType, zero, nil
}

// setRE is the regexp that matches MySQL SET() type definitions.
var setRE = regexp.MustCompile(`(?i)^set\([^)]*\)$`)

// MysqlEnumValues loads the enum values.
func MysqlEnumValues(ctx context.Context, db models.DB, schema string, enum string) ([]*models.EnumValue, error) {
	// load enum vals
	res, err := models.MysqlEnumValues(ctx, db, schema, enum)
	if err != nil {
		return nil, err
	}
	// process enum vals
	var values []*models.EnumValue
	for i, val := range strings.Split(res.EnumValues[1:len(res.EnumValues)-1], "','") {
		values = append(values, &models.EnumValue{
			EnumValue:  val,
			ConstValue: i + 1,
		})
	}
	return values, nil
}
