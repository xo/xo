// {{ .Type }} returns the {{ .RefType }} associated with the {{ .Type }}'s {{ .Field }} ({{ .ColumnName }}).
//
// Generated from {{ .ForeignKeyName }}. {{ $t := (print (shortname .Type) "." .Field (nulltypeext .FieldType)) }}
func ({{ shortname .Type }} *{{ .Type }}) {{ .RefType }}(db XODB) (*{{ .RefType }}, error) {
	return {{ .RefType }}By{{ .RefField }}(db, {{ if ne .FieldType .RefFieldType }}{{ .RefFieldType }}({{ $t }}){{ else }}{{ $t }}{{ end }})
}

