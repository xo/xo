// {{ .Type }} returns the {{ .RefType }} associated with the {{ .Type }}'s {{ .Field }} ({{ .ColumnName }}).
func ({{ shortname .Type }} *{{ .Type }}) {{ .RefType }}(db XODB) (*{{ .RefType }}, error) {
	return {{ .RefType }}By{{ .RefField }}(db, {{ shortname .Type }}.{{ .Field }})
}

