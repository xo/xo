{{- $q := .Data -}}
{{- if $q.Comment -}}
// {{ $q.Comment | eval $q.Name }}
{{- else -}}
// {{ $q.Name }} represents a row from '{{ schema $q.Table.TableName }}'.
{{- end }}
type {{ $q.Name }} struct {
{{ range $q.Fields -}}
    {{ field . }}
{{ end -}}
}

