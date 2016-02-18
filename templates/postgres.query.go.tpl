{{ $QueryComments := .QueryComments }}
{{- if .Comment -}}
// {{ .Comment }}
{{- else -}}
// {{ .Name }} runs a custom query, returning results as {{ .Type }}.
{{- end }}
func {{ .Name }} (db XODB{{ range .Parameters }}, {{ .Name }} {{ .Type }}{{ end }}) ({{ if not .OnlyOne }}[]{{ end }}*{{ .Type }}, error) {
	var err error

	// sql query
	{{ if .Interpolate }}var{{ else }}const{{ end }} sqlstr = {{ range $i, $l := .Query }}{{ if $i }} +{{ end }}{{ if (index $QueryComments $i) }} // {{ index $QueryComments $i }}{{ end }}{{ if $i }}
	{{end -}}`{{ $l }}`{{ end }}

	// run query
	XOLog(sqlstr{{ range .Parameters }}{{ if not .Interpolate }}, {{ .Name }}{{ end }}{{ end }})
{{- if .OnlyOne }}
	var {{ shortname .Type }} {{ .Type }}
	err = db.QueryRow(sqlstr{{ range .Parameters }}, {{ .Name }}{{ end }}).Scan({{ fieldnames .Table.Fields (print "&" (shortname .Type)) }})
	if err != nil {
		return nil, err
	}

	return &{{ shortname .Type }}, nil
{{- else }}
	q, err := db.Query(sqlstr{{ range .Parameters }}, {{ .Name }}{{ end }})
	if err != nil {
		return nil, err
	}
	defer q.Close()

	// load results
	res := []*{{ .Type }}{}
	for q.Next() {
		{{ shortname .Type}} := {{ .Type }}{}

		// scan
		err = q.Scan({{ fieldnames .Table.Fields (print "&" (shortname .Type)) }})
		if err != nil {
			return nil, err
		}

		res = append(res, &{{ shortname .Type }})
	}

	return res, nil
{{- end }}
}

