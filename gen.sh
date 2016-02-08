#!/bin/bash

DEST=$1

if [ -z "$DEST" ]; then
  DEST=x
fi

set -ex

if [[ "$DEST" != "models" ]]; then
  go generate
  go build
fi

rm -rf $DEST
mkdir -p $DEST

# enum query
cat << ENDSQL | ./xo pgsql://xodb:xodb@localhost/xodb -N -M -B -T Enum --comment='Enum represents a PostgreSQL enum value.' -o $DEST
SELECT
  t.typname::varchar AS enum_type,
  e.enumlabel::varchar AS enum_value,
  e.enumsortorder::integer AS const_value,
  ''::varchar AS type,
  ''::varchar AS value,
  ''::varchar AS comment
FROM pg_type t
  LEFT JOIN pg_namespace n ON n.oid = t.typnamespace
  JOIN pg_enum e ON t.oid = e.enumtypid
WHERE n.nspname = %%schema string%%
ENDSQL

# column query
cat << ENDSQL | ./xo pgsql://xodb:xodb@localhost/xodb -N -M -B -T Column --comment='Column represents PostgreSQL class (ie, table, view, etc) attributes.' -o $DEST
SELECT
  a.attname::varchar AS column_name,
  c.relname::varchar AS table_name,
  format_type(a.atttypid, a.atttypmod)::varchar AS data_type,
  a.attnum::integer AS field_ordinal,
  (NOT a.attnotnull)::boolean AS is_nullable,
  COALESCE(i.oid <> 0, false)::boolean AS is_index,
  COALESCE(ct.contype = 'u' OR ct.contype = 'p', false)::boolean AS is_unique,
  COALESCE(ct.contype = 'p', false)::boolean AS is_primary_key,
  COALESCE(cf.contype = 'f', false)::boolean AS is_foreign_key,
  COALESCE(i.relname, '')::varchar AS index_name,
  COALESCE(cf.conname, '')::varchar AS foreign_index_name,
  a.atthasdef::boolean AS has_default,
  COALESCE(pg_get_expr(ad.adbin, ad.adrelid), '')::varchar AS default_value,
  ''::varchar AS field,
  ''::varchar AS type,
  ''::varchar AS nil_type,
  ''::varchar AS tag,
  0::integer AS len,
  ''::varchar AS comment
FROM pg_attribute a
  JOIN ONLY pg_class c ON c.oid = a.attrelid
  JOIN ONLY pg_namespace n ON n.oid = c.relnamespace
  LEFT JOIN pg_constraint ct ON ct.conrelid = c.oid AND a.attnum = ANY(ct.conkey) AND ct.contype IN('p', 'u')
  LEFT JOIN pg_constraint cf ON cf.conrelid = c.oid AND a.attnum = ANY(cf.conkey) AND cf.contype IN('f')
  LEFT JOIN pg_attrdef ad ON ad.adrelid = c.oid AND ad.adnum = a.attnum
  LEFT JOIN pg_index ix ON a.attnum = ANY(ix.indkey) AND c.oid = a.attrelid AND c.oid = ix.indrelid
  LEFT JOIN pg_class i ON i.oid = ix.indexrelid
WHERE c.relkind = %%relkind string%% AND a.attnum > 0 AND n.nspname = %%schema string%%
ORDER BY c.relname, a.attnum
ENDSQL

# proc query
cat << ENDSQL | ./xo pgsql://xodb:xodb@localhost/xodb -N -M -B -T Proc --comment='Proc represents a PostgreSQL stored procedure.' -o $DEST
SELECT
  p.proname::varchar AS proc_name,
  oidvectortypes(p.proargtypes)::varchar AS parameter_types,
  pg_get_function_result(p.oid)::varchar AS return_type,
  ''::varchar AS comment
FROM pg_proc p
  INNER JOIN pg_namespace n ON p.pronamespace = n.oid
WHERE n.nspname = %%schema string%%
ENDSQL

# foreign key query
cat << ENDSQL | ./xo pgsql://xodb:xodb@localhost/xodb -N -M -B -T ForeignKey --comment='ForeignKey represents a PostgreSQL foreign key.' -o $DEST
SELECT
  r.conname::varchar AS foreign_key_name,
  a.relname::varchar AS table_name,
  b.attname::varchar AS column_name,
  i.relname::varchar AS ref_index_name,
  c.relname::varchar AS ref_table_name,
  d.attname::varchar AS ref_column_name,
  ''::varchar AS comment
FROM pg_constraint r
  JOIN ONLY pg_class a ON a.oid = r.conrelid
  JOIN ONLY pg_attribute b ON b.attnum = ANY(r.conkey) AND b.attrelid = r.conrelid
  JOIN ONLY pg_class i on i.oid = r.conindid
  JOIN ONLY pg_class c on c.oid = r.confrelid
  JOIN ONLY pg_attribute d ON d.attnum = ANY(r.confkey) AND d.attrelid = r.confrelid
  JOIN ONLY pg_namespace n ON n.oid = r.connamespace
WHERE r.contype = 'f' AND n.nspname = %%schema string%%
ENDSQL
