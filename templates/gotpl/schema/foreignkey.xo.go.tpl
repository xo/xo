{{- $fkey := .Data -}}
{{- $short := (shortname $fkey.Type.Name) -}}
// {{ $fkey.Name }}{{ if context_both }}Context{{ end }} returns the {{ $fkey.RefType.Name }} associated with the {{ $fkey.Type.Name }}'s {{ $fkey.Field.Name }} ({{ $fkey.Field.Col.ColumnName }}).
//
// Generated from foreign key '{{ $fkey.ForeignKey.ForeignKeyName }}'.
func ({{ $short }} *{{ $fkey.Type.Name }}) {{ $fkey.Name }}{{ if context_both }}Context{{ end }}({{ if context }}ctx context.Context, {{ end }}db DB) (*{{ $fkey.RefType.Name }}, error) {
	return {{ $fkey.RefType.Name }}By{{ $fkey.RefField.Name }}{{ if context_both }}Context{{ end }}({{ if context }}ctx, {{ end }}db, {{ convext $short $fkey.Field $fkey.RefField }})
}
{{- if context_both }}

// {{ $fkey.Name }} returns the {{ $fkey.RefType.Name }} associated with the {{ $fkey.Type.Name }}'s {{ $fkey.Field.Name }} ({{ $fkey.Field.Col.ColumnName }}).
//
// Generated from foreign key '{{ $fkey.ForeignKey.ForeignKeyName }}'.
func ({{ $short }} *{{ $fkey.Type.Name }}) {{ $fkey.Name }}(db DB) (*{{ $fkey.RefType.Name }}, error) {
	return {{ $fkey.RefType.Name }}By{{ $fkey.RefField.Name }}Context(context.Background(), db, {{ convext $short $fkey.Field $fkey.RefField }})
}
{{- end }}

