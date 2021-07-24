{{ $s := .Data -}}
{{ if driver "mysql" }}{{ addEnums $s.Enums }}{{ end -}}
-- SQL Schema template for the {{ $s.Name }} schema.
-- Generated on {{ time }} by xo.
{{ if driver "postgres" -}}
{{ range $e := $s.Enums }}
CREATE TYPE {{ escType $e.Name }} AS ENUM (
{{- range $i, $v := $e.Values }}
    {{ strLiteral $v.Name }}{{ comma $i $e.Values }}
{{- end }}
);
{{ end }}
{{- end }}

{{- /* Tables */ -}}

{{- range $t := $s.Tables }}
CREATE TABLE {{ escType $t.Name }} (
{{- range $i, $c := $t.Columns }}
    {{ coldef $t $c }}{{ comma $i $t.Columns }}
{{- end }}

{{- /* Primary keys and Unique indexes */ -}}
{{- range $idx := $t.Indexes -}}
    {{- /* Do not handle sequence primary keys for sqlite3 */ -}}
    {{- $valid := true -}}
    {{- if and (driver "sqlite3") (index $idx.Fields 0).IsSequence -}}
        {{ $valid = false }}
    {{- end -}}

    {{- if and $valid $idx.IsPrimary }},
    {{ constraint $idx.Name -}} PRIMARY KEY ({{ fields $idx.Fields  }})
    {{- else if and $valid $idx.IsUnique }},
    {{ constraint $idx.Name -}} UNIQUE ({{ fields $idx.Fields }})
    {{- end }}
{{- end }}

{{- /* Composite Foreign Keys */ -}}
{{- range $fk := $t.ForeignKeys -}}
    {{- if gt (len $fk.Fields) 1 }},
    {{ constraint $fk.Name -}} FOREIGN KEY ({{ fields $fk.Fields }}) REFERENCES {{ escType $fk.RefTable }} ({{ fields $fk.RefFields }})
    {{- end }}
{{- end }}
){{ engine }};

{{ range $idx := $t.Indexes -}}
{{ if not (or $idx.IsPrimary $idx.IsUnique) -}}
{{ indexdef $t $idx }};
{{ end }}
{{- end }}
{{- end }}

{{- range $v := $s.Views }}
{{ viewdef $v }}
{{ end }}

{{- /* Procs and Functions */ -}}
{{- range $p := $s.Procs }}
{{ procdef $p }};
{{ end }}
