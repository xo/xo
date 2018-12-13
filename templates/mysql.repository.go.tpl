{{- $shortRepo := (shortname .RepoName "err" "res" "sqlstr" "db" "XOLog") -}}
{{- $short := (shortname .Name "err" "res" "sqlstr" "db" "XOLog") -}}
{{- $table := (schema .Table.TableName) -}}
{{- $primaryKey := .PrimaryKey }}
{{- if .Comment -}}
// {{ .Comment }}
{{- else -}}

type I{{ .RepoName }} interface {
    Insert({{ $short }} entities.{{ .Name }}) (*entities.{{ .Name }}, error)
    Update({{ $short }} entities.{{ .Name }}) error
    Delete({{ $short }} entities.{{ .Name }}) error
    FindAll({{$short}}Filter *entities.{{ .Name }}Filter, pagination *entities.Pagination) ([]entities.{{ .Name }}, error)
    {{- range .Indexes }}
    {{ .FuncName }}({{ goparamlist .Fields false true }}) ({{ if not .Index.IsUnique }}[]{{ end }}*entities.{{ .Type.Name }}, error)
    {{- end }}
}

// {{ lowerfirst .RepoName }} represents a row from '{{ $table }}'.
{{- end }}
type {{ lowerfirst .RepoName }} struct {
    db *sqlx.DB
}

func New{{ .RepoName }}(db *sqlx.DB) I{{ .RepoName }} {
    return &{{ lowerfirst .RepoName }}{db: db}
}

{{ if .PrimaryKey }}

// Insert inserts the {{ lowerfirst .RepoName }} to the database.
func ({{ $shortRepo }} *{{ lowerfirst .RepoName }}) Insert({{ $short }} entities.{{ .Name }}) (*entities.{{ .Name }}, error) {
	var err error

{{ if .Table.ManualPk  }}
	// sql insert query, primary key must be provided
	qb := sq.Insert("{{ $table }}").Columns({{ colnameswrap .Fields }}).
	    Values({{ fieldnames .Fields $short }})
    query, args, err := qb.ToSql()
	if err != nil {
	    return nil, err
	}

	// run query
	res, err = {{ $shortRepo }}.db.Exec(query, args...)
	if err != nil {
		return nil, err
	}

{{ else }}
	// sql insert query, primary key provided by autoincrement
	qb := sq.Insert("{{ $table }}").Columns({{ colnameswrap .Fields .PrimaryKey.Name }}).
	    Values({{ fieldnames .Fields $short .PrimaryKey.Name }})
	query, args, err := qb.ToSql()
	if err != nil {
	    return nil, err
	}

	// run query
	res, err := {{ $shortRepo }}.db.Exec(query, args...)
	if err != nil {
		return nil, err
	}
{{ end }}

    // retrieve id
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	err = {{ $shortRepo }}.db.Select(&{{ $short }}, fmt.Sprintf("SELECT * FROM `{{ $table }}` WHERE `{{ .PrimaryKey.Col.ColumnName }}` = %d", id))

	return &{{ $short }}, err
}

{{ if ne (fieldnamesmulti .Fields $short .PrimaryKeyFields) "" }}
	// Update updates the {{ lowerfirst .RepoName }} in the database.
	func ({{ $shortRepo }} *{{ lowerfirst .RepoName }}) Update({{ $short }} entities.{{ .Name }}) error {
		var err error

		{{ if gt ( len .PrimaryKeyFields ) 1 }}
			// sql query with composite primary key
			qb := sq.Update("{{ $table }}").SetMap(map[string]interface{}{
            {{- range .Fields }}
                "{{ .Col.ColumnName }}": {{ $short }}.{{ .Name }},
            {{- end }}
            }).Where(sq.Eq{
            {{- range .PrimaryKeyFields }}
                "{{ .Col.ColumnName }}": {{ $short }}.{{ .Name }},
            {{- end }}
            })
		{{- else }}
			// sql query
			qb := sq.Update("{{ $table }}").SetMap(map[string]interface{}{
			{{- range .Fields }}
			    {{- if ne .Name $primaryKey.Name }}
			    "{{ .Col.ColumnName }}": {{ $short }}.{{ .Name }},
			    {{- end }}
            {{- end }}
            }).Where(sq.Eq{"{{ .PrimaryKey.Col.ColumnName }}": {{ $short }}.{{ .PrimaryKey.Name }}})
		{{- end }}
		query, args, err := qb.ToSql()
        if err != nil {
            return err
        }

        // run query
        _, err = {{ $shortRepo }}.db.Exec(query, args...)
        return err
	}
{{ else }}
	// Update statements omitted due to lack of fields other than primary key
{{ end }}

// Delete deletes the {{ lowerfirst .RepoName }} from the database.
func ({{ $shortRepo }} *{{ lowerfirst .RepoName }}) Delete({{ $short }} entities.{{ .Name }}) error {
	var err error
	{{ if gt ( len .PrimaryKeyFields ) 1 }}
		// sql query with composite primary key
		qb := sq.Delete("{{ $table }}").Where(sq.Eq{
        {{- range .PrimaryKeyFields }}
            "{{ .Col.ColumnName }}": {{ $short }}.{{ .Name }},
        {{- end }}
        })
	{{- else }}
		// sql query
		qb := sq.Delete("{{ $table }}").Where(sq.Eq{"{{ colname .PrimaryKey.Col}}": {{ $short }}.{{ .PrimaryKey.Name }}})
	{{- end }}

	query, args, err := qb.ToSql()
    if err != nil {
        return err
    }

    // run query
    _, err = {{ $shortRepo }}.db.Exec(query, args...)
    if err != nil {
        return err
    }

	return nil
}

func ({{ $shortRepo }} *{{ lowerfirst .RepoName }}) FindAll({{$short}}Filter *entities.{{ .Name }}Filter, pagination *entities.Pagination) ([]entities.{{ .Name }}, error) {
    qb := sq.Select("{{ $table }}")
    if {{$short}}Filter != nil {
        {{- range .Fields }}
            {{- if .Col.NotNull }}
            if ({{ $short }}Filter.{{ .Name }} != nil) {
                qb = qb.Where(sq.Eq{"{{ .Col.ColumnName }}": &{{ $short }}Filter.{{ .Name }}})
            }
            {{- else }}
                {{- if .Col.IsEnum }}
                if ({{ $short }}Filter.{{ .Name }} != nil) {
                    qb = qb.Where(sq.Eq{"{{ .Col.ColumnName }}": {{ $short }}Filter.{{ .Name }}})
                }
                {{- else }}
                if {{ $short }}Filter.{{ .Name }}.Valid {
                    qb = qb.Where(sq.Eq{"{{ .Col.ColumnName }}": {{ $short }}Filter.{{ .Name }}})
                }
                {{- end }}
            {{- end }}
        {{- end }}
    }
    if pagination != nil {
        if pagination.Page != nil && pagination.PerPage != nil {
            offset := uint64((*pagination.Page - 1) * *pagination.PerPage)
            qb = qb.Offset(offset).Limit(uint64(*pagination.PerPage))
        }
        if pagination.Sort != nil {
            orderStr := pagination.Sort.Field + " "
            if pagination.Sort.Direction != nil {
                orderStr += *pagination.Sort.Direction
            } else {
                orderStr += "ASC"
            }
            qb = qb.OrderBy(orderStr)
        }
    }

    var list []entities.{{ .Name }}
    query, args, err := qb.ToSql()
    if err != nil {
        return []entities.{{ .Name }}{}, err
    }
    err = {{ $shortRepo }}.db.Select(&list, query, args...)

    return list, err
}
{{- end }}

