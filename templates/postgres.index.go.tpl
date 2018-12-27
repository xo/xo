{{- $short := (shortname .Type.Name "err" "sqlstr" "db" "q" "res" "XOLog" .Fields) -}}
{{- $shortRepo := (shortname .Type.RepoName "err" "sqlstr" "db" "q" "res" "XOLog" .Fields) -}}
{{- $table := (schema .Type.Table.TableName) -}}
// {{ .FuncName }} retrieves a row from '{{ $table }}' as a {{ .Type.Name }}.
//
// Generated from index '{{ .Index.IndexName }}'.
{{- if .Index.IsUnique }}
func ({{$shortRepo}} *{{ lowerfirst .Type.RepoName }}) {{ .FuncName }}(ctx context.Context, {{ goparamlist .Fields false true }}) (entities.{{ .Type.Name }}, error) {
	var err error

	var db db_manager.DbInterface = {{ $shortRepo }}.Db
    tx := db_manager.GetTransactionContext(ctx)
    if tx != nil {
        db = tx
    }

	// sql query
	qb := sq.Select("*").From("`{{ $table }}`")
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
func ({{$shortRepo}} *{{ lowerfirst .Type.RepoName }}) {{ .FuncName }}(ctx context.Context, {{ goparamlist .Fields false true }}, filter *entities.{{ .Type.Name }}Filter, pagination *entities.Pagination) (list entities.List{{ .Type.Name }}, err error) {
	var db db_manager.DbInterface = {{ $shortRepo }}.Db
    tx := db_manager.GetTransactionContext(ctx)
    if tx != nil {
        db = tx
    }

	// sql query
	qb := {{$shortRepo}}.findAll{{ .Type.Name }}BaseQuery(ctx, filter, "*")
	{{- range $k, $v := .Fields }}
	    qb = qb.Where(sq.Eq{"`{{ colname .Col }}`": {{ goparam $v }}})
    {{- end }}
	qb, err = {{$shortRepo}}.addPagination(ctx, qb, pagination)
	if err != nil {
	    return list, err
	}

	query, args, err := qb.ToSql()
    if err != nil {
        return list, errors.Wrap(err, "error in {{ .Type.RepoName }}")
    }

	// run query
    var {{ $short }} entities.List{{ .Type.Name }}
    err = db.Select(&{{ $short }}.Data, query, args...)
    if err != nil {
        return list, errors.Wrap(err, "error in {{ .Type.RepoName }}")
    }

    return {{ $short }}, nil
}
{{- end }}

