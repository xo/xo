{{- $type := .Data -}}
{{- $short := (shortname $type.Name "err" "res" "sqlstr" "db" "logf") -}}
{{- $table := (schema $type.Table.TableName) -}}
{{- if $type.Comment -}}
// {{ $type.Comment }}
{{- else -}}
// {{ $type.Name }} represents a row from '{{ $table }}'.
{{- end }}
type {{ $type.Name }} struct {
{{ range $type.Fields -}}
	{{ .Name }} {{ retype .Type }} {{ fieldtag . }} // {{ .Col.ColumnName }}
{{ end -}}
{{ if $type.PrimaryKey -}}
	// xo fields
	_exists, _deleted bool
{{- end }}
}

{{ if $type.PrimaryKey -}}
// Exists returns true when the {{ $type.Name }} exists in the database.
func ({{ $short }} *{{ $type.Name }}) Exists() bool {
	return {{ $short }}._exists
}

// Deleted returns true when the {{ $type.Name }} has been marked for deletion from
// the database.
func ({{ $short }} *{{ $type.Name }}) Deleted() bool {
	return {{ $short }}._deleted
}

// Insert{{ if context_both }}Context{{ end }} inserts the {{ $type.Name }} to the database.
func ({{ $short }} *{{ $type.Name }}) Insert{{ if context_both }}Context{{ end }}({{ if context }}ctx context.Context, {{ end }}db DB) error {
	switch {
	case {{ $short }}._exists: // already exists
		return logerror(&ErrInsertFailed{ErrAlreadyExists})
	case {{ $short }}._deleted: // deleted
		return logerror(&ErrInsertFailed{ErrMarkedForDeletion})
	}
{{ if $type.Table.ManualPk -}}
	// insert (basic)
	const sqlstr = `INSERT INTO {{ $table }} (` +
		`{{ colnames $type.Fields }}` +
		`) VALUES (` +
		`{{ colvals $type.Fields }}` +
		`)`
	// run
	logf(sqlstr, {{ fieldnames $type.Fields $short }})
	if err := db.QueryRow{{ if context }}Context{{ end }}({{ if context }}ctx, {{ end }}sqlstr, {{ fieldnames $type.Fields $short }}).Scan(&{{ $short }}.{{ $type.PrimaryKey.Name }}); err != nil {
		return logerror(err)
	}
{{- else -}}
	// insert (primary key generated and returned by database)
	const sqlstr = `INSERT INTO {{ $table }} (` +
		`{{ colnames $type.Fields $type.PrimaryKey.Name }}` +
		`) VALUES (` +
		`{{ colvals $type.Fields $type.PrimaryKey.Name }}` +
		`){{ if (driver "postgres" "oracle") }} RETURNING {{ colname $type.PrimaryKey.Col }}{{ end }}{{ if (driver "oracle") }} /*LASTINSERTID*/ INTO :pk{{ end }}{{ if (driver "sqlserver") }}; select ID = convert(bigint, SCOPE_IDENTITY()){{ end }}`
	// run
	logf(sqlstr, {{ fieldnames $type.Fields $short $type.PrimaryKey.Name }}{{ if (driver "oracle") }}, nil{{ end }})
{{ if (driver "postgres") -}}
	if err := db.QueryRow{{ if context }}Context{{ end }}({{ if context }}ctx, {{ end }}sqlstr, {{ fieldnames $type.Fields $short $type.PrimaryKey.Name }}).Scan(&{{ $short }}.{{ $type.PrimaryKey.Name }}); err != nil {
		return logerror(err)
	}
{{- else if (driver "sqlserver") -}}
	rows, err := db.Query{{ if context }}Context{{ end }}({{ if context }}ctx, {{ end }}sqlstr, {{ fieldnames $type.Fields $short $type.PrimaryKey.Name }})
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
	if _, err := db.Exec{{ if context }}Context{{ end }}({{ if context }}ctx, {{ end }}sqlstr, {{ fieldnames $type.Fields $short $type.PrimaryKey.Name }}, sql.Named("pk", sql.Out{Dest: &id})); err != nil {
		return err
	}
{{- else -}}
	res, err := db.Exec{{ if context }}Context{{ end }}({{ if context }}ctx, {{ end }}sqlstr, {{ fieldnames $type.Fields $short $type.PrimaryKey.Name }})
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
	{{ $short }}.{{ $type.PrimaryKey.Name }} = {{ $type.PrimaryKey.Type }}(id)
{{- end }}
{{- end }}
	// set exists
	{{ $short }}._exists = true
	return nil
}

{{ if context_both -}}
// Insert inserts the {{ $type.Name }} to the database.
func ({{ $short }} *{{ $type.Name }}) Insert(db DB) error {
	return {{ $short }}.InsertContext(context.Background(), db)
}
{{- end }}

{{ if eq (fieldnamesmulti $type.Fields $short $type.PrimaryKeyFields) "" -}}
// ------ NOTE: Update statements omitted due to lack of fields other than primary key ------ 
{{- else -}}
// Update{{ if context_both }}Context{{ end }} updates a {{ $type.Name }} in the database.
func ({{ $short }} *{{ $type.Name }}) Update{{ if context_both }}Context{{ end }}({{ if context }}ctx context.Context, {{ end }}db DB) error {
	switch {
	case !{{ $short }}._exists: // doesn't exist
		return logerror(&ErrUpdateFailed{ErrDoesNotExist})
	case {{ $short }}._deleted: // deleted
		return logerror(&ErrUpdateFailed{ErrMarkedForDeletion})
	}
{{ if (driver "postgres") -}}
	// update with composite primary key
	const sqlstr = `UPDATE {{ $table }} SET (` +
		`{{ colnamesmulti $type.Fields $type.PrimaryKeyFields }}` +
		`) = ( ` +
		`{{ colvalsmulti $type.Fields $type.PrimaryKeyFields }}` +
		`) WHERE {{ colnamesquerymulti $type.PrimaryKeyFields " AND " (startcount $type.Fields $type.PrimaryKeyFields) nil }}`
{{- else -}}
	// update with primary key
	const sqlstr = `UPDATE {{ $table }} SET ` +
		`{{ colnamesquerymulti $type.Fields ", " 0 $type.PrimaryKeyFields }}` +
		` WHERE {{ colnamesquerymulti $type.PrimaryKeyFields " AND " (startcount $type.Fields $type.PrimaryKeyFields) nil }}`
{{- end }}
	// run
	logf(sqlstr, {{ fieldnamesmulti $type.Fields $short $type.PrimaryKeyFields }}, {{ fieldnames $type.PrimaryKeyFields $short}})
	if _, err := db.Exec{{ if context }}Context{{ end }}({{ if context }}ctx, {{ end }}sqlstr, {{ fieldnamesmulti $type.Fields $short $type.PrimaryKeyFields }}, {{ fieldnames $type.PrimaryKeyFields $short}}); err != nil {
		return logerror(err)
	}
	return nil
}

{{ if context_both -}}
// Update updates a {{ $type.Name }} in the database.
func ({{ $short }} *{{ $type.Name }}) Update(db DB) error {
	return {{ $short }}.UpdateContext(context.Background(), db)
}
{{- end }}

// Save{{ if context_both }}Context{{ end }} saves the {{ $type.Name }} to the database.
func ({{ $short }} *{{ $type.Name }}) Save{{ if context_both }}Context{{ end }}({{ if context }}ctx context.Context, {{ end }}db DB) error {
	if {{ $short }}.Exists() {
		return {{ $short }}.Update{{ if context_both }}Context{{ end }}({{ if context }}ctx, {{ end }}db)
	}
	return {{ $short }}.Insert{{ if context_both }}Context{{ end }}({{ if context}}ctx, {{ end }}db)
}

{{ if context_both -}}
// Save saves the {{ $type.Name }} to the database.
func ({{ $short }} *{{ $type.Name }}) Save(db DB) error {
	return {{ $short }}.SaveContext(context.Background(), db)
}
{{- end }}

{{ if (driver "postgres") -}}
// Upsert{{ if context_both }}Context{{ end }} performs an upsert for {{ $type.Name }}.
//
// NOTE: PostgreSQL 9.5+ only
func ({{ $short }} *{{ $type.Name }}) Upsert{{ if context_both }}Context{{ end }}({{ if context }}ctx context.Context, {{ end }}db DB) error {
	switch {
	case {{ $short }}._deleted: // deleted
		return logerror(&ErrUpsertFailed{ErrMarkedForDeletion})
	}
	// upsert
	const sqlstr = `INSERT INTO {{ $table }} (` +
		`{{ colnames $type.Fields }}` +
		`) VALUES (` +
		`{{ colvals $type.Fields }}` +
		`) ON CONFLICT ({{ colnames $type.PrimaryKeyFields }}) DO UPDATE SET (` +
		`{{ colnames $type.Fields }}` +
		`) = (` +
		`{{ colprefixnames $type.Fields "EXCLUDED" }}` +
		`)`
	// run
	logf(sqlstr, {{ fieldnames $type.Fields $short }})
	if _, err := db.Exec{{ if context }}Context{{ end }}({{ if context }}ctx, {{ end }}sqlstr, {{ fieldnames $type.Fields $short }}); err != nil {
		return err
	}
	// set exists
	{{ $short }}._exists = true
	return nil
}
{{- end -}}

{{ if context_both -}}
// Upsert performs an upsert for {{ $type.Name }}.
func ({{ $short }} *{{ $type.Name }}) Upsert(db DB) error {
	return {{ $short }}.UpsertContext(context.Background(), db)
}
{{- end }}
{{- end }}

// Delete{{ if context_both }}Context{{ end }} deletes the {{ $type.Name }} from the database.
func ({{ $short }} *{{ $type.Name }}) Delete{{ if context_both }}Context{{ end }}({{ if context }} ctx context.Context, {{ end }}db DB) error {
	switch {
	case !{{ $short }}._exists: // doesn't exist
		return nil
	case {{ $short }}._deleted: // deleted
		return nil
	}
{{ if gt (len $type.PrimaryKeyFields) 1 -}}
	// delete with composite primary key
	const sqlstr = `DELETE FROM {{ $table }} WHERE {{ colnamesquery $type.PrimaryKeyFields " AND " }}`
	// run
	logf(sqlstr, {{ fieldnames $type.PrimaryKeyFields $short }})
	if _, err := db.Exec{{ if context }}Context{{ end }}({{ if context }}ctx, {{ end }}sqlstr, {{ fieldnames $type.PrimaryKeyFields $short }}); err != nil {
		return logerror(err)
	}
{{- else -}}
	// delete with single primary key
	const sqlstr = `DELETE FROM {{ $table }} WHERE {{ colname $type.PrimaryKey.Col }} = {{ nthparam 0 }}`
	// run
	logf(sqlstr, {{ $short }}.{{ $type.PrimaryKey.Name }})
	if _, err := db.Exec{{ if context }}Context{{ end }}({{ if context }}ctx, {{ end }}sqlstr, {{ $short }}.{{ $type.PrimaryKey.Name }}); err != nil {
		return logerror(err)
	}
{{- end }}
	// set deleted
	{{ $short }}._deleted = true
	return nil
}
{{- end }}

{{ if context_both -}}
// Delete deletes the {{ $type.Name }} from the database.
func ({{ $short }} *{{ $type.Name }}) Delete(db DB) error {
	return {{ $short }}.DeleteContext(context.Background(), db)
}
{{- end }}
