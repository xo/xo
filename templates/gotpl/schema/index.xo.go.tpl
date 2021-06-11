{{- $index := .Data -}}
{{- $short := (shortname $index.Type.Name "err" "sqlstr" "db" "rows" "res" "logf" $index.Fields) -}}
{{- $table := (schema $index.Type.Table.TableName) -}}
// {{ $index.FuncName }}{{ if context_both }}Context{{ end }} retrieves a row from '{{ $table }}' as a {{ $index.Type.Name }}.
//
// Generated from index '{{ $index.Index.IndexName }}'.
func {{ $index.FuncName }}{{ if context_both }}Context{{ end }}({{ if context }}ctx context.Context, {{ end }}db DB{{ paramlist $index.Fields true true }}) ({{ if not $index.Index.IsUnique }}[]{{ end }}*{{ $index.Type.Name }}, error) {
	// query
	const sqlstr = `SELECT ` +
		`{{ colnames $index.Type.Fields }} ` +
		`FROM {{ $table }} ` +
		`WHERE {{ colnamesquery $index.Fields " AND " }}`
	// run
	logf(sqlstr{{ paramlist $index.Fields true false }})
{{ if $index.Index.IsUnique -}}
	{{ $short }} := {{ $index.Type.Name }}{
	{{- if $index.Type.PrimaryKey }}
		_exists: true,
	{{ end -}}
	}
	if err := db.QueryRow{{ if context }}Context{{ end }}({{ if context }}ctx, {{ end }}sqlstr{{ paramlist $index.Fields true false }}).Scan({{ fieldnames $index.Type.Fields (print "&" $short) }}); err != nil {
		return nil, logerror(err)
	}
	return &{{ $short }}, nil
{{- else -}}
	rows, err := db.Query{{ if context }}Context{{ end }}({{ if context }}ctx, {{ end }}sqlstr{{ paramlist $index.Fields true false }})
	if err != nil {
		return nil, logerror(err)
	}
	defer rows.Close()
	// process
	var res []*{{ $index.Type.Name }}
	for rows.Next() {
		{{ $short }} := {{ $index.Type.Name }}{
		{{- if $index.Type.PrimaryKey }}
			_exists: true,
		{{ end -}}
		}
		// scan
		if err := rows.Scan({{ fieldnames $index.Type.Fields (print "&" $short) }}); err != nil {
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

{{ if context_both -}}
// {{ $index.FuncName }} retrieves a row from '{{ $table }}' as a {{ $index.Type.Name }}.
//
// Generated from index '{{ $index.Index.IndexName }}'.
func {{ $index.FuncName }}(db DB{{ paramlist $index.Fields true true }}) ({{ if not $index.Index.IsUnique }}[]{{ end }}*{{ $index.Type.Name }}, error) {
	return {{ $index.FuncName }}Context(context.Background(), db{{ paramlist $index.Fields true false }})
}
{{- end }}

