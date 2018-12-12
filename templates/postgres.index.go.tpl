{{- $short := (shortname .Type.Name "err" "sqlstr" "db" "q" "res" "XOLog" .Fields) -}}
{{- $shortRepo := (shortname .Type.RepoName "err" "sqlstr" "db" "q" "res" "XOLog" .Fields) -}}
{{- $table := (schema .Schema .Type.Table.TableName) -}}
// {{ .FuncName }} retrieves a row from '{{ $table }}' as a {{ .Type.Name }}.
//
// Generated from index '{{ .Index.IndexName }}'.
func ({{$shortRepo}} *{{.Type.RepoName}}) {{ .FuncName }}({{ goparamlist .Fields false true }}) ({{ if not .Index.IsUnique }}[]{{ end }}*entities.{{ .Type.Name }}, error) {
	var err error

	// sql query
	qb := squirrel.Select("{{ $table }}")
	{{- range $k, $v := .Fields }}
	    qb = qb.Where(squirrel.Eq{"{{ colname .Col }}": {{ goparam $v }}})
	{{- end }}

	query, args, err := qb.ToSql()
    if err != nil {
        return nil, err
    }

	// run query
{{- if .Index.IsUnique }}
	{{ $short }} := entities.{{ .Type.Name }}{}
{{- else }}
    var {{ $short }} []*entities.{{ .Type.Name }}
{{- end }}

	err = {{ $shortRepo }}.db.Select(&{{ $short }}, query, args...)
	if err != nil {
		return nil, err
	}

{{- if .Index.IsUnique }}
	return &{{ $short }}, nil
{{- else }}
    return {{ $short }}, nil
{{- end }}
}

