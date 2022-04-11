package types

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gobwas/glob"
	"github.com/spf13/cobra"
)

// Flag is a option flag.
type Flag struct {
	ContextKey ContextKey
	Type       string
	Desc       string
	Default    string
	Short      string
	Enums      []string
	Aliases    []string
	Hidden     bool
}

// FlagSet is a set of option flags.
type FlagSet struct {
	Type string
	Name string
	Flag Flag
}

// Add adds the flag to the cmd.
func (flag FlagSet) Add(cmd *cobra.Command, values map[ContextKey]*Value) error {
	switch flag.Flag.Type {
	case "bool", "int", "string", "[]string", "glob":
	default:
		return fmt.Errorf("unknown flag type %s", flag.Flag.Type)
	}
	// create value
	if _, ok := values[flag.Flag.ContextKey]; !ok {
		values[flag.Flag.ContextKey] = NewValue(flag.Flag.Type, flag.Flag.Default, flag.Flag.Desc, flag.Flag.Enums...)
	}
	flags, name, desc := cmd.Flags(), flag.Type+"-"+flag.Name, values[flag.Flag.ContextKey].Desc()
	// force pflag flag as bool
	noOptDefValue := ""
	if flag.Flag.Type == "bool" {
		noOptDefValue = "true"
	}
	// add flag
	switch {
	case flag.Flag.Short == "":
		flags.Var(values[flag.Flag.ContextKey], name, desc)
	case flags.ShorthandLookup(flag.Flag.Short) != nil:
		desc += fmt.Sprintf(" (short flag -%s also available)", flag.Flag.Short)
		flags.Var(values[flag.Flag.ContextKey], name, desc)
	default:
		flags.VarP(values[flag.Flag.ContextKey], name, flag.Flag.Short, desc)
	}
	// add aliases
	for _, alias := range flag.Flag.Aliases {
		if flags.Lookup(alias) != nil {
			continue
		}
		flags.Var(values[flag.Flag.ContextKey], alias, desc)
		f := flags.Lookup(alias)
		f.Hidden, f.NoOptDefVal = true, noOptDefValue
	}
	// copy short, hidden, bool
	f := flags.Lookup(name)
	f.Hidden, f.NoOptDefVal = flag.Flag.Hidden, noOptDefValue
	return nil
}

// Value wraps a flag value.
type Value struct {
	typ   string
	def   string
	desc  string
	enums []string
	set   bool
	v     interface{}
}

// NewValue creates a new flag value.
func NewValue(typ, def, desc string, enums ...string) *Value {
	var z interface{}
	switch typ {
	case "bool":
		var b bool
		z = b
	case "int":
		var i int
		z = i
	case "string":
		var s string
		z = s
	case "[]string":
		var s []string
		z = s
	case "glob":
		var v []glob.Glob
		z = v
	}
	v := &Value{
		typ:   typ,
		def:   def,
		desc:  desc,
		enums: enums,
		v:     z,
	}
	if v.def != "" {
		if err := v.Set(v.def); err != nil {
			panic(err)
		}
		v.set = false
	}
	return v
}

// String satisfies the pflag.Value interface.
func (v *Value) String() string {
	return v.def
}

// Desc returns the usage description for the flag value.
func (v *Value) Desc() string {
	if v.enums != nil {
		return v.desc + " <" + strings.Join(v.enums, "|") + ">"
	}
	return v.desc
}

// Set satisfies the pflag.Value interface.
func (v *Value) Set(s string) error {
	v.set = true
	if v.enums != nil {
		if !contains(v.enums, s) {
			return fmt.Errorf("invalid value %q", s)
		}
	}
	switch v.typ {
	case "bool":
		b, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		v.v = b
	case "int":
		i, err := strconv.Atoi(s)
		if err != nil {
			return err
		}
		v.v = i
	case "string":
		v.v = s
	case "[]string":
		v.v = append(v.v.([]string), strings.Split(s, ",")...)
	case "glob":
		g, err := glob.Compile(s)
		if err != nil {
			return err
		}
		v.v = append(v.v.([]glob.Glob), g)
	}
	return nil
}

// Interface returns the value.
func (v *Value) Interface() interface{} {
	if v.v == nil {
		panic("v should not be nil!")
	}
	return v.v
}

// AsBool returns the value as a bool.
func (v *Value) AsBool() bool {
	b, _ := v.v.(bool)
	return b
}

// AsInt returns the value as a int.
func (v *Value) AsInt() int {
	i, _ := v.v.(int)
	return i
}

// AsString returns the value as a string.
func (v *Value) AsString() string {
	s, _ := v.v.(string)
	return s
}

// AsStringSlice returns the value as a string slice.
func (v *Value) AsStringSlice() []string {
	z, _ := v.v.([]string)
	return z
}

// AsGlob returns the value as a glob slice.
func (v *Value) AsGlob() []glob.Glob {
	z, _ := v.v.([]glob.Glob)
	return z
}

// Type satisfies the pflag.Value interface.
func (v *Value) Type() string {
	if v.typ == "[]string" {
		return "string"
	}
	return v.typ
}

// contains determines if v contains str.
func contains(v []string, str string) bool {
	for _, s := range v {
		if s == str {
			return true
		}
	}
	return false
}
