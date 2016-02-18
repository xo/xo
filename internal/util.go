package internal

import (
	"bytes"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"
)

// QueryParameter is an extracted query parameter from a query.
type QueryParameter struct {
	Name        string
	Type        string
	Interpolate bool
}

// ParseQuery takes the query in args and looks for strings in the form of
// "%%<name> <type>[,<option>,...]%%", replacing them with the supplied mask.
// mask can contain "%d" to indicate current position. The modified query is
// returned, and the slice of extracted QueryParameter's.
func (a *ArgType) ParseQuery(mask string, interpol bool) (string, []QueryParameter) {
	dl := a.QueryParamDelimiter

	// create the regexp for the delimiter
	placeholderRE := regexp.MustCompile(
		dl + `[^` + dl[:1] + `]+` + dl,
	)

	// grab matches from query string
	matches := placeholderRE.FindAllStringIndex(a.Query, -1)

	// return vals and placeholders
	str := ""
	params := []QueryParameter{}
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
		param := QueryParameter{
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

// lenRE is a regular expression that matches precision (length) definitions in
// a database.
var LenRE = regexp.MustCompile(`\([0-9]+\)$`)

// intRE matches Go int types.
var IntRE = regexp.MustCompile(`^int[0-9]*$`)

// TBuf is to hold compiled template strings.
type TBuf struct {
	Type TemplateType
	Name string
	Buf  *bytes.Buffer
}

// TBufSlice is a sortable slice of TBuf.
type TBufSlice []TBuf

func (t TBufSlice) Len() int {
	return len(t)
}

func (t TBufSlice) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t TBufSlice) Less(i, j int) bool {
	if t[i].Type == XO {
		return false
	} else if t[j].Type == XO {
		return true
	}

	if t[i].Name == t[j].Name {
		return t[i].Type <= t[j].Type
	}

	return strings.Compare(t[i].Name, t[j].Name) <= 0
}
