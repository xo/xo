{{- $query := .Data -}}
{{- $short := (shortname $query.Type.Name "err" "sqlstr" "db" "rows" "res" "logf" $query.Params) -}}
{{- $queryComments := $query.Comments -}}
{{- if $query.Comment -}}
// {{ $query.Comment }}
{{- else -}}
// {{ $query.Name }}{{ if context_both }}Context{{ end }} runs a custom query{{ if not $query.Flat }}, returning results as {{ $query.Type.Name }}{{ end }}.
{{- end }}
func {{ $query.Name }}{{ if context_both }}Context{{ end }}({{ if context }}ctx context.Context, {{ end }}db DB{{ range $query.Params }}, {{ .Name }} {{ .Type }}{{ end }}) ({{ if not $query.One }}[]{{ end }}{{ if $query.Flat }}{{ range $query.Type.Fields }}{{ retype .Type }}, {{ end }}{{ else }}*{{ $query.Type.Name }}, {{ end }}error) {
	// query
	{{ if $query.Interpolate }}var{{ else }}const{{ end }} sqlstr = {{ range $i, $l := $query.Query }}{{ if $i }} +{{ end }}{{ if (index $queryComments $i) }} // {{ index $queryComments $i }}{{ end }}{{ if $i }}
	{{end -}}`{{ $l }}`{{ end }}
	// run
	logf(sqlstr{{ range $query.Params }}{{ if not .Interpolate }}, {{ .Name }}{{ end }}{{ end }})
{{ if $query.One -}}
{{- if $query.Flat -}}
{{ range $query.Type.Fields -}}
	var {{ .Name }} {{ retype .Type }}
{{- end }}
	if err := db.QueryRow{{ if context }}Context{{ end }}({{ if context }}ctx, {{ end }}sqlstr{{ range $query.Params }}{{ if not .Interpolate }}, {{ .Name }}{{ end }}{{ end }}).Scan({{ fieldnames $query.Type.Fields "" }}); err != nil {
		return {{ range $query.Type.Fields }}{{ reniltype .Zero }}, {{ end }}logerror(err)
	}
	return {{ range $query.Type.Fields }}{{ .Name }}, {{ end }}nil
{{- else -}}
	var {{ $short }} {{ $query.Type.Name }}
	if err := db.QueryRow{{ if context }}Context{{ end }}({{ if context }}ctx, {{ end }}sqlstr{{ range $query.Params }}{{ if not .Interpolate }}, {{ .Name }}{{ end }}{{ end }}).Scan({{ fieldnames $query.Type.Fields (print "&" $short) }}); err != nil {
		return nil, logerror(err)
	}
	return &{{ $short }}, nil
{{- end }}
{{- else -}}
	rows, err := db.Query{{ if context }}Context{{ end }}({{ if context }}ctx, {{ end }}sqlstr{{ range $query.Params }}{{ if not .Interpolate }}, {{ .Name }}{{ end }}{{ end }})
	if err != nil {
		return nil, logerror(err)
	}
	defer rows.Close()
	// load results
	var res []*{{ $query.Type.Name }}
	for rows.Next() {
		var {{ $short }} {{ $query.Type.Name }}
		// scan
		if err := rows.Scan({{ fieldnames $query.Type.Fields (print "&" $short) }}); err != nil {
			return nil, logerror(err)
		}
		res = append(res, &{{ $short }})
	}
	if err := rows.Err(); err != nil {
		return nil, logerror(err)
	}
	return res, nil
{{- end }}
}
{{- if context_both }}

{{ if $query.Comment -}}
// {{ $query.Comment }}
{{- else -}}
// {{ $query.Name }} runs a custom query{{ if not $query.Flat }}, returning results as {{ $query.Type.Name }}{{ end }}.
{{- end }}
func {{ $query.Name }}(db DB{{ range $query.Params }}, {{ .Name }} {{ .Type }}{{ end }}) ({{ if not $query.One }}[]{{ end }}{{ if $query.Flat }}{{ range $query.Type.Fields }}{{ retype .Type }}, {{ end }}{{ else }}*{{ $query.Type.Name }}, {{ end }}error) {
	return {{ $query.Name }}Context(context.Background(), db{{ range $query.Params }}{{ if not .Interpolate }}, {{ .Name }}{{ end }}{{ end }})
}
{{- end }}

