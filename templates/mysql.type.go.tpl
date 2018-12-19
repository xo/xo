{{- $short := (shortname .Name "err" "res" "sqlstr" "db" "XOLog") -}}
{{- $table := (schema .Table.TableName) -}}
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

type List{{ .Name }} struct {
    TotalCount int
    Data []{{ .Name }}
}

