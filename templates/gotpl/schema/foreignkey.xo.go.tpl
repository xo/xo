{{- $k := .Data -}}
// {{ func_name_context $k }} returns the {{ $k.RefTable }} associated with the {{ $k.Table.GoName }}'s {{ $k.Field.GoName }} ({{ $k.Field.SQLName }}).
//
// Generated from foreign key '{{ $k.SQLName }}'.
{{ recv_context $k.Table $k }} {
	return {{ foreign_key_context $k }}
}
{{- if context_both }}

// {{ func_name $k }} returns the {{ $k.RefTable }} associated with the {{ $k.Table }}'s {{ $k.Field.GoName }} ({{ $k.Field.SQLName }}).
//
// Generated from foreign key '{{ $k.SQLName }}'.
{{ recv $k.Table $k }} {
	return {{ foreign_key $k }}
}
{{- end }}

