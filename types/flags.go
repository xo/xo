package types

import (
	"github.com/alecthomas/kingpin"
)

// Flag is a option flag.
type Flag struct {
	ContextKey  ContextKey
	Desc        string
	PlaceHolder string
	Default     string
	Short       rune
	Value       interface{}
	Enums       []string
}

// FlagSet is a set of flags for a driver.
type FlagSet struct {
	Type string
	Name string
	Flag Flag
}

// Add adds the flag to the cmd.
func (flag FlagSet) Add(cmd *kingpin.CmdClause, flags map[ContextKey]interface{}) {
	f := cmd.Flag(flag.Type+"-"+flag.Name, flag.Flag.Desc).
		PlaceHolder(flag.Flag.PlaceHolder).
		Short(flag.Flag.Short).
		Default(flag.Flag.Default)
	switch flag.Flag.Value.(type) {
	case bool:
		flags[flag.Flag.ContextKey] = newBool(f, flags[flag.Flag.ContextKey])
	case string:
		flags[flag.Flag.ContextKey] = newString(f, flags[flag.Flag.ContextKey], flag.Flag.Enums)
	case []string:
		flags[flag.Flag.ContextKey] = newStrings(f, flags[flag.Flag.ContextKey], flag.Flag.Enums)
	}
}

// newBool creates a new bool when v is nil, otherwise it converts v and returns.
func newBool(f *kingpin.FlagClause, v interface{}) *bool {
	if v == nil {
		b := false
		f.BoolVar(&b)
		return &b
	}
	b := v.(*bool)
	f.BoolVar(b)
	return b
}

// newString creates a new string when v is nil, otherwise it converts v and returns.
func newString(f *kingpin.FlagClause, v interface{}, enums []string) *string {
	if v == nil {
		s := ""
		if len(enums) != 0 {
			f.EnumVar(&s, enums...)
		} else {
			f.StringVar(&s)
		}
		return &s
	}
	s := v.(*string)
	if len(enums) != 0 {
		f.EnumVar(s, enums...)
	} else {
		f.StringVar(s)
	}
	return s
}

// newStrings creates a new string when v is nil, otherwise it converts v and returns.
func newStrings(f *kingpin.FlagClause, v interface{}, enums []string) *[]string {
	switch {
	case v == nil && enums == nil:
		var s []string
		f.StringsVar(&s)
		return &s
	case v != nil && enums == nil:
		s := v.(*[]string)
		f.StringsVar(s)
		return s
	case v == nil:
		var s []string
		f.EnumsVar(&s, enums...)
		return &s
	}
	s := v.(*[]string)
	f.EnumsVar(s, enums...)
	return s
}
