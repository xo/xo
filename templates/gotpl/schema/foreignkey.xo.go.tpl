{{- $fkey := .Data -}}
{{- $short := (shortname $fkey.Type.Name) -}}
// {{ $fkey.Name }} returns the {{ $fkey.RefType.Name }} associated with the {{ $fkey.Type.Name }}'s {{ $fkey.Field.Name }} ({{ $fkey.Field.Col.ColumnName }}).
//
// Generated from foreign key '{{ $fkey.ForeignKey.ForeignKeyName }}'.
func ({{ $short }} *{{ $fkey.Type.Name }}) {{ $fkey.Name }}(ctx context.Context, db DB) (*{{ $fkey.RefType.Name }}, error) {
	return {{ $fkey.RefType.Name }}By{{ $fkey.RefField.Name }}(ctx, db, {{ convext $short $fkey.Field $fkey.RefField }})
}

