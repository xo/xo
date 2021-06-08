{{- $proc := .Data -}}
{{- $notVoid := (ne $proc.Proc.ReturnType "void") -}}
{{- $name := (schema $proc.Proc.ProcName) -}}
{{- if ne $proc.Proc.ReturnType "trigger" -}}
// {{ $proc.Name }} calls the stored procedure '{{ $name }}({{ $proc.ProcParams }}) {{ $proc.Proc.ReturnType }}' on db.
func {{ $proc.Name }}(ctx context.Context, db DB{{ paramlist $proc.Params true true }}) ({{ if $notVoid }}{{ retype $proc.Return.Type }}, {{ end }}error) {
	// call {{ $name }}
	const sqlstr = `SELECT {{ $name }}({{ colvals $proc.Params }})`
	// run
{{ if $notVoid -}}
	var ret {{ retype $proc.Return.Type }}
	logf(sqlstr{{ paramlist $proc.Params true false }})
	if err := db.QueryRowContext(ctx, sqlstr{{ paramlist $proc.Params true false }}).Scan(&ret); err != nil {
		return {{ reniltype $proc.Return.Zero }}, logerror(err)
	}
	return ret, nil
{{- else -}}
	logf(sqlstr)
	if _, err := db.ExecContext(ctx, sqlstr); err != nil {
		return logerror(err)
	}
	return nil
{{- end }}
}
{{- end }}

