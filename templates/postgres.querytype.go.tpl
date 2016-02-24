{{- if .Comment -}}
// {{ .Comment }}
{{- else -}}
// {{ .Name }} represents a row from {{ schema .Schema .Table.TableName }}.
{{- end }}
type {{ .Name }} struct {
{{- range .Fields }}
	{{ .Name }} {{ retype .Type }} // {{ .Col.ColumnName }}
{{- end }}
}

