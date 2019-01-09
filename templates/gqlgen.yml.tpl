schema:
  {{- range $k, $v := allschemas }}
  - schema/{{ $v }}.graphql
  {{- end }}
  - schema/query.graphql
  - schema/mutation.graphql
exec:
  filename: generated.go
model:
  filename: ../entities/models_gen.go
resolver:
  filename: tmp/resolver.go
  type: Resolver
models:
  Pagination:
    model: {{ entitiespkg }}.Pagination
  Sort:
    model: {{ entitiespkg }}.Sort
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
  FilterOnField:
    model: {{ entitiespkg }}.FilterOnField
  Point:
    model github.com/paulmach/go.geo.Point
