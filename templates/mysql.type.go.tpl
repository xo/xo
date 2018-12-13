{{- $short := (shortname .Name "err" "res" "sqlstr" "db" "XOLog") -}}
{{- $table := (schema .Schema .Table.TableName) -}}
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

// GraphQL
/**
scalar Datetime

type {{ .Name }} {
{{- range .Fields }}
    {{ .Name }}: {{ retypegraphql .Type }} {{- if .Col.NotNull }}!{{- end }}
{{- end }}
}

type {{ .Name }}Filter {
{{- range .Fields }}
    {{ .Name }}: {{ retypegraphql .Type }}
{{- end }}
}

input Sort {
    field: String!
    direction: String
}

input Pagination {
    page:           Int
    perPage:        Int
    sort:           Sort
}

type Query {
    all{{ .Name }}(filter: {{ .Name }}Filter, pagination: Pagination): [{{ .Name }}!]!
}
*/

