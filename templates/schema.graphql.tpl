{{- $short := (shortname .Name "err" "res" "sqlstr" "db" "XOLog") -}}
{{- $table := (schema .Schema .Table.TableName) -}}
{{- $tableVar := .Table }}
{{- $primaryKey := .PrimaryKey }}
type {{ .Name }} {
{{- range .Fields }}
    {{ lowerfirst .Name }}: {{ retypegraphql .Type }} {{- if .Col.NotNull }}!{{- end }}
{{- end }}
}

input {{ .Name }}Filter {
{{- range .Fields }}
    {{ lowerfirst .Name }}: {{ retypegraphql .Type }}
{{- end }}
}

input {{ .Name }}Create {
{{- range .Fields }}
    {{- if and (or (ne .Col.ColumnName $primaryKey.Col.ColumnName) $tableVar.ManualPk) (ne .Col.ColumnName "created_at") (ne .Col.ColumnName "updated_at") }}
	{{ lowerfirst .Name }}: {{ retypegraphql .Type }}{{- if .Col.NotNull }}!{{- end }}
	{{- end }}
{{- end }}
}

type List{{ .Name }} {
    totalCount: Int!
    data: [{{ .Name }}!]!
}
