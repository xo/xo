// {{ .Type }} returns the {{ .RefType }} associated with the {{ .Type }}'s {{ .Field }} ({{ .ColumnName }}).
func (t *{{ .Type }}) {{ .RefType }}(db XODB) (*{{ .RefType }}, error) {
    return {{ .RefType }}By{{ .RefField }}(db, t.{{ .Field }})
}

