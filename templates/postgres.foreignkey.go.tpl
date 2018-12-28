{{- $shortRepo := (shortname .Name "err" "res" "sqlstr" "db" "XOLog") -}}
{{- $name := .Name }}
{{- $typeName := .TypeName }}

type I{{ .Name }} interface {

}

type {{ lowerfirst .Name }} struct {
    {{- range $k, $v := .DependOnRepo }}
    {{ lowerfirst $v}} I{{ $v }}
    {{- end }}
}

{{- range .ManyToOneKeys }}
{{- if ne .CallFuncName "" }}
func ({{ $shortRepo }} {{ lowerfirst $name }}) {{ .FuncName }}(ctx context.Context, obj *entities.{{ .Type.Name }}, filter *entities.{{ .RefType.Name }}Filter) (entities.{{ .RefType.Name }}, error) {
    if obj ==  nil {
        return entities.{{ .RefType.Name }}{}, nil
    }
    return {{ $shortRepo}}.{{ lowerfirst .RefType.RepoName}}.{{ .CallFuncName }}(ctx, obj.{{ .Field.Name }}, filter)
}
{{- end }}
{{ end }}

{{- range .OneToManyKeys }}
{{- if ne .RevertCallFuncName "" }}
{{- if .IsUnique }}
func ({{ $shortRepo }} {{ lowerfirst $name }}) {{ .RevertFuncName }}(ctx context.Context, obj *entities.{{ .RefType.Name }}, filter *entities.{{ .Type.Name }}Filter) (entities.{{ .Type.Name }}, error) {
    if obj ==  nil {
        return entities.{{ .Type.Name }}{}, nil
    }
    return {{ $shortRepo }}.{{ lowerfirst .Type.RepoName}}.{{ .RevertCallFuncName }}(ctx, obj.{{ .RefField.Name }}, filter)
}
{{- else }}
func ({{ $shortRepo }} {{ lowerfirst $name }}) {{ .RevertFuncName }}(ctx context.Context, obj *entities.{{ .RefType.Name }}, filter *entities.{{ .Type.Name }}Filter, pagination *entities.Pagination) (entities.List{{ .Type.Name }}, error) {
    if obj ==  nil {
        return entities.List{{ .Type.Name }}{}, nil
    }
    return {{ $shortRepo }}.{{ lowerfirst .Type.RepoName}}.{{ .RevertCallFuncName }}(ctx, obj.{{ .RefField.Name }}, filter, pagination)
}
{{- end }}
{{- end }}
{{ end }}