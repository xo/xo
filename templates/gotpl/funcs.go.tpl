// Package funcs holds custom template funcs for xo.
package funcs

import (
	"context"
	"text/template"
	"time"
)

// Init returns a func map for use by the xo.
func Init(ctx context.Context) (template.FuncMap, error) {
	return template.FuncMap{
		"date": date,
	}, nil
}

// date returns the current date for the specified format.
func date(s string) string {
	if v, ok := timeConstMap[s]; ok {
		s = v
	}
	return time.Now().Format(s)
}

// timeConstMap is the time const name to value map.
var timeConstMap = map[string]string{
	"ANSIC":       time.ANSIC,
	"UnixDate":    time.UnixDate,
	"RubyDate":    time.RubyDate,
	"RFC822":      time.RFC822,
	"RFC822Z":     time.RFC822Z,
	"RFC850":      time.RFC850,
	"RFC1123":     time.RFC1123,
	"RFC1123Z":    time.RFC1123Z,
	"RFC3339":     time.RFC3339,
	"RFC3339Nano": time.RFC3339Nano,
	"Kitchen":     time.Kitchen,
	"Stamp":       time.Stamp,
	"StampMilli":  time.StampMilli,
	"StampMicro":  time.StampMicro,
	"StampNano":   time.StampNano,
}
