{{- $query := .Data -}}
{{- $table := (schema $query.Table.TableName) -}}
{{- if $query.Comment -}}
// {{ $query.Comment }}
{{- else -}}
// {{ $query.Name }} represents a row from '{{ $table }}'.
{{- end }}
type {{ $query.Name }} struct {
{{ range $query.Fields -}}
	{{ .Name }} {{ retype .Type }} {{ fieldtag . }} // {{ .Col.ColumnName }}
{{ end -}}
}

