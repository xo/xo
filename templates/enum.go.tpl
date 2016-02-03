// {{ .Type }} is the '{{ .TypeNative }}' enum type.
type {{ .Type }} uint16

const (
{{- range .Values }}
	// {{ .EnumType }}{{ .Type }} is the {{ .Type }} for '{{ .Value }}'.
	{{ .EnumType }}{{ .Type }} {{ .Type }} = {{ .ConstValue }}
{{ end -}}
)

// String returns the string value of the {{ .Type }}.
func (c {{ .Type }}) String() string {
	var s string

	switch c {
{{- range .Values }}
	case {{ .EnumType }}{{ .Type }}:
		s = "{{ .Value }}"
{{ end -}}
	}

	return s
}

// MarshalText marshals {{ .Type }} into text.
func (c {{ .Type }}) MarshalText() ([]byte, error) {
	return []byte(c.String()), nil
}

// UnmarshalText unmarshals {{ .Type }} from text.
func (c *{{ .Type }}) UnmarshalText(text []byte) error {
	switch string(text)	{
{{- range .Values }}
	case "{{ .Value }}":
		*c = {{ .EnumType }}{{ .Type }}
{{ end -}}
	default:
		return errors.New("invalid {{ .Type }}")
	}

	return nil
}
