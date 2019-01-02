{{- $shortRepo := (shortname .Name "err" "res" "sqlstr" "db" "XOLog") -}}
{{- $name := .Name }}
{{- $typeName := .TypeName }}

type I{{ .Name }} interface {
{{- range .ManyToOneKeys }}
{{- if ne .CallFuncName "" }}
    {{ .FuncName }}(ctx context.Context, obj *entities.{{ .Type.Name }}, filter *entities.{{ .RefType.Name }}Filter) (entities.{{ .RefType.Name }}, error)
{{- end }}
{{- end }}

{{- range .OneToManyKeys }}
{{- if ne .RevertCallFuncName "" }}
{{- if .IsUnique }}
    {{ .RevertFuncName }}(ctx context.Context, obj *entities.{{ .RefType.Name }}, filter *entities.{{ .Type.Name }}Filter) (entities.{{ .Type.Name }}, error)
{{- else }}
    {{ .RevertFuncName }}(ctx context.Context, obj *entities.{{ .RefType.Name }}, filter *entities.{{ .Type.Name }}Filter, pagination *entities.Pagination) (entities.List{{ .Type.Name }}, error)
{{- end }}
{{- end }}
{{- end }}
}

type {{ .Name }} struct {
    {{- range $k, $v := .DependOnRepo }}
    {{ $v}} I{{ $v }}
    {{- end }}
}

var New{{ .Name }} = wire.NewSet({{ .Name }}{}, wire.Bind(new(I{{ .Name }}), new({{ .Name }})))

{{- range .ManyToOneKeys }}
{{- if ne .CallFuncName "" }}
func ({{ $shortRepo }} *{{ $name }}) {{ .FuncName }}(ctx context.Context, obj *entities.{{ .Type.Name }}, filter *entities.{{ .RefType.Name }}Filter) (entities.{{ .RefType.Name }}, error) {
    if obj ==  nil {
        return entities.{{ .RefType.Name }}{}, nil
    }
    {{- if eq .Field.Type .RefField.Type }}
    return {{ $shortRepo}}.{{ .RefType.RepoName}}.{{ .CallFuncName }}(ctx, obj.{{ .Field.Name }}, filter)
    {{- else }}
    return {{ $shortRepo}}.{{ .RefType.RepoName}}.{{ .CallFuncName }}(ctx, {{ convertToNonNull (print "obj." .Field.Name) .Field.Type }}, filter)
    {{- end }}
}
{{- end }}
{{ end }}

{{- range .OneToManyKeys }}
{{- if ne .RevertCallFuncName "" }}
{{- if .IsUnique }}
func ({{ $shortRepo }} *{{ $name }}) {{ .RevertFuncName }}(ctx context.Context, obj *entities.{{ .RefType.Name }}, filter *entities.{{ .Type.Name }}Filter) (entities.{{ .Type.Name }}, error) {
    if obj ==  nil {
        return entities.{{ .Type.Name }}{}, nil
    }
    {{- if eq .Field.Type .RefField.Type }}
    return {{ $shortRepo }}.{{ .Type.RepoName}}.{{ .RevertCallFuncName }}(ctx, obj.{{ .RefField.Name }}, filter)
    {{- else }}
    return {{ $shortRepo }}.{{ .Type.RepoName}}.{{ .RevertCallFuncName }}(ctx, {{convertToNull (print "obj." .RefField.Name) .RefField.Type}}, filter)
    {{- end }}
}
{{- else }}
func ({{ $shortRepo }} *{{ $name }}) {{ .RevertFuncName }}(ctx context.Context, obj *entities.{{ .RefType.Name }}, filter *entities.{{ .Type.Name }}Filter, pagination *entities.Pagination) (entities.List{{ .Type.Name }}, error) {
    if obj ==  nil {
        return entities.List{{ .Type.Name }}{}, nil
    }
    {{- if eq .Field.Type .RefField.Type }}
    return {{ $shortRepo }}.{{ .Type.RepoName}}.{{ .RevertCallFuncName }}(ctx, obj.{{ .RefField.Name }}, filter, pagination)
    {{- else }}
    return {{ $shortRepo }}.{{ .Type.RepoName}}.{{ .RevertCallFuncName }}(ctx, {{convertToNull (print "obj." .RefField.Name) .RefField.Type}}, filter, pagination)
    {{- end }}
}
{{- end }}
{{- end }}
{{ end }}