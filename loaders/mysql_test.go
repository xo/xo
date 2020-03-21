package loaders_test

import (
	"testing"

	"github.com/xo/xo/internal"
	"github.com/xo/xo/loaders"
)

func Test_MyParseType(t *testing.T) {
	tests := []struct {
		desc       string
		dt         string
		precision  int
		underlying string
		nilVal     string
		typ        string
		nullable   bool
	}{
		{
			desc:       "bit(1) parses",
			dt:         "bit(1)",
			precision:  1,
			underlying: "bool",
			nilVal:     "false",
			typ:        "bool",
		},
		{
			desc:       "bit(2) parses",
			dt:         "bit(2)",
			precision:  2,
			underlying: "uint8",
			nilVal:     "0",
			typ:        "uint8",
		},
		{
			desc:       "bit(8) parses",
			dt:         "bit(8)",
			precision:  8,
			underlying: "uint8",
			nilVal:     "0",
			typ:        "uint8",
		},
		{
			desc:       "bit(9) parses",
			dt:         "bit(9)",
			precision:  9,
			underlying: "uint16",
			nilVal:     "0",
			typ:        "uint16",
		},
		{
			desc:       "bit(16) parses",
			dt:         "bit(16)",
			precision:  16,
			underlying: "uint16",
			nilVal:     "0",
			typ:        "uint16",
		},
		{
			desc:       "bit(17) parses",
			dt:         "bit(17)",
			precision:  17,
			underlying: "uint32",
			nilVal:     "0",
			typ:        "uint32",
		},
		{
			desc:       "bit(32) parses",
			dt:         "bit(32)",
			precision:  32,
			underlying: "uint32",
			nilVal:     "0",
			typ:        "uint32",
		},
		{
			desc:       "bit(33) parses",
			dt:         "bit(33)",
			precision:  33,
			underlying: "uint64",
			nilVal:     "0",
			typ:        "uint64",
		},
		{
			desc:       "bit(64) parses",
			dt:         "bit(64)",
			precision:  64,
			underlying: "uint64",
			nilVal:     "0",
			typ:        "uint64",
		},
		{
			desc:       "nullable bit type with precision == 1 parses",
			dt:         "bit(1)",
			precision:  1,
			underlying: "*bool",
			nilVal:     "sql.NullBool{}",
			typ:        "sql.NullBool",
			nullable:   true,
		},
		{
			desc:       "nullable bit type with precision > 1 parses",
			dt:         "bit(64)",
			precision:  64,
			underlying: "*int64",
			nilVal:     "sql.NullInt64{}",
			typ:        "sql.NullInt64",
			nullable:   true,
		},
		{
			desc:       "tinyint with precision one parses into bool",
			dt:         "tinyint(1)",
			precision:  1,
			underlying: "bool",
			nilVal:     "false",
			typ:        "bool",
			nullable:   false,
		},
		{
			desc:       "nullable tinyint with precision one parses into bool",
			dt:         "tinyint(1)",
			precision:  1,
			underlying: "*bool",
			nilVal:     "sql.NullBool{}",
			typ:        "sql.NullBool",
			nullable:   true,
		},
		{
			desc:       "tinyint with greater than one precision parses into int8",
			dt:         "tinyint(4)",
			precision:  4,
			nilVal:     "0",
			underlying: "int8",
			typ:        "int8",
			nullable:   false,
		},
	}

	for i, tt := range tests {
		precision, nilVal, typ, underlying := loaders.MyParseType(&internal.ArgType{}, tt.dt, tt.nullable)
		if precision != tt.precision || nilVal != tt.nilVal || typ != tt.typ || underlying != tt.underlying {
			t.Fatalf("test #%d: %s\n\texp: %d, %s, %s, %s\n\tgot: %d, %s, %s, %s", i+1, tt.desc, tt.precision, tt.nilVal, tt.typ, tt.underlying, precision, nilVal, typ, underlying)
		}
	}
}
