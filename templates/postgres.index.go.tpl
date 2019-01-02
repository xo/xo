{{- $short := (shortname .Type.Name "err" "sqlstr" "db" "q" "res" "XOLog" .Fields) -}}
{{- $shortRepo := (shortname .Type.RepoName "err" "sqlstr" "db" "q" "res" "XOLog" .Fields) -}}
{{- $table := (schema .Type.Table.TableName) -}}
// {{ .FuncName }} retrieves a row from '{{ $table }}' as a {{ .Type.Name }}.
//
// Generated from index '{{ .Index.IndexName }}'.
{{- if .Index.IsUnique }}
func ({{$shortRepo}} *{{ .Type.RepoName }}) {{ .FuncName }}(ctx context.Context, {{ goparamlist .Fields false true }}, filter *entities.{{ .Type.Name }}Filter) (entities.{{ .Type.Name }}, error) {
	var err error

	var db = {{ $shortRepo }}.Db
    tx := db_manager.GetTransactionContext(ctx)
    if tx != nil {
        db = tx
    }

	// sql query
    qb := {{$shortRepo}}.findAll{{ .Type.Name }}BaseQuery(ctx, filter, "*")
    {{- range $k, $v := .Fields }}
        qb = qb.Where(sq.Eq{"`{{ colname .Col }}`": {{ goparam $v }}})
    {{- end }}

	query, args, err := qb.ToSql()
    if err != nil {
        return entities.{{ .Type.Name }}{}, errors.Wrap(err, "error in {{ .Type.RepoName }}")
    }

	// run query
	{{ $short }} := entities.{{ .Type.Name }}{}
	err = db.Get(&{{ $short }}, query, args...)
    if err != nil {
        return entities.{{ .Type.Name }}{}, errors.Wrap(err, "error in {{ .Type.RepoName }}")
    }
	return {{ $short }}, nil
}
{{- else }}
func ({{$shortRepo}} *{{ .Type.RepoName }}) {{ .FuncName }}(ctx context.Context, {{ goparamlist .Fields false true }}, filter *entities.{{ .Type.Name }}Filter, pagination *entities.Pagination) (list entities.List{{ .Type.Name }}, err error) {
	var db = {{ $shortRepo }}.Db
    tx := db_manager.GetTransactionContext(ctx)
    if tx != nil {
        db = tx
    }

	// sql query
	qb := {{$shortRepo}}.findAll{{ .Type.Name }}BaseQuery(ctx, filter, "*")
	{{- range $k, $v := .Fields }}
	    qb = qb.Where(sq.Eq{"`{{ colname .Col }}`": {{ goparam $v }}})
    {{- end }}
	if qb, err = {{$shortRepo}}.addPagination(ctx, qb, pagination); err != nil {
	    return list, err
	}

	query, args, err := qb.ToSql()
    if err != nil {
        return list, errors.Wrap(err, "error in {{ .Type.RepoName }}")
    }

	// run query
    if err = db.Select(&list.Data, query, args...); err != nil {
        return list, errors.Wrap(err, "error in {{ .Type.RepoName }}")
    }

    var listMeta entities.ListMetadata
    qb = {{ $shortRepo }}.findAll{{ .Type.Name }}BaseQuery(ctx, filter, "COUNT(*) AS count")
    {{- range $k, $v := .Fields }}
        qb = qb.Where(sq.Eq{"`{{ colname .Col }}`": {{ goparam $v }}})
    {{- end }}
    if query, args, err = qb.ToSql(); err != nil {
        return list, errors.Wrap(err, "error in {{ .Type.RepoName }}")
    }
    if err = db.Get(&listMeta, query, args...); err != nil {
        return list, errors.Wrap(err, "error in {{ .Type.RepoName }}")
    }

    list.TotalCount = listMeta.Count

    return list, nil
}
{{- end }}

