{{ $QueryComments := .QueryComments }}
{{- if .Comment -}}
// {{ .Comment }}
{{- else -}}
// {{ .Name }} runs a custom query, returning results as {{ .Type.Name }}.
{{- end }}
func {{ .Name }} (db XODB{{ range .QueryParams }}, {{ .Name }} {{ .Type }}{{ end }}) ({{ if not .OnlyOne }}[]{{ end }}*{{ .Type.Name }}, error) {
	var err error

	// sql query
	{{ if .Interpolate }}var{{ else }}const{{ end }} sqlstr = {{ range $i, $l := .Query }}{{ if $i }} +{{ end }}{{ if (index $QueryComments $i) }} // {{ index $QueryComments $i }}{{ end }}{{ if $i }}
	{{end -}}`{{ $l }}`{{ end }}

	// run query
	XOLog(sqlstr{{ range .QueryParams }}{{ if not .Interpolate }}, {{ .Name }}{{ end }}{{ end }})
{{- if .OnlyOne }}
	var {{ shortname .Type.Name }} {{ .Type.Name }}
	err = db.QueryRow(sqlstr{{ range .QueryParams }}, {{ .Name }}{{ end }}).Scan({{ fieldnames .Type.Fields (print "&" (shortname .Type.Name)) }})
	if err != nil {
		return nil, err
	}

	return &{{ shortname .Type.Name }}, nil
{{- else }}
	q, err := db.Query(sqlstr{{ range .QueryParams }}, {{ .Name }}{{ end }})
	if err != nil {
		return nil, err
	}
	defer q.Close()

	// load results
	res := []*{{ .Type.Name }}{}
	for q.Next() {
		{{ shortname .Type.Name}} := {{ .Type.Name }}{}

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

