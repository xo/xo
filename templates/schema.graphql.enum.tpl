{{- $type := .Name -}}

enum {{ $type }} {
{{- range .Values }}
    {{ .Val.EnumValue }}
{{- end }}
}
