{{- $enum := .Data -}}
{{- $type := $enum.Name -}}
{{- $short := (shortname $type "val" "buf" "ok" "v") -}}
// {{ $type }} is the '{{ $enum.Enum.EnumName }}' enum type from schema '{{ schema }}'.
type {{ $type }} uint16

const (
{{ range $enum.Values -}}
	// {{ $type }}{{ .Name }} is the '{{ .Val.EnumValue }}' {{ $type }}.
	{{ $type }}{{ .Name }} = {{ $type }}({{ .Val.ConstValue }})
{{ end -}}
)

// String satisfies the fmt.Stringer interface.
func ({{ $short }} {{ $type }}) String() string {
	switch {{ $short }} {
{{ range $enum.Values -}}
	case {{ $type }}{{ .Name }}:
		return "{{ .Val.EnumValue }}"
{{ end -}}
	}
	return fmt.Sprintf("{{ $type }}(%d)", {{ $short }})
}

// MarshalText marshals {{ $type }} into text.
func ({{ $short }} {{ $type }}) MarshalText() ([]byte, error) {
	return []byte({{ $short }}.String()), nil
}

// UnmarshalText unmarshals {{ $type }} from text.
func ({{ $short }} *{{ $type }}) UnmarshalText(buf []byte) error {
	switch s :=string(buf); s {
{{ range $enum.Values -}}
	case "{{ .Val.EnumValue }}":
		*{{ $short }} = {{ $type }}{{ .Name }}
{{ end -}}
	default:
		return ErrInvalid{{ $type }}(s)
	}
	return nil
}

// Value satisfies the driver.Valuer interface.
func ({{ $short }} {{ $type }}) Value() (driver.Value, error) {
	return {{ $short }}.String(), nil
}

// Scan satisfies the sql.Scanner interface.
func ({{ $short }} *{{ $type }}) Scan(v interface{}) error {
	if buf, ok := v.([]byte); ok {
		return {{ $short }}.UnmarshalText(buf)
	}
	return ErrInvalid{{ $type }}(fmt.Sprintf("%T", v))
}

// ErrInvalid{{ $type }} is the invalid {{ $type }} error.
type ErrInvalid{{ $type }} string

// Error satisfies the error interface.
func (err ErrInvalid{{ $type }}) Error() string {
	return fmt.Sprintf("invalid {{ $type }} (%s)", string(err))
}

