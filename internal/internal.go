// Package internal contains internal code for xo.
package internal

import (
	"reflect"
)

// Symbols are extracted (generated) symbols from the types package.
//
//go:generate ./gen.sh
var Symbols map[string]map[string]reflect.Value = make(map[string]map[string]reflect.Value)
