// {{ .Type }} represents a row from {{ if .TableSchema }}{{ .TableSchema }}.{{ end }}{{ .TableName }}.
type {{ .Type }} struct {
{{- range .Fields }}
	{{ .Field }} {{ retype .GoType }}{{ if .Tag }} `{{ .Tag }}`{{ end }} // {{ .ColumnName }}
{{- end }}
}

// Save saves the {{ .Type }} to the database.
func (c *{{ .Type }}) Save(db *sql.DB) error {
	return nil
}

// Delete deletes the {{ .Type }} from the database.
func (c *{{ .Type }}) Delete(db *sql.DB) error {
	return nil
}
