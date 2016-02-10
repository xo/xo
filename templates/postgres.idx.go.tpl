{{ if .IsUnique }}
// {{ .Type }}By{{ if gt (len .Fields) 1 }}{{ .Name }}{{ else }}{{ range .Fields }}{{ .Field }}{{ end }}{{ end }} retrieves a row from {{ .TableSchema }}.{{ .TableName }} as a {{ .Type }}.
//
// Looks up using index {{ .IndexName }}.
func {{ .Type }}By{{ if gt (len .Fields) 1 }}{{ .Name }}{{ else }}{{ range .Fields }}{{ .Field }}{{ end }}{{ end }}(db XODB{{ goparamlist .Fields true }}) (*{{ .Type }}, error) {
	var err error

	// sql query
	const sqlstr = `SELECT ` +
		`{{ colnames .Table.Fields }} ` +
		`FROM {{ .TableSchema }}.{{ .TableName }} ` +
		`WHERE {{ range $i, $f := .Fields }}{{ if $i }} AND {{ end }}{{ $f.ColumnName }} = ${{ inc $i }}{{ end }}`

	// run query
	{{ shortname .Type }} := {{ .Type }}{
	{{- if .Table.PrimaryKeyField }}
		_exists: true,
	{{ end -}}
	}
	err = db.QueryRow(sqlstr{{ goparamlist .Fields false }}).Scan({{ fieldnames .Table.Fields (print "&" (shortname .Type)) }})
	if err != nil {
		return nil, err
	}

	return &{{ shortname .Type }}, nil
}
{{ else }}
// {{ .Plural }}By{{ .Name }} retrieves rows from {{ .TableSchema }}.{{ .TableName }}, each as a {{ .Type }}.
//
// Looks up using index {{ .IndexName }}.
func {{ .Plural }}By{{ .Name }}(db XODB{{ goparamlist .Fields true }}) ([]*{{ .Type }}, error) {
	var err error

	// sql query
	const sqlstr = `SELECT ` +
		`{{ colnames .Table.Fields }} ` +
		`FROM {{ .TableSchema }}.{{ .TableName }} ` +
		`WHERE {{ range $i, $f := .Fields }}{{ if $i }} AND {{ end }}{{ $f.ColumnName }} = ${{ inc $i }}{{ end }}`

	// run query
	q, err := db.Query(sqlstr{{ goparamlist .Fields false }})
	if err != nil {
		return nil, err
	}
	defer q.Close()

	// load results
	res := []*{{ .Type }}{}
	for q.Next() {
		{{ shortname .Type }} := {{ .Type }}{
		{{- if .Table.PrimaryKeyField }}
			_exists: true,
		{{ end -}}
		}

		// scan
		err = q.Scan({{ fieldnames .Table.Fields (print "&" (shortname .Type)) }})
		if err != nil {
			return nil, err
		}

		res = append(res, &{{ shortname .Type }})
	}

	return res, nil
}
{{ end }}

