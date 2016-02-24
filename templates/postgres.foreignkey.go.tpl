// {{ .RefType.Name }} returns the {{ .RefType.Name }} associated with the {{ .Type.Name }}'s {{ .Field.Name }} ({{ .Field.Col.ColumnName }}).
//
// Generated from {{ .ForeignKey.ForeignKeyName }}.{{ $t := (print (shortname .Type.Name) "." .Field.Name (nulltypeext .Field.Name)) }}
func ({{ shortname .Type.Name }} *{{ .Type.Name }}) {{ .RefType.Name }}(db XODB) (*{{ .RefType.Name }}, error) {
	return {{ .RefType.Name }}By{{ .RefField.Name }}(db, {{ if ne .Field.Type .RefField.Type }}{{ .RefField.Type }}({{ $t }}){{ else }}{{ $t }}{{ end }})
}

