{{- $type := .Name -}}
// {{ $type }} is the '{{ .Enum.EnumName }}' enum type from schema '{{ .Schema  }}'.
type {{ $type }} uint16

const (
{{- range .Values }}
	// {{ .Name }}{{ $type }} is the '{{ .Val.EnumValue }}' {{ $type }}.
	{{ .Name }}{{ $type }} = {{ $type }}({{ .Val.ConstValue }})
{{ end -}}
)

// String returns the string value of the {{ $type }}.
func ({{ shortname $type }} {{ $type }}) String() string {
	var enumVal string

	switch {{ shortname $type }} {
{{- range .Values }}
	case {{ .Name }}{{ $type }}:
		enumVal = "{{ .Val.EnumValue }}"
{{ end -}}
	}

	return enumVal
}

// MarshalText marshals {{ $type }} into text.
func ({{ shortname $type }} {{ $type }}) MarshalText() ([]byte, error) {
	return []byte({{ shortname $type }}.String()), nil
}

// UnmarshalText unmarshals {{ $type }} from text.
func ({{ shortname $type }} *{{ $type }}) UnmarshalText(text []byte) error {
	switch string(text)	{
{{- range .Values }}
	case "{{ .Val.EnumValue }}":
		*{{ shortname $type }} = {{ .Name }}{{ $type }}
{{ end }}

	default:
		return errors.New("invalid {{ $type }}")
	}

	return nil
}

// Value satisfies the sql/driver.Valuer interface for {{ $type }}.
func ({{ shortname $type }} {{ $type }}) Value() (driver.Value, error) {
	return {{ shortname $type }}.String(), nil
}

// Scan satisfies the database/sql.Scanner interface for {{ $type }}.
func ({{ shortname $type }} *{{ $type }}) Scan(src interface{}) error {
	buf, ok := src.([]byte)
	if !ok {
	   return errors.New("invalid {{ $type }}")
	}

	return {{ shortname $type }}.UnmarshalText(buf)
}

