{{- $e := .Data -}}
// {{ $e.Name }} is the '{{ $e.Enum.EnumName }}' enum type from schema '{{ schema }}'.
type {{ $e.Name }} uint16

const (
{{ range $e.Values -}}
	// {{ $e.Name }}{{ .Name }} is the '{{ .Val.EnumValue }}' {{ $e.Name }}.
	{{ $e.Name }}{{ .Name }} {{ $e.Name }} = {{ .Val.ConstValue }}
{{ end -}}
)

// String satisfies the fmt.Stringer interface.
func ({{ short $e.Name }} {{ $e.Name }}) String() string {
	switch {{ short $e.Name }} {
{{ range $e.Values -}}
	case {{ $e.Name }}{{ .Name }}:
		return "{{ .Val.EnumValue }}"
{{ end -}}
	}
	return fmt.Sprintf("{{ $e.Name }}(%d)", {{ short $e.Name }})
}

// MarshalText marshals {{ $e.Name }} into text.
func ({{ short $e.Name }} {{ $e.Name }}) MarshalText() ([]byte, error) {
	return []byte({{ short $e.Name }}.String()), nil
}

// UnmarshalText unmarshals {{ $e.Name }} from text.
func ({{ short $e.Name }} *{{ $e.Name }}) UnmarshalText(buf []byte) error {
	switch s := string(buf); s {
{{ range $e.Values -}}
	case "{{ .Val.EnumValue }}":
		*{{ short $e.Name }} = {{ $e.Name }}{{ .Name }}
{{ end -}}
	default:
		return ErrInvalid{{ $e.Name }}(s)
	}
	return nil
}

// Value satisfies the driver.Valuer interface.
func ({{ short $e.Name }} {{ $e.Name }}) Value() (driver.Value, error) {
	return {{ short $e.Name }}.String(), nil
}

// Scan satisfies the sql.Scanner interface.
func ({{ short $e.Name }} *{{ $e.Name }}) Scan(v interface{}) error {
	if buf, ok := v.([]byte); ok {
		return {{ short $e.Name }}.UnmarshalText(buf)
	}
	return ErrInvalid{{ $e.Name }}(fmt.Sprintf("%T", v))
}

// ErrInvalid{{ $e.Name }} is the invalid {{ $e.Name }} error.
type ErrInvalid{{ $e.Name }} string

// Error satisfies the error interface.
func (err ErrInvalid{{ $e.Name }}) Error() string {
	return fmt.Sprintf("invalid {{ $e.Name }} (%s)", string(err))
}

