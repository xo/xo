{{ $QueryComments := .QueryComments }}
{{- if .Comment -}}
// {{ .Comment }}
{{- else -}}
// {{ .Name }} runs a custom query, returning results as {{ .Type }}.
{{- end }}
func {{ .Name }} (db XODB{{ range $i, $p := .Parameters }}, {{ index $p 0 }} {{ index $p 1 }}{{ end }}) ({{ if not .OnlyOne }}[]{{ end }}*{{ .Type }}, error) {
	var err error

	// sql query
	const sqlstr = {{ range $i, $l := .Query }}{{ if $i }} +{{ end }}{{ if (index $QueryComments $i) }} // {{ index $QueryComments $i }}{{ end }}{{ if $i }}
	{{end -}}`{{ $l }}`{{ end }}

	// run query
{{- if .OnlyOne }}
	var ret {{ .Type }}
	err = db.QueryRow(sqlstr{{ range .Parameters }}, {{ index . 0 }}{{ end }}).Scan({{ fieldnames .Table.Fields "" "&ret" }})
	if err != nil {
		return nil, err
	}

	return &ret, nil
{{- else }}
	q, err := db.Query(sqlstr{{ range .Parameters }}, {{ index . 0 }}{{ end }})
	if err != nil {
		return nil, err
	}
	defer q.Close()

	// load results
	res := []*{{ .Type }}{}
	for q.Next() {
		{{ shortname .Type}} := {{ .Type }}{}

		// scan
		err = q.Scan({{ fieldnames .Table.Fields "" (print "&" (shortname .Type)) }})
		if err != nil {
			return nil, err
		}

		res = append(res, &{{ shortname .Type }})
	}

	return res, nil
{{- end }}
}

