// Code generated by 'yaegi extract github.com/gobwas/glob'. DO NOT EDIT.

package internal

import (
	"github.com/gobwas/glob"
	"reflect"
)

func init() {
	Symbols["github.com/gobwas/glob/glob"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"Compile":     reflect.ValueOf(glob.Compile),
		"MustCompile": reflect.ValueOf(glob.MustCompile),
		"QuoteMeta":   reflect.ValueOf(glob.QuoteMeta),

		// type definitions
		"Glob": reflect.ValueOf((*glob.Glob)(nil)),

		// interface wrapper definitions
		"_Glob": reflect.ValueOf((*_github_com_gobwas_glob_Glob)(nil)),
	}
}

// _github_com_gobwas_glob_Glob is an interface wrapper for Glob type
type _github_com_gobwas_glob_Glob struct {
	IValue interface{}
	WMatch func(a0 string) bool
}

func (W _github_com_gobwas_glob_Glob) Match(a0 string) bool {
	return W.WMatch(a0)
}