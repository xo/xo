package loader

import (
	"context"
	"testing"

	xo "github.com/xo/xo/types"
)

func TestMysqlGoType(t *testing.T) {
	tests := []struct {
		name     string
		typ      string
		nullable bool
		goType   string
		zero     string
		prec     int
		scale    int
	}{
		{
			name:   "bit(1) parses",
			typ:    "bit(1)",
			goType: "bool",
			zero:   "false",
			prec:   1,
		},
		{
			name:   "bit(2) parses",
			typ:    "bit(2)",
			goType: "uint8",
			zero:   "0",
			prec:   2,
		},
		{
			name:   "bit(8) parses",
			typ:    "bit(8)",
			goType: "uint8",
			zero:   "0",
			prec:   8,
		},
		{
			name:   "bit(9) parses",
			typ:    "bit(9)",
			goType: "uint16",
			zero:   "0",
			prec:   9,
		},
		{
			name:   "bit(16) parses",
			typ:    "bit(16)",
			goType: "uint16",
			zero:   "0",
			prec:   16,
		},
		{
			name:   "bit(17) parses",
			typ:    "bit(17)",
			goType: "uint32",
			zero:   "0",
			prec:   17,
		},
		{
			name:   "bit(32) parses",
			typ:    "bit(32)",
			goType: "uint32",
			zero:   "0",
			prec:   32,
		},
		{
			name:   "bit(33) parses",
			typ:    "bit(33)",
			goType: "uint64",
			zero:   "0",
			prec:   33,
		},
		{
			name:   "bit(64) parses",
			typ:    "bit(64)",
			goType: "uint64",
			zero:   "0",
			prec:   64,
		},
		{
			name:     "nullable bit type with precision == 1 parses",
			typ:      "bit(1)",
			nullable: true,
			goType:   "sql.NullBool",
			zero:     "sql.NullBool{}",
			prec:     1,
		},
		{
			name:     "nullable bit type with precision > 1 parses",
			typ:      "bit(64)",
			nullable: true,
			goType:   "sql.NullInt64",
			zero:     "sql.NullInt64{}",
			prec:     64,
		},
		{
			name:   "tinyint with precision 1 parses into bool",
			typ:    "tinyint(1)",
			goType: "bool",
			zero:   "false",
			prec:   1,
		},
		{
			name:     "nullable tinyint with precision 1 parses into sql.NullBool",
			typ:      "tinyint(1)",
			nullable: true,
			goType:   "sql.NullBool",
			zero:     "sql.NullBool{}",
			prec:     1,
		},
		{
			name:   "tinyint with greater than precision 1 parses into int8",
			typ:    "tinyint(4)",
			goType: "int8",
			zero:   "0",
			prec:   4,
		},
		{
			name:     "nullable smallint with precision 5 parses into sql.NullInt64",
			typ:      "smallint(5)",
			nullable: true,
			goType:   "sql.NullInt64",
			zero:     "sql.NullInt64{}",
			prec:     5,
		},
		{
			name:   "smallint with precision 5 parses into int16",
			typ:    "smallint(5)",
			goType: "int16",
			zero:   "0",
			prec:   5,
		},
		{
			name:     "nullable unsigned smallint with precision 5 parses into sql.NullInt64",
			typ:      "smallint(5) unsigned",
			nullable: true,
			goType:   "sql.NullInt64",
			zero:     "sql.NullInt64{}",
			prec:     5,
		},
		{
			name:   "nullable unsigned smallint with precision 5 parses into uint16",
			typ:    "smallint(5) unsigned",
			goType: "uint16",
			zero:   "0",
			prec:   5,
		},
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, xo.DriverKey, "mysql")
	ctx = context.WithValue(ctx, xo.SchemaKey, "mysql")
	for i, test := range tests {
		d, err := xo.ParseType(ctx, test.typ)
		if err != nil {
			t.Fatalf("test %d (%s) %q (nullable: %t) expected no error, got: %v", i, test.name, test.typ, test.nullable, err)
		}
		if d.Prec != test.prec {
			t.Errorf("test %d (%s) %q (nullable: %t) expected d.Prec = %d, got: %d", i, test.name, test.typ, test.nullable, test.prec, d.Prec)
		}
		if d.Scale != test.scale {
			t.Errorf("test %d (%s) %q (nullable: %t) expected d.Scale = %d, got: %d", i, test.name, test.typ, test.nullable, test.scale, d.Scale)
		}
		if d.IsArray {
			t.Errorf("test %d (%s) %q (nullable: %t) expected d.IsArray = false", i, test.name, test.typ, test.nullable)
		}
		d.Nullable = test.nullable
		goType, zero, err := MysqlGoType(ctx, d)
		if err != nil {
			t.Fatalf("test %d (%s) %q MysqlGoType(%#v) expected no error, got: %v", i, test.name, test.typ, d, err)
		}
		if goType != test.goType {
			t.Errorf("test %d (%s) %q (nullable: %t) expected goType = %q, got: %q", i, test.name, test.typ, test.nullable, test.goType, goType)
		}
		if zero != test.zero {
			t.Errorf("test %d (%s) %q (nullable: %t) expected zero = %q, got: %q", i, test.name, test.typ, test.nullable, test.zero, zero)
		}
	}
}
