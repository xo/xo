// Package internal contains internal code for xo.
package internal

import (
	"reflect"
)

// Symbols are extracted (generated) symbols from the types package.
//
//go:generate yaegi extract github.com/xo/xo/loader
//go:generate yaegi extract github.com/xo/xo/types
//
//go:generate yaegi extract os/exec
//
//go:generate yaegi extract github.com/gobwas/glob
//go:generate yaegi extract github.com/goccy/go-yaml
//go:generate yaegi extract github.com/kenshaw/inflector
//go:generate yaegi extract github.com/kenshaw/snaker
//go:generate yaegi extract github.com/Masterminds/sprig/v3
//go:generate yaegi extract github.com/yookoala/realpath
//go:generate yaegi extract golang.org/x/tools/imports
//go:generate yaegi extract mvdan.cc/gofumpt/format
var Symbols map[string]map[string]reflect.Value = make(map[string]map[string]reflect.Value)
