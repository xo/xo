{{- $proc := .Data -}}
{{- $notVoid := (ne $proc.Proc.ReturnType "void") -}}
{{- $name := (schema $proc.Proc.ProcName) -}}
{{- if ne $proc.Proc.ReturnType "trigger" -}}
// {{ $proc.Name }}{{ if context_both }}Context{{ end }} calls the stored procedure '{{ $name }}({{ $proc.ProcParams }}) {{ $proc.Proc.ReturnType }}' on db.
func {{ $proc.Name }}{{ if context_both }}Context{{ end }}({{ if context }}ctx context.Context, {{ end }}db DB{{ paramlist $proc.Params true true }}) ({{ if $notVoid }}{{ retype $proc.Return.Type }}, {{ end }}error) {
	// call {{ $name }}
	const sqlstr = `SELECT {{ $name }}({{ colvals $proc.Params }})`
	// run
{{ if $notVoid -}}
	var ret {{ retype $proc.Return.Type }}
	logf(sqlstr{{ paramlist $proc.Params true false }})
	if err := db.QueryRow{{ if context }}Context{{ end }}({{ if context }}ctx, {{ end }}sqlstr{{ paramlist $proc.Params true false }}).Scan(&ret); err != nil {
		return {{ reniltype $proc.Return.Zero }}, logerror(err)
	}
	return ret, nil
{{- else -}}
	logf(sqlstr)
	if _, err := db.Exec{{ if context }}Context{{ end }}({{ if context }}ctx, {{ end }}sqlstr); err != nil {
		return logerror(err)
	}
	return nil
{{- end -}}
}

{{ if context_both -}}
// {{ $proc.Name }} calls the stored procedure '{{ $name }}({{ $proc.ProcParams }}) {{ $proc.Proc.ReturnType }}' on db.
func {{ $proc.Name }}(db DB{{ paramlist $proc.Params true true }}) ({{ if $notVoid }}{{ retype $proc.Return.Type }}, {{ end }}error) {
	return {{ $proc.Name }}Context(context.Background(), db{{ paramlist $proc.Params true false }})
}
{{- end }}
{{- end }}

