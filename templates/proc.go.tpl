// {{ .FuncName }} calls the stored procedure '{{ .Schema }}.{{ .Name }}({{ .ParameterTypes }}) {{ .ReturnType }}' on db.
func {{ .FuncName }}(db *sql.DB{{ range $i, $t := .GoParameterTypes }}, v{{ $i }} {{ $t }}{{ end }}) ({{ .GoReturnType }}, error) {
	var err error

	// sql query
	const sqlstr = `SELECT {{ .Schema }}.{{ .Name }}({{ range $i, $t := .GoParameterTypes }}{{ if $i }}, {{ end }}${{inc $i }}{{ end }})`

	// run query
	var ret {{ .GoReturnType }}
	err = db.QueryRow(sqlstr{{ range $i, $t := .GoParameterTypes }}, v{{ $i }}{{ end }}).Scan(&ret)
	if err != nil {
		return {{ .GoNilReturnType }}, err
	}

	return ret, nil
}
