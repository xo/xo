{{- $type := .Data -}}
{{- $short := (shortname $type.Name "err" "res" "sqlstr" "db" "logf") -}}
{{- $table := (schema $type.Table.TableName) -}}
{{- if $type.Comment -}}
// {{ $type.Comment }}
{{- else -}}
// {{ $type.Name }} represents a row from '{{ $table }}'.
{{- end }}
type {{ $type.Name }}_ struct {
{{ range $type.Fields -}}
	{{ .Name }} {{ retype .Type }} {{ fieldtag . }}
{{ end -}}
}

