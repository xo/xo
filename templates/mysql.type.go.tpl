{{- $short := (shortname .Name "err" "res" "sqlstr" "db" "XOLog") -}}
{{- $table := (schema .Table.TableName) -}}
{{- $tableVar := .Table }}
{{- $primaryKey := .PrimaryKey }}
{{- if .Comment -}}
// {{ .Comment }}
{{- else -}}
// {{ .Name }} represents a row from '{{ $table }}'.
{{- end }}
type {{ .Name }} struct {
{{- range .Fields }}
    {{- if and .Col.IsEnum (ne .Col.NotNull true) }}
        {{ .Name }} *{{ retype .Type }} `json:"{{ .Col.ColumnName }}" db:"{{ .Col.ColumnName }}"` // {{ .Col.ColumnName }}
    {{- else }}
	    {{ .Name }} {{ retype .Type }} `json:"{{ .Col.ColumnName }}" db:"{{ .Col.ColumnName }}"` // {{ .Col.ColumnName }}
    {{- end }}
{{- end }}
}

type {{ .Name }}Filter struct {
{{- range .Fields }}
	{{ .Name }} FilterOnField
{{- end }}
}

{{- $typeName := .Name }}

func (f *{{ $typeName }}Filter) NewFilter() interface{} {
    if f == nil {
        return &{{ $typeName }}Filter{}
    }
    return f
}

func (f *{{ $typeName}}Filter) IsNil() bool {
    return f == nil
}

{{- range .Fields }}
func (f *{{ $typeName }}Filter) Add{{ .Name }}(filterType FilterType, v interface{}) {
    f.{{ .Name }} = append(f.{{ .Name }}, map[FilterType]interface{}{filterType: v})
}
{{- end }}

func (f *{{ $typeName }}Filter) Hash() (string, error) {
    var err error
    var hash string
    if f == nil {
        return "", nil
    }
    h := md5.New()
    list := []struct{
        filter FilterOnField
        name string
    }{
        {{- range .Fields }}{ filter: f.{{.Name}}, name: "{{ .Name }}" },{{- end}}
    }
    for _, item := range list {
        hash, err = item.filter.Hash()
        if err != nil {
            return "", err
        }
        _,err = io.WriteString(h, item.name+" -> "+hash)
        if err != nil {
            return "", err
        }
    }
    return fmt.Sprintf("%x", h.Sum(nil)), nil
}

type {{ .Name }}Create struct {
{{- range .Fields }}
    {{- if and (or (ne .Col.ColumnName $primaryKey.Col.ColumnName) $tableVar.ManualPk) (ne .Col.ColumnName "created_at") (ne .Col.ColumnName "updated_at") }}
	{{ .Name }} {{- if .Col.NotNull}} {{ retype .Type }}{{ else }} {{retypeNull .Type}}{{- end}} `json:"{{ .Col.ColumnName }}" db:"{{ .Col.ColumnName }}"` // {{ .Col.ColumnName }}
	{{- end }}
{{- end }}
}

type {{ .Name }}Update struct {
{{- range .Fields }}
    {{- if and (or (ne .Col.ColumnName $primaryKey.Col.ColumnName) $tableVar.ManualPk) (ne .Col.ColumnName "created_at") (ne .Col.ColumnName "updated_at") }}
	{{ .Name }} *{{ retype .Type }} // {{ .Col.ColumnName }}
	{{- end }}
{{- end }}
}

type List{{ .Name }} struct {
    TotalCount int
    Data []{{ .Name }}
}

func (l *List{{ .Name }}) GetInterfaceItems() []interface{} {
    var arr []interface{}
	for _, item := range l.Data {
		arr = append(arr, item)
	}
	return arr
}

