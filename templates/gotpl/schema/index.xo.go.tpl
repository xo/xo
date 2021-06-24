{{- $i := .Data -}}
// {{ func_name_context $i }} retrieves a row from '{{ schema $i.Type.Table.TableName }}' as a {{ $i.Type.Name }}.
//
// Generated from index '{{ $i.Index.IndexName }}'.
{{ func_context $i }} {
	// query
	{{ sqlstr "index" $i }}
	// run
	logf(sqlstr, {{ params $i.Fields false }})
{{- if $i.Index.IsUnique }}
	{{ short $i.Type }} := {{ $i.Type.Name }}{
	{{- if $i.Type.PrimaryKey }}
		_exists: true,
	{{ end -}}
	}
	if err := {{ db "QueryRow"  $i }}.Scan({{ names (print "&" (short $i.Type) ".") $i.Type }}); err != nil {
		return nil, logerror(err)
	}
	return &{{ short $i.Type }}, nil
{{- else }}
	rows, err := {{ db "Query" $i }}
	if err != nil {
		return nil, logerror(err)
	}
	defer rows.Close()
	// process
	var res []*{{ $i.Type.Name }}
	for rows.Next() {
		{{ short $i.Type }} := {{ $i.Type.Name }}{
		{{- if $i.Type.PrimaryKey }}
			_exists: true,
		{{ end -}}
		}
		// scan
		if err := rows.Scan({{ names_ignore (print "&" (short $i.Type) ".")  $i.Type }}); err != nil {
			return nil, logerror(err)
		}
		res = append(res, &{{ short $i.Type }})
	}
	if err := rows.Err(); err != nil {
		return nil, logerror(err)
	}
	return res, nil
{{- end }}
}

{{ if context_both -}}
// {{ func_name $i }} retrieves a row from '{{ schema $i.Type.Table.TableName }}' as a {{ $i.Type.Name }}.
//
// Generated from index '{{ $i.Index.IndexName }}'.
{{ func $i }} {
	return {{ func_name_context $i }}({{ names "" "context.Background()" "db" $i }})
}
{{- end }}

