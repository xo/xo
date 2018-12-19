schema:
  {{- range $k, $v := allschemas }}
  - schema/{{ $v }}.graphql
  {{- end }}
  - schema/query.graphql
  - schema/mutation.graphql
exec:
  filename: generated.go
model:
  filename: models_gen.go
resolver:
  filename: tmp/resolver.go
  type: Resolver
models:
  Pagination:
    model: {{ entitiespkg }}.Pagination
  Sort:
    model: {{ entitiespkg }}.Sort
  ListMetadata:
    model: {{ entitiespkg }}.ListMetadata
  Datetime:
    model: {{ entitiespkg }}.Datetime
  IntBool:
    model: {{ entitiespkg }}.IntBool
  NullInt64:
    model: {{ entitiespkg }}.NullInt64
  NullFloat64:
    model: {{ entitiespkg }}.NullFloat64
  NullBool:
    model: {{ entitiespkg }}.NullBool
  NullString:
    model: {{ entitiespkg }}.NullString
  NullTime:
    model: {{ entitiespkg }}.NullTime
  Map:
    model: {{ entitiespkg }}.Map
