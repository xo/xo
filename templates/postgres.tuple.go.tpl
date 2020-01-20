type {{ .Name }} struct {
{{- range .Fields }}
	{{ .Name }} {{ retype .Type }} `json:"{{ .FieldName }}"` // {{ .FieldName }}
{{- end }}
}
