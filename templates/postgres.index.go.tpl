// {{ .FuncName }} retrieves a row from '{{ schema .Schema .Type.Table.TableName }}' as a {{ .Type.Name }}.
//
// Generated from index '{{ .Index.IndexName }}'.
func {{ .FuncName }}(db XODB{{ goparamlist .Fields true }}) ({{ if not .Index.IsUnique }}[]{{ end }}*{{ .Type.Name }}, error) {
	var err error

	// sql query
	const sqlstr = `SELECT ` +
		`{{ colnames .Type.Fields }} ` +
		`FROM {{ schema .Schema .Type.Table.TableName }} ` +
		`WHERE {{ colnamesquery .Fields " AND " }}`

	// run query
	XOLog(sqlstr{{ goparamlist .Fields false }})
{{- if .Index.IsUnique }}
	{{ shortname .Type.Name }} := {{ .Type.Name }}{
	{{- if .Type.PrimaryKey }}
		_exists: true,
	{{ end -}}
	}

	err = db.QueryRow(sqlstr{{ goparamlist .Fields false }}).Scan({{ fieldnames .Type.Fields (print "&" (shortname .Type.Name)) }})
	if err != nil {
		return nil, err
	}

	return &{{ shortname .Type.Name }}, nil
{{- else }}
	q, err := db.Query(sqlstr{{ goparamlist .Fields false }})
	if err != nil {
		return nil, err
	}
	defer q.Close()

	// load results
	res := []*{{ .Type.Name }}{}
	for q.Next() {
		{{ shortname .Type.Name }} := {{ .Type.Name }}{
		{{- if .Type.PrimaryKey }}
			_exists: true,
		{{ end -}}
		}

		// scan
		err = q.Scan({{ fieldnames .Type.Fields (print "&" (shortname .Type.Name)) }})
		if err != nil {
			return nil, err
		}

		res = append(res, &{{ shortname .Type.Name }})
	}

	return res, nil
{{- end }}
}

