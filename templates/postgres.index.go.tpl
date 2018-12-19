{{- $short := (shortname .Type.Name "err" "sqlstr" "db" "q" "res" "XOLog" .Fields) -}}
{{- $shortRepo := (shortname .Type.RepoName "err" "sqlstr" "db" "q" "res" "XOLog" .Fields) -}}
{{- $table := (schema .Type.Table.TableName) -}}
// {{ .FuncName }} retrieves a row from '{{ $table }}' as a {{ .Type.Name }}.
//
// Generated from index '{{ .Index.IndexName }}'.
func ({{$shortRepo}} *{{ lowerfirst .Type.RepoName }}) {{ .FuncName }}(ctx context.Context, {{ goparamlist .Fields false true }}) ({{ if not .Index.IsUnique }}[]{{ end }}*entities.{{ .Type.Name }}, error) {
	var err error

	// sql query
	qb := sq.Select("*").From("`{{ $table }}`")
	{{- range $k, $v := .Fields }}
	    qb = qb.Where(sq.Eq{"`{{ colname .Col }}`": {{ goparam $v }}})
	{{- end }}

	query, args, err := qb.ToSql()
    if err != nil {
        return nil, errors.Wrap(err, "error in {{ .Type.RepoName }}")
    }

	// run query
{{- if .Index.IsUnique }}
	{{ $short }} := entities.{{ .Type.Name }}{}
	err = {{ $shortRepo }}.Db.Get(&{{ $short }}, query, args...)
    if err != nil {
        return nil, errors.Wrap(err, "error in {{ .Type.RepoName }}")
    }
{{- else }}
    var {{ $short }} []*entities.{{ .Type.Name }}
    err = {{ $shortRepo }}.Db.Select(&{{ $short }}, query, args...)
    if err != nil {
        return nil, errors.Wrap(err, "error in {{ .Type.RepoName }}")
    }
{{- end }}

{{- if .Index.IsUnique }}
	return &{{ $short }}, nil
{{- else }}
    return {{ $short }}, nil
{{- end }}
}

