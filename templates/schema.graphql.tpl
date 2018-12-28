{{- $short := (shortname .Name "err" "res" "sqlstr" "db" "XOLog") -}}
{{- $table := (schema .Schema .Table.TableName) -}}
{{- $tableVar := .Table }}
{{- $primaryKey := .PrimaryKey }}
{{- $fkGroup := .ForeignKeyGroup }}
type {{ .Name }} {
{{- range .Fields }}
    {{ lowerfirst .Name }}: {{ retypegraphql .Type }} {{- if .Col.NotNull }}!{{- end }}
{{- end }}

{{- if $fkGroup }}

{{- range $fkGroup.ManyToOneKeys }}
    {{ lowerfirst .FuncName }}(filter: {{ .RefType.Name }}Filter): {{ .RefType.Name }}! @hasAccess(module: "{{ .RefType.Table.TableName }}", accessType: 1)
{{- end }}

{{- range $fkGroup.OneToManyKeys }}
    {{- if .IsUnique }}
    {{ lowerfirst .RevertFuncName }}(filter: {{ .Type.Name }}Filter): {{ .Type.Name }}! @hasAccess(module: "{{ .Type.Table.TableName }}", accessType: 1)
    {{- else }}
    {{ lowerfirst .RevertFuncName }}(filter: {{ .Type.Name }}Filter, pagination: Pagination): List{{ .Type.Name }}! @hasAccess(module: "{{ .Type.Table.TableName }}", accessType: 1)
    {{- end }}
{{- end }}
{{- end }}
}
input {{ .Name }}Filter {
{{- range .Fields }}
    {{ lowerfirst .Name }}: FilterOnField
{{- end }}
}

{{- if canhavecreatestruct .Fields $primaryKey }}
input {{ .Name }}Create {
{{- range .Fields }}
    {{- if and (or (ne .Col.ColumnName $primaryKey.Col.ColumnName) $tableVar.ManualPk) (ne .Col.ColumnName "created_at") (ne .Col.ColumnName "updated_at") }}
	{{ lowerfirst .Name }}: {{ retypegraphql .Type }}{{- if .Col.NotNull }}!{{- end }}
	{{- end }}
{{- end }}
}
{{- end }}

type List{{ .Name }} {
    totalCount: Int!
    data: [{{ .Name }}!]!
}
