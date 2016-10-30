{{- $notVoid := (ne .Proc.ReturnType "void") -}}
{{- $proc := (schema .Schema .Proc.ProcName) -}}
{{- if ne .Proc.ReturnType "trigger" -}}
// {{ .Name }} calls the stored procedure '{{ $proc }}({{ .ProcParams }}) {{ .Proc.ReturnType }}' on db.
func {{ .Name }}(db XODB{{ goparamlist .Params true true }}) ({{ if $notVoid }}{{ retype .Return.Type }}, {{ end }}error) {
	var err error

	// sql query
	const sqlstr = `SELECT {{ $proc }}({{ colvals .Params }})`

	// run query
{{- if $notVoid }}
	var ret {{ retype .Return.Type }}
	XOLog(sqlstr{{ goparamlist .Params true false }})
	err = db.QueryRow(sqlstr{{ goparamlist .Params true false }}).Scan(&ret)
	if err != nil {
		return {{ reniltype .Return.NilType }}, err
	}

	return ret, nil
{{- else }}
	XOLog(sqlstr)
	_, err = db.Exec(sqlstr)
	return err
{{- end }}
}
{{- end }}

