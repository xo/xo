// {{ .Type }} is the '{{ .EnumType }}' enum type.
type {{ .Type }} uint16

const (
{{- range .Values }}
	// {{ .Value }}{{ .Type }} is the {{ .EnumType }} for '{{ .EnumValue }}'.
	{{ .Value }}{{ .Type }} = {{ .Type }}({{ .ConstValue }})
{{ end -}}
)

// String returns the string value of the {{ .Type }}.
func (t {{ .Type }}) String() string {
	var s string

	switch t {
{{- range .Values }}
	case {{ .Value }}{{ .Type }}:
		s = "{{ .EnumValue }}"
{{ end -}}
	}

	return s
}

// MarshalText marshals {{ .Type }} into text.
func (t {{ .Type }}) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

// UnmarshalText unmarshals {{ .Type }} from text.
func (t *{{ .Type }}) UnmarshalText(text []byte) error {
	switch string(text)	{
{{- range .Values }}
	case "{{ .EnumValue }}":
		*t = {{ .Value }}{{ .Type }}
{{ end -}}
	default:
		return errors.New("invalid {{ .Type }}")
	}

	return nil
}

