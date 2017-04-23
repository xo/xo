{{- $short := (shortname .Name "err" "res" "sqlstr" "db" "XOLog") -}}
{{- $table := (schema .Schema .Table.TableName) -}}
{{- if .Comment -}}
// {{ .Comment }}
{{- else -}}
// {{ .Name }} represents a row from '{{ $table }}'.
{{- end }}
type {{ .Name }} struct {
{{- range .Fields }}
	{{ .Name }} {{ retype .Type }} `json:"{{ .Col.ColumnName }}"` // {{ .Col.ColumnName }}
{{- end }}
{{- if .PrimaryKey }}

	// xo fields
	_exists, _deleted bool
{{ end }}
}

{{ if .PrimaryKey }}
// Exists determines if the {{ .Name }} exists in the database.
func ({{ $short }} *{{ .Name }}) Exists() bool {
	return {{ $short }}._exists
}

// Deleted provides information if the {{ .Name }} has been deleted from the database.
func ({{ $short }} *{{ .Name }}) Deleted() bool {
	return {{ $short }}._deleted
}

// Insert inserts the {{ .Name }} to the database.
func ({{ $short }} *{{ .Name }}) Insert(db XODB) error {
	var err error

	// if already exist, bail
	if {{ $short }}._exists {
		return errors.New("insert failed: already exists")
	}

    err = db.BeforeInsert({{ $short }})
    if err != nil {
        return err
    }

{{ if .Table.ManualPk  }}
	// sql insert query, primary key must be provided
	const sqlstr = `INSERT INTO {{ $table }} (` +
		`{{ colnames .Fields }}` +
		`) VALUES (` +
		`{{ colvals .Fields }}` +
		`)`

	// run query
	XOLog(sqlstr, {{ fieldnames .Fields $short }})
	_, err = db.Exec(sqlstr, {{ fieldnames .Fields $short }})
	if err != nil {
		return err
	}

	// set existence
	{{ $short }}._exists = true
{{ else }}
	// sql insert query, primary key provided by autoincrement
	const sqlstr = `INSERT INTO {{ $table }} (` +
		`{{ colnames .Fields .PrimaryKey.Name }}` +
		`) VALUES (` +
		`{{ colvals .Fields .PrimaryKey.Name }}` +
		`)`

	// run query
	XOLog(sqlstr, {{ fieldnames .Fields $short .PrimaryKey.Name }})
	res, err := db.Exec(sqlstr, {{ fieldnames .Fields $short .PrimaryKey.Name }})
	if err != nil {
		return err
	}

	// retrieve id
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	// set primary key and existence
	{{ $short }}.{{ .PrimaryKey.Name }} = {{ .PrimaryKey.Type }}(id)
	{{ $short }}._exists = true
{{ end }}

    return db.AfterInsert({{ $short }})
}

{{ if ne (fieldnames .Fields $short .PrimaryKey.Name) "" }}
	// Update updates the {{ .Name }} in the database.
	func ({{ $short }} *{{ .Name }}) Update(db XODB) error {
		var err error

		// if doesn't exist, bail
		if !{{ $short }}._exists {
			return errors.New("update failed: does not exist")
		}

		// if deleted, bail
		if {{ $short }}._deleted {
			return errors.New("update failed: marked for deletion")
		}

        err = db.BeforeUpdate({{ $short }})
        if err != nil {
            return err
        }

		// sql query
		const sqlstr = `UPDATE {{ $table }} SET ` +
			`{{ colnamesquery .Fields ", " .PrimaryKey.Name }}` +
			` WHERE {{ colname .PrimaryKey.Col }} = ?`

		// run query
		XOLog(sqlstr, {{ fieldnames .Fields $short .PrimaryKey.Name }}, {{ $short }}.{{ .PrimaryKey.Name }})
		_, err = db.Exec(sqlstr, {{ fieldnames .Fields $short .PrimaryKey.Name }}, {{ $short }}.{{ .PrimaryKey.Name }})
		if err != nil {
            return err
        }

        return db.AfterUpdate({{ $short }})
	}

	// Save saves the {{ .Name }} to the database.
	func ({{ $short }} *{{ .Name }}) Save(db XODB) error {
        var err error

        err = db.BeforeSave({{ $short }})
        if err != nil {
            return err
        }

		if {{ $short }}.Exists() {
			err = {{ $short }}.Update(db)
		} else {
            err = {{ $short }}.Insert(db)
        }

        if err != nil {
            return err
        }

        return db.AfterSave({{ $short }})
	}
{{ else }}
	// Update statements omitted due to lack of fields other than primary key
{{ end }}

// Delete deletes the {{ .Name }} from the database.
func ({{ $short }} *{{ .Name }}) Delete(db XODB) error {
	var err error

	// if doesn't exist, bail
	if !{{ $short }}._exists {
		return nil
	}

	// if deleted, bail
	if {{ $short }}._deleted {
		return nil
	}

    err = db.BeforeDelete({{ $short }})
    if err != nil {
        return err
    }

	// sql query
	const sqlstr = `DELETE FROM {{ $table }} WHERE {{ colname .PrimaryKey.Col }} = ?`

	// run query
	XOLog(sqlstr, {{ $short }}.{{ .PrimaryKey.Name }})
	_, err = db.Exec(sqlstr, {{ $short }}.{{ .PrimaryKey.Name }})
	if err != nil {
		return err
	}

	// set deleted
	{{ $short }}._deleted = true

	return db.AfterDelete({{ $short }})
}
{{- end }}

