{{- $s := .Data -}}
-- Generated by xo for the {{ $s.Name }} schema.
{{ if and $s.Enums (driver "postgres") }}
{{ range $e := $s.Enums }}
-- enum {{ $e.Name }}
CREATE TYPE {{ esc $e.Name }} AS ENUM (
{{- range $i, $v := $e.Values }}
  {{ literal $v.Name }}{{ comma $i $e.Values }}
{{- end }}
);
{{ end -}}
{{- end -}}
{{- if $s.Tables }}
{{- range $t := $s.Tables }}
-- table {{ $t.Name }}
CREATE TABLE {{ esc $t.Name }} (
{{- range $i, $c := $t.Columns }}
  {{ coldef $t $c }}{{ comma $i $t.Columns }}
{{- end -}}
{{- range $idx := $t.Indexes -}}{{- if isEndConstraint $idx }},
  {{ constraint $idx.Name -}} {{ if $idx.IsPrimary }}PRIMARY KEY{{ else }}UNIQUE{{ end }} ({{ fields $idx.Fields }})
{{- end -}}{{- end -}}
{{- range $fk := $t.ForeignKeys -}}{{- if gt (len $fk.Fields) 1 }},
  {{ constraint $fk.Name -}} FOREIGN KEY ({{ fields $fk.Fields }}) REFERENCES {{ esc $fk.RefTable }} ({{ fields $fk.RefFields }})
{{- end -}}{{- end }}
){{ engine }};
{{- if $t.Indexes }}
{{ range $idx := $t.Indexes }}{{ if not (or $idx.IsPrimary $idx.IsUnique) }}
-- index {{ $idx.Name }}
CREATE INDEX {{ esc $idx.Name }} ON {{ esc $t.Name }} ({{ fields $idx.Fields }});
{{ end -}}{{- end -}}{{- end }}
{{ end -}}
{{- end -}}
{{- if $s.Views }}
{{- range $v := $s.Views }}
-- view {{ $v.Name }}
{{ viewdef $v }};
{{ end }}
{{ end -}}
{{- if $s.Procs }}
{{- range $p := $s.Procs }}
-- {{ $p.Type }} {{ $p.Name }}
{{ procdef $p }};
{{ end -}}
{{ end -}}
