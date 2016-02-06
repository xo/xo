// {{ .Type }} is the '{{ .EnumType }}' enum type.
type {{ .Type }} uint16

const (
{{- range .Values }}
	// {{ .Value }}{{ .Type }} is the {{ .EnumType }} for '{{ .EnumValue }}'.
	{{ .Value }}{{ .Type }} = {{ .Type }}({{ .ConstValue }})
{{ end -}}
)

// String returns the string value of the {{ .Type }}.
func ({{ shortname .Type }} {{ .Type }}) String() string {
	var enumVal string

	switch {{ shortname .Type }} {
{{- range .Values }}
	case {{ .Value }}{{ .Type }}:
		enumVal = "{{ .EnumValue }}"
{{ end -}}
	}

	return enumVal
}

// MarshalText marshals {{ .Type }} into text.
func ({{ shortname .Type }} {{ .Type }}) MarshalText() ([]byte, error) {
	return []byte({{ shortname .Type }}.String()), nil
}

// UnmarshalText unmarshals {{ .Type }} from text.
func ({{ shortname .Type }} *{{ .Type }}) UnmarshalText(text []byte) error {
	switch string(text)	{
{{- range .Values }}
	case "{{ .EnumValue }}":
		*{{ shortname .Type }} = {{ .Value }}{{ .Type }}
{{ end }}

	default:
		return errors.New("invalid {{ .Type }}")
	}

	return nil
}

