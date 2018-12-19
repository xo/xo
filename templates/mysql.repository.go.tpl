{{- $shortRepo := (shortname .RepoName "err" "res" "sqlstr" "db" "XOLog") -}}
{{- $short := (shortname .Name "err" "res" "sqlstr" "db" "XOLog") -}}
{{- $table := (schema .Table.TableName) -}}
{{- $primaryKey := .PrimaryKey }}
{{- if .Comment -}}
// {{ .Comment }}
{{- else -}}

type I{{ .RepoName }} interface {
    Insert{{ .Name }}(ctx context.Context, {{ $short }} entities.{{ .Name }}Create) (*entities.{{ .Name }}, error)
    {{- if ne (fieldnamesmulti .Fields $short .PrimaryKeyFields) "" }}
    Update{{ .Name }}By{{ .Name }}Create(ctx context.Context, {{- range .PrimaryKeyFields }}{{ .Name }} {{ retype .Type }}{{- end }}, {{ $short }} entities.{{ .Name }}Create) (*entities.{{ .Name }}, error)
    Update{{ .Name }}(ctx context.Context, {{ $short }} entities.{{ .Name }}) (*entities.{{ .Name }}, error)
    {{- end }}
    Delete{{ .Name }}(ctx context.Context, {{ $short }} entities.{{ .Name }}) error
    FindAll{{ .Name }}(ctx context.Context, {{$short}}Filter *entities.{{ .Name }}Filter, pagination *entities.Pagination) (entities.List{{ .Name }}, error)
    {{- range .Indexes }}
    {{ .FuncName }}(ctx context.Context, {{ goparamlist .Fields false true }}) ({{ if not .Index.IsUnique }}[]{{ end }}*entities.{{ .Type.Name }}, error)
    {{- end }}
}

func New{{ .RepoName }}(db *sqlx.DB) I{{ .RepoName }} {
    return &{{ lowerfirst .RepoName }}{Db: db}
}

// {{ lowerfirst .RepoName }} represents a row from '{{ $table }}'.
{{- end }}
type {{ lowerfirst .RepoName }} struct {
    Db *sqlx.DB
}

{{ if .PrimaryKey }}

// Insert inserts the {{ .Name }}Create to the database.
func ({{ $shortRepo }} *{{ lowerfirst .RepoName }}) Insert{{ .Name }}(ctx context.Context, {{ $short }} entities.{{ .Name }}Create) (*entities.{{ .Name }}, error) {
	var err error

{{ if .Table.ManualPk  }}
	// sql insert query, primary key must be provided
	qb := sq.Insert("`{{ $table }}`").Columns({{ colnameswrap .Fields "CreatedAt" "UpdatedAt" }}).
	    Values({{ fieldnames .Fields $short "CreatedAt" "UpdatedAt" }})
    query, args, err := qb.ToSql()
	if err != nil {
	    return nil, err
	}

	// run query
	res, err = {{ $shortRepo }}.Db.Exec(query, args...)
	if err != nil {
		return nil, err
	}

{{ else }}
	// sql insert query, primary key provided by autoincrement
	qb := sq.Insert("`{{ $table }}`").Columns({{ colnameswrap .Fields .PrimaryKey.Name "CreatedAt" "UpdatedAt" }}).
	    Values({{ fieldnames .Fields $short .PrimaryKey.Name "CreatedAt" "UpdatedAt" }})
	query, args, err := qb.ToSql()
	if err != nil {
	    return nil, err
	}

	// run query
	res, err := {{ $shortRepo }}.Db.Exec(query, args...)
	if err != nil {
		return nil, err
	}
{{ end }}

    // retrieve id
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	new{{ $short }} := entities.{{ .Name }}{}

	err = {{ $shortRepo }}.Db.Get(&new{{ $short }}, "SELECT * FROM `{{ $table }}` WHERE `{{ .PrimaryKey.Col.ColumnName }}` = ?", id)

	return &new{{ $short }}, err
}

{{ if ne (fieldnamesmulti .Fields $short .PrimaryKeyFields) "" }}
	// Update updates the {{ .Name }}Create in the database.
	func ({{ $shortRepo }} *{{ lowerfirst .RepoName }}) Update{{ .Name }}By{{ .Name }}Create(ctx context.Context, {{- range .PrimaryKeyFields }}{{ .Name }} {{ retype .Type }}{{- end }}, {{ $short }} entities.{{ .Name }}Create) (*entities.{{ .Name }}, error) {
		var err error

		{{ if gt ( len .PrimaryKeyFields ) 1 }}
			// sql query with composite primary key
			qb := sq.Update("`{{ $table }}`").SetMap(map[string]interface{}{
            {{- range .Fields }}
                {{- if and (ne .Col.ColumnName "created_at") (ne .Col.ColumnName "updated_at") }}
                "`{{ .Col.ColumnName }}`": {{ $short }}.{{ .Name }},
                {{- end }}
            {{- end }}
            }).Where(sq.Eq{
            {{- range .PrimaryKeyFields }}
                "`{{ .Col.ColumnName }}`": .{{ .Name }},
            {{- end }}
            })
		{{- else }}
			// sql query
			qb := sq.Update("`{{ $table }}`").SetMap(map[string]interface{}{
			{{- range .Fields }}
			    {{- if ne .Name $primaryKey.Name }}
			    {{- if and (ne .Col.ColumnName "created_at") (ne .Col.ColumnName "updated_at") }}
			    "`{{ .Col.ColumnName }}`": {{ $short }}.{{ .Name }},
			    {{- end }}
			    {{- end }}
            {{- end }}
            }).Where(sq.Eq{"`{{ .PrimaryKey.Col.ColumnName }}`": {{ .PrimaryKey.Name }}})
		{{- end }}
		query, args, err := qb.ToSql()
        if err != nil {
            return nil, err
        }

        // run query
        _, err = {{ $shortRepo }}.Db.Exec(query, args...)
        if err != nil {
            return nil, err
        }

        selectQb := sq.Select("*").From("`{{ $table }}`")
        {{- if gt ( len .PrimaryKeyFields ) 1 }}
            selectQb = selectQb.Where(sq.Eq{
                {{- range .PrimaryKeyFields }}
                    "`{{ .Col.ColumnName }}`": .{{ .Name }},
                {{- end }}
                })
        {{- else }}
            selectQb = selectQb.Where(sq.Eq{"`{{ .PrimaryKey.Col.ColumnName }}`": {{ .PrimaryKey.Name }}})
        {{- end }}

        query, args, err = selectQb.ToSql()
        if err != nil {
            return nil, err
        }

        result := entities.{{ .Name }}{}
        err = {{ $shortRepo }}.Db.Get(&result, query, args...)
        return &result, err
	}

    // Update updates the {{ .Name }} in the database.
	func ({{ $shortRepo }} *{{ lowerfirst .RepoName }}) Update{{ .Name }}(ctx context.Context, {{ $short }} entities.{{ .Name }}) (*entities.{{ .Name }}, error) {
    		var err error

    		{{ if gt ( len .PrimaryKeyFields ) 1 }}
    			// sql query with composite primary key
    			qb := sq.Update("`{{ $table }}`").SetMap(map[string]interface{}{
                {{- range .Fields }}
                    {{- if and (ne .Col.ColumnName "created_at") (ne .Col.ColumnName "updated_at") }}
                    "`{{ .Col.ColumnName }}`": {{ $short }}.{{ .Name }},
                    {{- end }}
                {{- end }}
                }).Where(sq.Eq{
                {{- range .PrimaryKeyFields }}
                    "`{{ .Col.ColumnName }}`": {{ $short}}.{{ .Name }},
                {{- end }}
                })
    		{{- else }}
    			// sql query
    			qb := sq.Update("`{{ $table }}`").SetMap(map[string]interface{}{
    			{{- range .Fields }}
    			    {{- if ne .Name $primaryKey.Name }}
    			    {{- if and (ne .Col.ColumnName "created_at") (ne .Col.ColumnName "updated_at") }}
    			    "`{{ .Col.ColumnName }}`": {{ $short }}.{{ .Name }},
    			    {{- end }}
    			    {{- end }}
                {{- end }}
                }).Where(sq.Eq{"`{{ .PrimaryKey.Col.ColumnName }}`": {{ $short}}.{{ .PrimaryKey.Name }}})
    		{{- end }}
    		query, args, err := qb.ToSql()
            if err != nil {
                return nil, err
            }

            // run query
            _, err = {{ $shortRepo }}.Db.Exec(query, args...)
            if err != nil {
                return nil, err
            }

            selectQb := sq.Select("*").From("`{{ $table }}`")
            {{- if gt ( len .PrimaryKeyFields ) 1 }}
                selectQb = selectQb.Where(sq.Eq{
                    {{- range .PrimaryKeyFields }}
                        "`{{ .Col.ColumnName }}`": {{ $short}}.{{ .Name }},
                    {{- end }}
                    })
            {{- else }}
                selectQb = selectQb.Where(sq.Eq{"`{{ .PrimaryKey.Col.ColumnName }}`": {{ $short}}.{{ .PrimaryKey.Name }}})
            {{- end }}

            query, args, err = selectQb.ToSql()
            if err != nil {
                return nil, err
            }

            result := entities.{{ .Name }}{}
            err = {{ $shortRepo }}.Db.Get(&result, query, args...)
            return &result, err
    	}
{{ else }}
	// Update statements omitted due to lack of fields other than primary key
{{ end }}

// Delete deletes the {{ .Name }} from the database.
func ({{ $shortRepo }} *{{ lowerfirst .RepoName }}) Delete{{ .Name }}(ctx context.Context, {{ $short }} entities.{{ .Name }}) error {
	var err error
	{{ if gt ( len .PrimaryKeyFields ) 1 }}
		// sql query with composite primary key
		qb := sq.Delete("`{{ $table }}`").Where(sq.Eq{
        {{- range .PrimaryKeyFields }}
            "`{{ .Col.ColumnName }}`": {{ $short }}.{{ .Name }},
        {{- end }}
        })
	{{- else }}
		// sql query
		qb := sq.Delete("`{{ $table }}`").Where(sq.Eq{"`{{ colname .PrimaryKey.Col}}`": {{ $short }}.{{ .PrimaryKey.Name }}})
	{{- end }}

	query, args, err := qb.ToSql()
    if err != nil {
        return err
    }

    // run query
    _, err = {{ $shortRepo }}.Db.Exec(query, args...)
    return err
}

func ({{ $shortRepo }} *{{ lowerfirst .RepoName }}) findAll{{ .Name }}Query(ctx context.Context, {{$short}}Filter *entities.{{ .Name }}Filter, pagination *entities.Pagination, fields string) (string, []interface{},  error) {
    qb := sq.Select(fields).From("`{{ $table }}`")
    if {{$short}}Filter != nil {
        {{- range .Fields }}
            {{- if .Col.NotNull }}
            if ({{ $short }}Filter.{{ .Name }} != nil) {
                qb = qb.Where(sq.Eq{"`{{ .Col.ColumnName }}`": &{{ $short }}Filter.{{ .Name }}})
            }
            {{- else }}
                {{- if .Col.IsEnum }}
                if ({{ $short }}Filter.{{ .Name }} != nil) {
                    qb = qb.Where(sq.Eq{"`{{ .Col.ColumnName }}`": {{ $short }}Filter.{{ .Name }}})
                }
                {{- else }}
                if {{ $short }}Filter.{{ .Name }}.Valid {
                    qb = qb.Where(sq.Eq{"`{{ .Col.ColumnName }}`": {{ $short }}Filter.{{ .Name }}})
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

    return qb.ToSql()
}

func ({{ $shortRepo }} *{{ lowerfirst .RepoName }}) FindAll{{ .Name }}(ctx context.Context, filter *entities.{{ .Name }}Filter, pagination *entities.Pagination) (list entities.List{{ .Name }}, err error) {
    query, args, err := {{ $shortRepo }}.findAll{{ .Name }}Query(ctx, filter, pagination, "*")
    if err != nil {
        return list, err
    }
    err = {{ $shortRepo }}.Db.Select(&list.Data, query, args...)

    if err != nil {
        return list, err
    }

    var listMeta entities.ListMetadata
    query, args, err = {{ $shortRepo }}.findAll{{ .Name }}Query(ctx, filter, pagination, "COUNT(*) AS count")
    if err != nil {
        return list, err
    }
    err = {{ $shortRepo }}.Db.Get(&listMeta, query, args...)

    list.TotalCount = listMeta.Count

    return list, err
}
{{- end }}

