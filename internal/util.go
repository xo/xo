package internal

import (
	"bytes"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/serenize/snaker"
)

// ParseQuery takes the query in args and looks for strings in the form of
// "%%<name> <type>[,<option>,...]%%", replacing them with the supplied mask.
// mask can contain "%d" to indicate current position. The modified query is
// returned, and the slice of extracted QueryParam's.
func (a *ArgType) ParseQuery(mask string, interpol bool) (string, []*QueryParam) {
	dl := a.QueryParamDelimiter

	// create the regexp for the delimiter
	placeholderRE := regexp.MustCompile(
		dl + `[^` + dl[:1] + `]+` + dl,
	)

	// grab matches from query string
	matches := placeholderRE.FindAllStringIndex(a.Query, -1)

	// return vals and placeholders
	str := ""
	params := []*QueryParam{}
	i := 1
	last := 0

	// loop over matches, extracting each placeholder and splitting to name/type
	for _, m := range matches {
		// generate place holder value
		pstr := mask
		if strings.Contains(mask, "%d") {
			pstr = fmt.Sprintf(mask, i)
		}

		// extract parameter info
		paramStr := a.Query[m[0]+len(dl) : m[1]-len(dl)]
		p := strings.SplitN(paramStr, " ", 2)
		param := &QueryParam{
			Name: p[0],
			Type: p[1],
		}

		// parse parameter options if present
		if strings.Contains(param.Type, ",") {
			opts := strings.Split(param.Type, ",")
			param.Type = opts[0]
			for _, opt := range opts[1:] {
				switch opt {
				case "interpolate":
					if !a.QueryInterpolate {
						panic("query interpolate is not enabled")
					}
					param.Interpolate = true

				default:
					panic(fmt.Errorf("unknown option encountered on query parameter '%s'", paramStr))
				}
			}
		}

		// add to string
		str = str + a.Query[last:m[0]]
		if interpol && param.Interpolate {
			// handle interpolation case
			xstr := `fmt.Sprintf("%v", ` + param.Name + `)`
			if param.Type == "string" {
				xstr = param.Name
			}
			str = str + "` + " + xstr + " + `"
		} else {
			str = str + pstr
		}

		params = append(params, param)
		last = m[1]
		i++
	}

	// add part of query remains
	str = str + a.Query[last:]

	return str, params
}

// IntRE matches Go int types.
var IntRE = regexp.MustCompile(`^int(32|64)?$`)

// PrecScaleRE is the regexp that matches "(precision[,scale])" definitions in a
// database.
var PrecScaleRE = regexp.MustCompile(`\(([0-9]+)(\s*,[0-9]+)?\)$`)

// ParsePrecision extracts (precision[,scale]) strings from a data type and
// returns the data type without the string.
func (a *ArgType) ParsePrecision(dt string) (string, int, int) {
	var err error

	precision := -1
	scale := -1

	m := PrecScaleRE.FindStringSubmatchIndex(dt)
	if m != nil {
		// extract precision
		precision, err = strconv.Atoi(dt[m[2]:m[3]])
		if err != nil {
			panic("could not convert precision")
		}

		// extract scale
		if m[4] != -1 {
			scale, err = strconv.Atoi(dt[m[4]+1 : m[5]])
			if err != nil {
				panic("could not convert scale")
			}
		}

		// change dt
		dt = dt[:m[0]] + dt[m[1]:]
	}

	return dt, precision, scale
}

// letters for GenRandomID
var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

// GenRandomID generates a 8 character random string.
func GenRandomID() string {
	rand.Seed(time.Now().UTC().UnixNano())
	b := make([]rune, 8)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// TBuf is to hold the executed templates.
type TBuf struct {
	TemplateType TemplateType
	Name         string
	Subname      string
	Buf          *bytes.Buffer
}

// TBufSlice is a slice of TBuf compatible with sort.Interface.
type TBufSlice []TBuf

func (t TBufSlice) Len() int {
	return len(t)
}

func (t TBufSlice) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t TBufSlice) Less(i, j int) bool {
	if t[i].TemplateType < t[j].TemplateType {
		return true
	} else if t[j].TemplateType < t[i].TemplateType {
		return false
	}

	if strings.Compare(t[i].Name, t[j].Name) < 0 {
		return true
	} else if strings.Compare(t[j].Name, t[i].Name) < 0 {
		return false
	}

	return strings.Compare(t[i].Subname, t[j].Subname) < 0
}

var uRE = regexp.MustCompile(`_+`)

// SnakeToCamel provides a safer version of snaker.SnakeToCamel
func SnakeToCamel(s string) string {
	// remove leading/trailing underscores
	s = strings.TrimLeft(s, "_")
	s = strings.TrimRight(s, "_")

	// fix 2 or more __
	s = uRE.ReplaceAllString(s, "_")

	return snaker.SnakeToCamel(s)
}
