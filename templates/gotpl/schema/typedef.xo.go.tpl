{{- $t := .Data -}}
{{- if $t.Comment -}}
// {{ $t.Comment | eval $t.Name }}
{{- else -}}
// {{ $t.Name }} represents a row from '{{ schema $t.Table.TableName }}'.
{{- end }}
type {{ $t.Name }} struct {
{{ range $t.Fields -}}
	{{ field . }}
{{ end }}{{ if $t.PrimaryKey -}}
	// xo fields
	_exists, _deleted bool
{{ end -}}
}

{{ if $t.PrimaryKey -}}
// Exists returns true when the {{ $t.Name }} exists in the database.
func ({{ short $t.Name }} *{{ $t.Name }}) Exists() bool {
	return {{ short $t.Name }}._exists
}

// Deleted returns true when the {{ $t.Name }} has been marked for deletion from
// the database.
func ({{ short $t.Name }} *{{ $t.Name }}) Deleted() bool {
	return {{ short $t.Name }}._deleted
}

// {{ func_name_context "Insert" }} inserts the {{ $t.Name }} to the database.
{{ recv_context $t "Insert" }} {
	switch {
	case {{ short $t.Name }}._exists: // already exists
		return logerror(&ErrInsertFailed{ErrAlreadyExists})
	case {{ short $t.Name }}._deleted: // deleted
		return logerror(&ErrInsertFailed{ErrMarkedForDeletion})
	}
{{ if $t.Table.ManualPk -}}
	// insert (manual)
	{{ sqlstr "insert_manual" $t }}
	// run
	{{ logf $t }}
	if _, err := {{ db_prefix "Exec" false $t }}; err != nil {
		return logerror(err)
	}
{{- else -}}
	// insert (primary key generated and returned by database)
	{{ sqlstr "insert" $t }}
	// run
	{{ logf $t $t.PrimaryKey.Name }}
{{ if (driver "postgres") -}}
	if err := {{ db_prefix "QueryRow" true $t }}.Scan(&{{ short $t.Name }}.{{ $t.PrimaryKey.Name }}); err != nil {
		return logerror(err)
	}
{{- else if (driver "sqlserver") -}}
	rows, err := {{ db_prefix "Query" true $t }}
	if err != nil {
		return logerror(err)
	}
	defer rows.Close()
	// retrieve id
	var id int64
	for rows.Next() {
		if err := rows.Scan(&id); err != nil {
			return logerror(err)
		}
	}
	if err := rows.Err(); err != nil {
		return logerror(err)
	}
{{- else if (driver "oracle") -}}
	var id int64
	if _, err := {{ db_prefix "Exec" true $t (named "pk" "&id" true) }}; err != nil {
		return err
	}
{{- else -}}
	res, err := {{ db_prefix "Exec" true $t }}
	if err != nil {
		return err
	}
	// retrieve id
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
{{- end -}}
{{ if not (driver "postgres") -}}
	// set primary key
	{{ short $t.Name }}.{{ $t.PrimaryKey.Name }} = {{ $t.PrimaryKey.Type }}(id)
{{- end }}
{{- end }}
	// set exists
	{{ short $t.Name }}._exists = true
	return nil
}

{{ if context_both -}}
// Insert inserts the {{ $t.Name }} to the database.
{{ recv $t "Insert" }} {
	return {{ short $t.Name }}.InsertContext(context.Background(), db)
}
{{- end }}


{{ if eq (len $t.Fields) (len $t.PrimaryKeyFields) -}}
// ------ NOTE: Update statements omitted due to lack of fields other than primary key ------
{{- else -}}
// {{ func_name_context "Update" }} updates a {{ $t.Name }} in the database.
{{ recv_context $t "Update" }} {
	switch {
	case !{{ short $t.Name }}._exists: // doesn't exist
		return logerror(&ErrUpdateFailed{ErrDoesNotExist})
	case {{ short $t.Name }}._deleted: // deleted
		return logerror(&ErrUpdateFailed{ErrMarkedForDeletion})
	}
	// update with {{ if driver "postgres" }}composite {{ end }}primary key
	{{ sqlstr "update" $t }}
	// run
	{{ logf_update $t }}
	if _, err := {{ db_update "Exec" $t }}; err != nil {
		return logerror(err)
	}
	return nil
}

{{ if context_both -}}
// Update updates a {{ $t.Name }} in the database.
{{ recv $t "Update" }} {
	return {{ short $t.Name }}.UpdateContext(context.Background(), db)
}
{{- end }}

// {{ func_name_context "Save" }} saves the {{ $t.Name }} to the database.
{{ recv_context $t "Save" }} {
	if {{ short $t.Name }}.Exists() {
		return {{ short $t.Name}}.{{ func_name_context "Update" }}({{ if context }}ctx, {{ end }}db)
	}
	return {{ short $t.Name}}.{{ func_name_context "Insert" }}({{ if context }}ctx, {{ end }}db)
}

{{ if context_both -}}
// Save saves the {{ $t.Name }} to the database.
{{ recv $t "Save" }} {
	if {{ short $t.Name }}._exists {
		return {{ short $t.Name }}.UpdateContext(context.Background(), db)
	} 
	return {{ short $t.Name }}.InsertContext(context.Background(), db)
}
{{- end }}

{{ if (driver "postgres") -}}
// {{ func_name_context "Upsert" }} performs an upsert for {{ $t.Name }}.
//
// NOTE: PostgreSQL 9.5+ only
{{ recv_context $t "Upsert" }} {
	switch {
	case {{ short $t.Name }}._deleted: // deleted
		return logerror(&ErrUpsertFailed{ErrMarkedForDeletion})
	}
	// upsert
	{{ sqlstr "upsert" $t }}
	// run
	{{ logf $t }}
	if _, err := {{ db_prefix "Exec" false $t }}; err != nil {
		return err
	}
	// set exists
	{{ short $t.Name }}._exists = true
	return nil
}
{{- end -}}

{{ if context_both -}}
// Upsert performs an upsert for {{ $t.Name }}.
{{ recv $t "Upsert" }} {
	return {{ short $t.Name }}.UpsertContext(context.Background(), db)
}
{{- end -}}
{{- end -}}
{{- end }}

// {{ func_name_context "Delete" }} deletes the {{ $t.Name }} from the database.
{{ recv_context $t "Delete" }} {
	switch {
	case !{{ short $t.Name }}._exists: // doesn't exist
		return nil
	case {{ short $t.Name }}._deleted: // deleted
		return nil
	}
{{ if gt (len $t.PrimaryKeyFields) 1 -}}
	// delete with composite primary key
	{{ sqlstr "delete" $t }}
	// run
	{{ logf_pkeys $t }}
	if _, err := {{ db "Exec" (names (print (short $t.Name) ".") $t.PrimaryKeyFields) }}; err != nil {
		return logerror(err)
	}
{{- else -}}
	// delete with single primary key
	{{ sqlstr "delete" $t }}
	// run
	{{ logf_pkeys $t }}
	if _, err := {{ db "Exec" (print (short $t.Name) "." $t.PrimaryKey.Name) }}; err != nil {
		return logerror(err)
	}
{{- end }}
	// set deleted
	{{ short $t.Name }}._deleted = true
	return nil
}

{{ if context_both -}}
// Delete deletes the {{ $t.Name }} from the database.
{{ recv $t "Delete" }} {
	return {{ short $t.Name }}.DeleteContext(context.Background(), db)
}
{{- end }}
