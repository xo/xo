{{- $shortRepo := (shortname .RepoName "err" "res" "sqlstr" "db" "XOLog") -}}
{{- $short := (shortname .Name "err" "res" "sqlstr" "db" "XOLog") -}}
{{- $table := (schema .Schema .Table.TableName) -}}
{{- $primaryKey := .PrimaryKey }}
{{- if .Comment -}}
// {{ .Comment }}
{{- else -}}
// {{ .RepoName }} represents a row from '{{ $table }}'.
{{- end }}
type {{ .RepoName }} struct {
    db *sqlx.DB
}

{{ if .PrimaryKey }}

// Insert inserts the {{ .RepoName }} to the database.
func ({{ $shortRepo }} *{{ .RepoName }}) Insert({{ $short }} {{ .Name }}) error {
	var err error

{{ if .Table.ManualPk  }}
	// sql insert query, primary key must be provided
	qb := squirrel.Insert("{{ $table }}").Columns("{{ colnames .Fields }}").Values({{ fieldnames .Fields $short }})
    query, args, err := qb.ToSql()
	if err != nil {
	    return err
	}

	// run query
	_, err = {{ $shortRepo }}.db.Exec(query, args...)
	if err != nil {
		return err
	}

{{ else }}
	// sql insert query, primary key provided by autoincrement
	qb := squirrel.Insert("{{ $table }}").Columns("{{ colnames .Fields .PrimaryKey.Name }}").Values({{ fieldnames .Fields $short .PrimaryKey.Name }})
	query, args, err := qb.ToSql()
	if err != nil {
	    return err
	}

	// run query
	res, err := {{ $shortRepo }}.db.Exec(query, args...)
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
{{ end }}

	return nil
}

{{ if ne (fieldnamesmulti .Fields $short .PrimaryKeyFields) "" }}
	// Update updates the {{ .RepoName }} in the database.
	func ({{ $shortRepo }} *{{ .RepoName }}) Update({{ $short }} {{ .Name }}) error {
		var err error

		{{ if gt ( len .PrimaryKeyFields ) 1 }}
			// sql query with composite primary key
			const sqlstr = `UPDATE {{ $table }} SET ` +
				`{{ colnamesquerymulti .Fields ", " 0 .PrimaryKeyFields }}` +
				` WHERE {{ colnamesquery .PrimaryKeyFields " AND " }}`

			// run query
			XOLog(sqlstr, {{ fieldnamesmulti .Fields $short .PrimaryKeyFields }}, {{ fieldnames .PrimaryKeyFields $short}})
			_, err = db.Exec(sqlstr, {{ fieldnamesmulti .Fields $short .PrimaryKeyFields }}, {{ fieldnames .PrimaryKeyFields $short}})
			return err
		{{- else }}
			// sql query
			qb := squirrel.Update("{{ $table }}").SetMap(map[string]interface{}{
			{{- range .Fields }}
			    {{- if ne .Name $primaryKey.Name }}
			    "{{ .Col.ColumnName }}": {{ $short }}.{{ .Name }},
			    {{- end }}
            {{- end }}
            }).Where(squirrel.Eq{"{{ .PrimaryKey.Col.ColumnName }}": {{ $short }}.{{ .PrimaryKey.Name }}})
            query, args, err := qb.ToSql()
            if err != nil {
                return err
            }

			// run query
			_, err = {{ $shortRepo }}.db.Exec(query, args...)
			return err
		{{- end }}
	}
{{ else }}
	// Update statements omitted due to lack of fields other than primary key
{{ end }}

// Delete deletes the {{ .RepoName }} from the database.
func ({{ $shortRepo }} *{{ .RepoName }}) Delete({{ $short }} {{ .Name }}) error {
	var err error
	{{ if gt ( len .PrimaryKeyFields ) 1 }}
		// sql query with composite primary key
		const sqlstr = `DELETE FROM {{ $table }} WHERE {{ colnamesquery .PrimaryKeyFields " AND " }}`

		// run query
		XOLog(sqlstr, {{ fieldnames .PrimaryKeyFields $short }})
		_, err = db.Exec(sqlstr, {{ fieldnames .PrimaryKeyFields $short }})
		if err != nil {
			return err
		}
	{{- else }}
		// sql query
		qb := squirrel.Delete("{{ $table }}").Where("{{ colname .PrimaryKey.Col}}", {{ $short }}.{{ .PrimaryKey.Name }})
		query, args, err := qb.ToSql()
        if err != nil {
            return err
        }

		// run query
		_, err = {{ $shortRepo }}.db.Exec(query, args...)
		if err != nil {
			return err
		}
	{{- end }}

	return nil
}
{{- end }}

