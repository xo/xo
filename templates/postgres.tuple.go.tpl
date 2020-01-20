type {{ .Name }} struct {
{{- range .Fields }}
	{{ .Name }} {{ retype .Type }} `json:"{{ .Name }}"` // {{ .Name }}
{{- end }}
}
