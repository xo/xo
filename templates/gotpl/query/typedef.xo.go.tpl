{{- $q := .Data -}}
{{- if $q.Comment -}}
// {{ $q.Comment | eval $q.GoName }}
{{- else -}}
// {{ $q.GoName }} represents a row from '{{ schema $q.SQLName }}'.
{{- end }}
type {{ $q.GoName }} struct {
{{ range $q.Fields -}}
    {{ field . }}
{{ end -}}
}

