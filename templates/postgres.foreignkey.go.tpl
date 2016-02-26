// {{ .RefType.Name }} returns the {{ .RefType.Name }} associated with the {{ .Type.Name }}'s {{ .Field.Name }} ({{ .Field.Col.ColumnName }}).
//
// Generated from foreign key '{{ .ForeignKey.ForeignKeyName }}'.
func ({{ shortname .Type.Name }} *{{ .Type.Name }}) {{ .RefType.Name }}(db XODB) (*{{ .RefType.Name }}, error) {
	return {{ .RefType.Name }}By{{ .RefField.Name }}(db, {{ convext (shortname .Type.Name) .Field .RefField }})
}

