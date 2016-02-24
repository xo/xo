{{- if .Comment -}}
// {{ .Comment }}
{{- else -}}
// {{ .Name }} represents a row from {{ schema .Schema .Table.TableName }}.
{{- end }}
type {{ .Name }} struct {
{{- range .Fields }}
	{{ .Name }} {{ retype .Type }} // {{ .Col.ColumnName }}
{{- end }}
{{- if .PrimaryKey }}

	// xo fields
	_exists, _deleted bool
{{ end }}
}

{{ if .PrimaryKey }}
// Exists determines if the {{ .Name }} exists in the database.
func ({{ shortname .Name }} *{{ .Name }}) Exists() bool {
	return {{ shortname .Name }}._exists
}

// Deleted provides information if the {{ .Name }} has been deleted from the database.
func ({{ shortname .Name }} *{{ .Name }}) Deleted() bool {
	return {{ shortname .Name }}._deleted
}

// Insert inserts the {{ .Name }} to the database.
func ({{ shortname .Name }} *{{ .Name }}) Insert(db XODB) error {
	var err error

	// if already exist, bail
	if {{ shortname .Name }}._exists {
		return errors.New("insert failed: already exists")
	}

	// sql query
	const sqlstr = `INSERT INTO {{ schema .Schema .Table.TableName }} (` +
		`{{ colnames .Fields .PrimaryKey.Name }}` +
		`) VALUES (` +
		`{{ colvals .Fields .PrimaryKey.Name }}` +
		`)`

	// run query
	XOLog(sqlstr, {{ fieldnames .Fields (shortname .Name) .PrimaryKey.Name }})
	res, err := db.Exec(sqlstr, {{ fieldnames .Fields (shortname .Name) .PrimaryKey.Name }})
	if err != nil {
		return err
	}

	// retrieve id
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	// set primary key and existence
	{{ shortname .Name }}.{{ .PrimaryKey.Name }} = {{ .PrimaryKey.Type }}(id)
	{{ shortname .Name }}._exists = true

	return nil
}

// Update updates the {{ .Name }} in the database.
func ({{ shortname .Name }} *{{ .Name }}) Update(db XODB) error {
	var err error

	// if doesn't exist, bail
	if !{{ shortname .Name }}._exists {
		return errors.New("update failed: does not exist")
	}

	// if deleted, bail
	if {{ shortname .Name }}._deleted {
		return errors.New("update failed: marked for deletion")
	}

	// sql query
	const sqlstr = `UPDATE {{ schema .Schema .Table.TableName }} SET ` +
		`{{ colnamesquery .Fields ", " .PrimaryKey.Name }}` +
		` WHERE {{ .PrimaryKey.Col.ColumnName }} = ?`

	// run query
	XOLog(sqlstr, {{ fieldnames .Fields (shortname .Name) .PrimaryKey.Name }}, {{ shortname .Name }}.{{ .PrimaryKey.Name }})
	_, err = db.Exec(sqlstr, {{ fieldnames .Fields (shortname .Name) .PrimaryKey.Name }}, {{ shortname .Name }}.{{ .PrimaryKey.Name }})
	return err
}

// Save saves the {{ .Name }} to the database.
func ({{ shortname .Name }} *{{ .Name }}) Save(db XODB) error {
	if {{ shortname .Name }}.Exists() {
		return {{ shortname .Name }}.Update(db)
	}

	return {{ shortname .Name }}.Insert(db)
}

// Delete deletes the {{ .Name }} from the database.
func ({{ shortname .Name }} *{{ .Name }}) Delete(db XODB) error {
	var err error

	// if doesn't exist, bail
	if !{{ shortname .Name }}._exists {
		return nil
	}

	// if deleted, bail
	if {{ shortname .Name }}._deleted {
		return nil
	}

	// sql query
	const sqlstr = `DELETE FROM {{ schema .Schema .Table.TableName }} WHERE {{ .PrimaryKey.Col.ColumnName }} = ?`

	// run query
	XOLog(sqlstr, {{ shortname .Name }}.{{ .PrimaryKey.Name }})
	_, err = db.Exec(sqlstr, {{ shortname .Name }}.{{ .PrimaryKey.Name }})
	if err != nil {
		return err
	}

	// set deleted
	{{ shortname .Name }}._deleted = true

	return nil
}
{{- end }}

