{{- $short := (shortname .Type.Name "err" "sqlstr" "db" "q" "res" "XOLog" .QueryParams) -}}
{{- $queryComments := .QueryComments -}}
{{- if .Comment -}}
// {{ .Comment }}
{{- else -}}
// {{ .Name }} runs a custom query, returning results as {{ .Type.Name }}.
{{- end }}
func {{ .Name }} (db XODB{{ range .QueryParams }}, {{ .Name }} {{ .Type }}{{ end }}) ({{ if not .OnlyOne }}[]{{ end }}*{{ .Type.Name }}, error) {
	var err error

	// sql query
	{{ if .Interpolate }}var{{ else }}const{{ end }} sqlstr = {{ range $i, $l := .Query }}{{ if $i }} +{{ end }}{{ if (index $queryComments $i) }} // {{ index $queryComments $i }}{{ end }}{{ if $i }}
	{{end -}}`{{ $l }}`{{ end }}

	// run query
	XOLog(sqlstr{{ range .QueryParams }}{{ if not .Interpolate }}, {{ .Name }}{{ end }}{{ end }})
{{- if .OnlyOne }}
	var {{ $short }} {{ .Type.Name }}
	err = db.QueryRow(sqlstr{{ range .QueryParams }}, {{ .Name }}{{ end }}).Scan({{ fieldnames .Type.Fields (print "&" $short) }})
	if err != nil {
		return nil, err
	}

	return &{{ $short }}, nil
{{- else }}
	q, err := db.Query(sqlstr{{ range .QueryParams }}, {{ .Name }}{{ end }})
	if err != nil {
		return nil, err
	}
	defer q.Close()

	// load results
	res := []*{{ .Type.Name }}{}
	for q.Next() {
		{{ $short }} := {{ .Type.Name }}{}

		// scan
		err = q.Scan({{ fieldnames .Type.Fields (print "&" $short) }})
		if err != nil {
			return nil, err
		}

		res = append(res, &{{ $short }})
	}

	return res, nil
{{- end }}
}

