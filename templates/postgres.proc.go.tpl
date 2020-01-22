{{- $notVoid := (ne .Proc.ReturnType "void") -}}
{{- $tupleReturn := false -}}
{{- if ge (len .Proc.ReturnType) 5 }}
        {{$tupleReturn = (eq (slice .Proc.ReturnType 0 5) "SETOF")}}
{{- end}}
{{- $proc := (schema .Schema .Proc.ProcName) -}}
{{- if ne .Proc.ReturnType "trigger" -}}
// {{ .Name }} calls the stored procedure '{{ $proc }}({{ .ProcParams }}) {{ .Proc.ReturnType }}' on db.
func {{ .Name }}({{- if $tupleReturn}}db *sqlx.DB{{- else}}db XODB{{- end }}{{ goparamlist .Params true true }}) ({{ if $notVoid }}{{ retype .Return.Type }}, {{ end }}error) {
	var err error

	// sql query
	const sqlstr = `SELECT {{ if $tupleReturn}}* from {{ end }}{{ $proc }}({{ colvals .Params }})`

	// run query
{{- if $tupleReturn }}
     //tuple return
{{- end }}
{{- if $notVoid }}
	var ret {{ retype .Return.Type }}
	XOLog(sqlstr{{ goparamlist .Params true false }})
	{{- if $tupleReturn}}
	err = db.Select(&ret, sqlstr)
	{{- else }}
	err = db.QueryRow(sqlstr{{ goparamlist .Params true false }}).Scan(&ret)
	{{- end }}
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

