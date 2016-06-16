{{ $notVoid := (ne .Proc.ReturnType "void") }}
{{- if ne .Proc.ReturnType "trigger" -}}
// {{ .Name }} calls the stored procedure '{{ schema .Schema .Proc.ProcName }}({{ .ProcParams }}) {{ .Proc.ReturnType }}' on db.
func {{ .Name }}(db XODB{{ goparamlist .Params true }}) ({{ if $notVoid }}{{ retype .Return.Type }}, {{ end }}error) {
	var err error

	// sql query
	const sqlstr = `SELECT {{ schema .Schema .Proc.ProcName }}({{ colvals .Params }})`

	// run query
{{- if $notVoid }}
	var ret {{ retype .Return.Type }}
	XOLog(sqlstr{{ goparamlist .Params false }})
	err = db.QueryRow(sqlstr{{ goparamlist .Params false }}).Scan(&ret)
	if err != nil {
		return {{ reniltype .Return.NilType }}, err
	}

	return ret, nil
{{- else }}
	XOLog(sqlstr)
	return db.Exec(sqlstr)
{{- end }}
}
{{- end }}

