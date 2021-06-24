{{- $k := .Data -}}
// {{ func_name_context $k }} returns the {{ $k.RefType.Name }} associated with the {{ $k.Type.Name }}'s {{ $k.Field.Name }} ({{ $k.Field.Col.ColumnName }}).
//
// Generated from foreign key '{{ $k.ForeignKey.ForeignKeyName }}'.
{{ recv_context $k.Type $k }} {
	return {{ foreign_key_context $k }}
}
{{- if context_both }}

// {{ func_name $k }} returns the {{ $k.RefType.Name }} associated with the {{ $k.Type.Name }}'s {{ $k.Field.Name }} ({{ $k.Field.Col.ColumnName }}).
//
// Generated from foreign key '{{ $k.ForeignKey.ForeignKeyName }}'.
{{ recv $k.Type $k }} {
	return {{ foreign_key $k }}
}
{{- end }}

