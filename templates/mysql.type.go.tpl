{{- $short := (shortname .Name "err" "res" "sqlstr" "db" "XOLog") -}}
{{- $table := (schema .Schema .Table.TableName) -}}
{{- $tableVar := .Table }}
{{- $primaryKey := .PrimaryKey }}
{{- if .Comment -}}
// {{ .Comment }}
{{- else -}}
// {{ .Name }} represents a row from '{{ $table }}'.
{{- end }}
type {{ .Name }} struct {
{{- range .Fields }}
    {{- if and .Col.IsEnum (ne .Col.NotNull true) }}
        {{ .Name }} *{{ retype .Type }} `json:"{{ .Col.ColumnName }}" db:"{{ .Col.ColumnName }}"` // {{ .Col.ColumnName }}
    {{- else }}
	    {{ .Name }} {{ retype .Type }} `json:"{{ .Col.ColumnName }}" db:"{{ .Col.ColumnName }}"` // {{ .Col.ColumnName }}
    {{- end }}
{{- end }}
}

type {{ .Name }}Filter struct {
{{- range .Fields }}
	{{ .Name }} {{ retypeNull .Type }} `json:"{{ .Col.ColumnName }}" db:"{{ .Col.ColumnName }}"` // {{ .Col.ColumnName }}
{{- end }}
}

type {{ .Name }}Create struct {
{{- range .Fields }}
    {{- if and (or (ne .Col.ColumnName $primaryKey.Col.ColumnName) $tableVar.ManualPk) (ne .Col.ColumnName "created_at") (ne .Col.ColumnName "updated_at") }}
	{{ .Name }} {{- if .Col.NotNull}} {{ retype .Type }}{{ else }} {{retypeNull .Type}}{{- end}} `json:"{{ .Col.ColumnName }}" db:"{{ .Col.ColumnName }}"` // {{ .Col.ColumnName }}
	{{- end }}
{{- end }}
}

// GraphQL
/**
type {{ .Name }} {
{{- range .Fields }}
    {{ lowerfirst .Name }}: {{ retypegraphql .Type }} {{- if .Col.NotNull }}!{{- end }}
{{- end }}
}

input {{ .Name }}Filter {
{{- range .Fields }}
    {{ lowerfirst .Name }}: {{ retypegraphql .Type }}
{{- end }}
}

input {{ .Name }}Create {
{{- range .Fields }}
    {{- if and (or (ne .Col.ColumnName $primaryKey.Col.ColumnName) $tableVar.ManualPk) (ne .Col.ColumnName "created_at") (ne .Col.ColumnName "updated_at") }}
	{{ lowerfirst .Name }}: {{ retypegraphql .Type }}{{- if .Col.NotNull }}!{{- end }}
	{{- end }}
{{- end }}
}

type Query {
    all{{ .Name }}(filter: {{ .Name }}Filter, pagination: Pagination): [{{ .Name }}!]!
}

type Mutation {
    create{{ .Name }}(data: {{ .Name }}Create!): {{.Name}}
    update{{ .Name }}(id: Int!, data: {{ .Name }}Create!): {{.Name}}
}
*/

