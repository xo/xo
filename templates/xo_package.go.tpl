// Package {{ .Package }} contains the types for schema '{{ schema .Schema }}'.
package {{ .Package }}

// GENERATED BY XO. DO NOT EDIT.

import (
	"database/sql"
	"database/sql/driver"
	"encoding/csv"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"
)

