// {{ .Name }} calls the stored procedure '{{ schema .Schema .Proc.ProcName }}({{ .ProcParams }}) {{ .Proc.ReturnType }}' on db.
func {{ .Name }}(db XODB{{ goparamlist .Params true }}) ({{ retype .Return.Type }}, error) {
	var err error

	// sql query
	const sqlstr = `SELECT {{ schema .Schema .Proc.ProcName }}({{ colvals .Params }})`

	// run query
	var ret {{ retype .Return.Type }}
	XOLog(sqlstr{{ goparamlist .Params false }})
	err = db.QueryRow(sqlstr{{ goparamlist .Params false }}).Scan(&ret)
	if err != nil {
		return {{ reniltype .Return.NilType }}, err
	}

	return ret, nil
}

