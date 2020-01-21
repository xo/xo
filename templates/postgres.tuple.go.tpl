type {{ .Name }} struct {
{{- range .Fields }}
	{{ .Name }} {{ retype .Type }} `db:"{{ .FieldName }}"` // {{ .FieldName }}
{{- end }}
}
