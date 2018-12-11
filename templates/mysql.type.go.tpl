{{- $short := (shortname .Name "err" "res" "sqlstr" "db" "XOLog") -}}
{{- $table := (schema .Schema .Table.TableName) -}}
{{- if .Comment -}}
// {{ .Comment }}
{{- else -}}
// {{ .Name }} represents a row from '{{ $table }}'.
{{- end }}
type {{ .Name }} struct {
{{- range .Fields }}
	{{ .Name }} {{ retype .Type }} `json:"{{ .Col.ColumnName }}" db:"{{ .Col.ColumnName }}"` // {{ .Col.ColumnName }}
{{- end }}
}

type {{ .Name }}Filter struct {
{{- range .Fields }}
	{{ .Name }} {{ retypeNull .Type }} `json:"{{ .Col.ColumnName }}" db:"{{ .Col.ColumnName }}"` // {{ .Col.ColumnName }}
{{- end }}
}

