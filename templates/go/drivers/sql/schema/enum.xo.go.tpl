{{- $e := .Data -}}
// {{ $e.GoName }} is the '{{ $e.SQLName }}' enum type from schema '{{ schema }}'.
type {{ $e.GoName }} uint16

// {{ $e.GoName }} values.
const (
{{ range $e.Values -}}
	// {{ $e.GoName }}{{ .GoName }} is the '{{ .SQLName }}' {{ $e.SQLName }}.
	{{ $e.GoName }}{{ .GoName }} {{ $e.GoName }} = {{ .ConstValue }}
{{ end -}}
)

// String satisfies the fmt.Stringer interface.
func ({{ short $e.GoName }} {{ $e.GoName }}) String() string {
	switch {{ short $e.GoName }} {
{{ range $e.Values -}}
	case {{ $e.GoName }}{{ .GoName }}:
		return "{{ .SQLName }}"
{{ end -}}
	}
	return fmt.Sprintf("{{ $e.GoName }}(%d)", {{ short $e.GoName }})
}

// MarshalText marshals {{ $e.GoName }} into text.
func ({{ short $e.GoName }} {{ $e.GoName }}) MarshalText() ([]byte, error) {
	return []byte({{ short $e.GoName }}.String()), nil
}

// UnmarshalText unmarshals {{ $e.GoName }} from text.
func ({{ short $e.GoName }} *{{ $e.GoName }}) UnmarshalText(buf []byte) error {
	switch str := string(buf); str {
{{ range $e.Values -}}
	case "{{ .SQLName }}":
		*{{ short $e.GoName }} = {{ $e.GoName }}{{ .GoName }}
{{ end -}}
	default:
		return ErrInvalid{{ $e.GoName }}(str)
	}
	return nil
}

// Value satisfies the driver.Valuer interface.
func ({{ short $e.GoName }} {{ $e.GoName }}) Value() (driver.Value, error) {
	return {{ short $e.GoName }}.String(), nil
}

// Scan satisfies the sql.Scanner interface.
func ({{ short $e.GoName }} *{{ $e.GoName }}) Scan(v interface{}) error {
	if buf, ok := v.([]byte); ok {
		return {{ short $e.GoName }}.UnmarshalText(buf)
	}
	return ErrInvalid{{ $e.GoName }}(fmt.Sprintf("%T", v))
}

{{ $nullName := (printf "%s%s" "Null" $e.GoName) -}}
{{- $nullShort := (short $nullName) -}}
// {{ $nullName }} represents a null '{{ $e.SQLName }}' enum for schema '{{ schema }}'.
type {{ $nullName }} struct {
	{{ $e.GoName }} {{ $e.GoName }}
	// Valid is true if {{ $e.GoName }} is not null.
	Valid bool
}

// Value satisfies the driver.Valuer interface.
func ({{ $nullShort }} {{ $nullName }}) Value() (driver.Value, error) {
	if !{{ $nullShort }}.Valid {
		return nil, nil
	}
	return {{ $nullShort }}.{{ $e.GoName }}.Value()
}

// Scan satisfies the sql.Scanner interface.
func ({{ $nullShort }} *{{ $nullName }}) Scan(v interface{}) error {
	if v == nil {
		{{ $nullShort }}.{{ $e.GoName }}, {{ $nullShort }}.Valid = 0, false
		return nil
	}
	err := {{ $nullShort }}.{{ $e.GoName }}.Scan(v)
	{{ $nullShort }}.Valid = err == nil
	return err
}

// ErrInvalid{{ $e.GoName }} is the invalid {{ $e.GoName }} error.
type ErrInvalid{{ $e.GoName }} string

// Error satisfies the error interface.
func (err ErrInvalid{{ $e.GoName }}) Error() string {
	return fmt.Sprintf("invalid {{ $e.GoName }}(%s)", string(err))
}

