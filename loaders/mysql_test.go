package loaders_test

import (
	"testing"

	"github.com/knq/xo/internal"
	"github.com/knq/xo/loaders"
)

func Test_MyParseType(t *testing.T) {
	tests := []struct {
		desc      string
		dt        string
		precision int
		nilVal    string
		typ       string
		nullable  bool
	}{
		{
			desc:      "bit(1) parses",
			dt:        "bit(1)",
			precision: 1,
			nilVal:    "false",
			typ:       "bool",
		},
		{
			desc:      "bit(2) parses",
			dt:        "bit(2)",
			precision: 2,
			nilVal:    "0",
			typ:       "uint8",
		},
		{
			desc:      "bit(8) parses",
			dt:        "bit(8)",
			precision: 8,
			nilVal:    "0",
			typ:       "uint8",
		},
		{
			desc:      "bit(9) parses",
			dt:        "bit(9)",
			precision: 9,
			nilVal:    "0",
			typ:       "uint16",
		},
		{
			desc:      "bit(16) parses",
			dt:        "bit(16)",
			precision: 16,
			nilVal:    "0",
			typ:       "uint16",
		},
		{
			desc:      "bit(17) parses",
			dt:        "bit(17)",
			precision: 17,
			nilVal:    "0",
			typ:       "uint32",
		},
		{
			desc:      "bit(32) parses",
			dt:        "bit(32)",
			precision: 32,
			nilVal:    "0",
			typ:       "uint32",
		},
		{
			desc:      "bit(33) parses",
			dt:        "bit(33)",
			precision: 33,
			nilVal:    "0",
			typ:       "uint64",
		},
		{
			desc:      "bit(64) parses",
			dt:        "bit(64)",
			precision: 64,
			nilVal:    "0",
			typ:       "uint64",
		},
		{
			desc:      "nullable bit type with precision == 1 parses",
			dt:        "bit(1)",
			precision: 1,
			nilVal:    "sql.NullBool{}",
			typ:       "sql.NullBool",
			nullable:  true,
		},
		{
			desc:      "nullable bit type with precision > 1 parses",
			dt:        "bit(64)",
			precision: 64,
			nilVal:    "sql.NullInt64{}",
			typ:       "sql.NullInt64",
			nullable:  true,
		},
		{
			desc:      "tinyint with precision one parses into bool",
			dt:        "tinyint(1)",
			precision: 1,
			nilVal:    "false",
			typ:       "bool",
			nullable:  false,
		},
		{
			desc:      "nullable tinyint with precision one parses into bool",
			dt:        "tinyint(1)",
			precision: 1,
			nilVal:    "sql.NullBool{}",
			typ:       "sql.NullBool",
			nullable:  true,
		},
		{
			desc:      "tinyint with greater than one precision parses into int8",
			dt:        "tinyint(4)",
			precision: 4,
			nilVal:    "0",
			typ:       "int8",
			nullable:  false,
		},
	}

	for i, tt := range tests {
		precision, nilVal, typ := loaders.MyParseType(&internal.ArgType{}, tt.dt, tt.nullable)
		if precision != tt.precision || nilVal != tt.nilVal || typ != tt.typ {
			t.Fatalf("test #%d: %s\n\texp: %d, %s, %s\n\tgot: %d, %s, %s", i+1, tt.desc, tt.precision, tt.nilVal, tt.typ, precision, nilVal, typ)
		}
	}
}
