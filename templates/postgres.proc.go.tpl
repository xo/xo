// {{ .Name }} calls the stored procedure '{{ .TableSchema }}.{{ .ProcName }}({{ .ProcParameterTypes }}) {{ .ProcReturnType }}' on db.
func {{ .Name }}(db XODB{{ goparamlist .Parameters true }}) ({{ retype .ReturnType }}, error) {
	var err error

	// sql query
	const sqlstr = `SELECT {{ .TableSchema }}.{{ .ProcName }}({{ colvals .Parameters "" }})`

	// run query
	var ret {{ retype .ReturnType }}
	err = db.QueryRow(sqlstr{{ goparamlist .Parameters false }}).Scan(&ret)
	if err != nil {
		return {{ reniltype .NilReturnType }}, err
	}

	return ret, nil
}

