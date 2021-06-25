{{- $p := .Data -}}
{{- if ne $p.Return.GoName "trigger" -}}
// {{ func_name_context $p }} calls the stored procedure '{{ $p.Signature }}' on db.
{{ func_context $p }} {
	// call {{ schema $p.SQLName }}
	{{ sqlstr "proc" $p }}
	// run
{{- if ne $p.Return.Type "void" }}
	var {{ short $p.Return.Type }} {{ type $p.Return.Type }}
	logf(sqlstr, {{ params $p.Params false }})
{{- if driver "sqlserver" }}
	if _, err := {{ db_named "Exec" $p }}; err != nil {
{{- else }}
	if err := {{ db "QueryRow" $p }}.Scan(&{{ short $p.Return.Type}}); err != nil {
{{- end }}	
		return {{ zero $p.Return.Zero }}, logerror(err)
	}
	return {{ short $p.Return.Type }}, nil
{{- else -}}
	logf(sqlstr)
	if _, err := {{ db "Exec" }}
		return logerror(err)
	}
	return nil
{{- end -}}
}

{{ if context_both -}}
// {{ func_name $p }} calls the stored procedure '{{ $p.Signature }}' on db.
{{ func $p }} {
	return {{ func_name_context $p }}({{ names_all "" "context.Background()" "db" $p.Params }})
}
{{- end }}
{{- end }}

