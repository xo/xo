{{- $p := .Data -}}
// {{ func_name_context $p }} calls the stored {{ $p.Kind }} '{{ $p.Signature }}' on db.
{{ func_context $p }} {
{{- if and (driver "mysql") (eq $p.Kind "procedure") (not $p.Void) }}
	// At the moment, the Go MySQL driver does not support stored procedures
	// with out parameters
	return {{ range $p.Returns }}{{ zero .Zero }}, {{ end }}fmt.Errorf("unsupported")
{{- else }}
	// call {{ schema $p.SQLName }}
	{{ sqlstr "proc" $p }}
	// run
{{- if not $p.Void }}
{{- range $p.Returns }}
	var {{ .GoName }} {{ type .Type }}
{{- end }}
	logf(sqlstr, {{ params $p.Params false }})
{{- if and (driver "sqlserver" "oracle") (eq $p.Kind "procedure")}}
	if _, err := {{ db_named "Exec" $p }}; err != nil {
{{- else }}
	if err := {{ db "QueryRow" $p }}.Scan({{ names "&" $p.Returns }}); err != nil {
{{- end }}
		return {{ range $p.Returns }}{{ zero .Zero }}, {{ end }}logerror(err)
	}
	return {{ range $p.Returns }}{{ .GoName }}, {{ end }}nil
{{- else -}}
	logf(sqlstr)
{{- if driver "sqlserver" "oracle" }}
	if _, err := {{ db_named "Exec" $p }}; err != nil {
{{- else }}
	if _, err := {{ db "Exec" $p }}; err != nil {
{{- end }}
		return logerror(err)
	}
	return nil
{{- end }}
{{- end }}
}

{{ if context_both -}}
// {{ func_name $p }} calls the {{ $p.Kind }} '{{ $p.Signature }}' on db.
{{ func $p }} {
	return {{ func_name_context $p }}({{ names_all "" "context.Background()" "db" $p.Params }})
}
{{- end }}
