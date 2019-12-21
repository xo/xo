#!/bin/bash

PGDB=pg://xodb:xodb@localhost:5433/xodb?sslmode=disable

DEST=$1

if [ -z "$DEST" ]; then
  DEST=x
fi

EXTRA=$2

XOBIN=$(which xo)
if [ -e ./xo ]; then
  XOBIN=./xo
fi

set -ex

mkdir -p $DEST
rm -f *.sqlite3
rm -rf $DEST/*.xo.go

# postgres enum list query
COMMENT='Enum represents a enum.'
$XOBIN $PGDB -N -M -B -T Enum -F PgEnums --query-type-comment "$COMMENT" -o $DEST $EXTRA << ENDSQL
SELECT
  t.typname::varchar AS enum_name
FROM pg_type t
  JOIN ONLY pg_namespace n ON n.oid = t.typnamespace
  JOIN ONLY pg_enum e ON t.oid = e.enumtypid
WHERE n.nspname = %%schema string%%
ENDSQL

# postgres enum value list query
COMMENT='EnumValue represents a enum value.'
$XOBIN $PGDB -N -M -B -T EnumValue -F PgEnumValues --query-type-comment "$COMMENT" -o $DEST $EXTRA << ENDSQL
SELECT
  e.enumlabel::varchar AS enum_value,
  e.enumsortorder::integer AS const_value
FROM pg_type t
  JOIN ONLY pg_namespace n ON n.oid = t.typnamespace
  LEFT JOIN pg_enum e ON t.oid = e.enumtypid
WHERE n.nspname = %%schema string%% AND t.typname = %%enum string%%
ENDSQL

# postgres sequence list query
COMMENT='Sequence represents a table that references a sequence.'
$XOBIN $PGDB -N -M -B -T Sequence -F PgSequences -o $DEST $EXTRA << ENDSQL
SELECT
  t.relname::varchar AS table_name
FROM pg_class s
  JOIN pg_depend d ON d.objid = s.oid
  JOIN pg_class t ON d.objid = s.oid AND d.refobjid = t.oid
  JOIN pg_namespace n ON n.oid = s.relnamespace
WHERE n.nspname = %%schema string%% AND s.relkind = 'S'
ENDSQL

# postgres proc list query
COMMENT='Proc represents a stored procedure.'
$XOBIN $PGDB -N -M -B -T Proc -F PgProcs --query-type-comment "$COMMENT" -o $DEST $EXTRA << ENDSQL
SELECT
  p.proname::varchar AS proc_name,
  pg_get_function_result(p.oid)::varchar AS return_type
FROM pg_proc p
  JOIN ONLY pg_namespace n ON p.pronamespace = n.oid
WHERE n.nspname = %%schema string%%
ENDSQL

# postgres proc parameter list query
COMMENT='ProcParam represents a stored procedure param.'
$XOBIN $PGDB -N -M -B -T ProcParam -F PgProcParams --query-type-comment "$COMMENT" -o $DEST $EXTRA << ENDSQL
SELECT
  UNNEST(p.proargnames) as param_name,
  UNNEST(STRING_TO_ARRAY(oidvectortypes(p.proargtypes), ', '))::varchar AS param_type
FROM pg_proc p
  JOIN ONLY pg_namespace n ON p.pronamespace = n.oid
WHERE n.nspname = %%schema string%% AND p.proname = %%proc string%%
ENDSQL

# postgres table list query
COMMENT='Table represents table info.'
$XOBIN $PGDB -N -M -B -T Table -F PgTables --query-type-comment "$COMMENT" -o $DEST $EXTRA << ENDSQL
SELECT
  c.relkind::varchar AS type,
  c.relname::varchar AS table_name,
  false::boolean AS manual_pk
FROM pg_class c
  JOIN ONLY pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname = %%schema string%% AND c.relkind = %%relkind string%%
ENDSQL

# postgres table column list query
FIELDS='FieldOrdinal int,ColumnName string,DataType string,NotNull bool,DefaultValue sql.NullString,IsPrimaryKey bool'
COMMENT='Column represents column info.'
$XOBIN $PGDB -N -M -B -T Column -F PgTableColumns -Z "$FIELDS" --query-type-comment "$COMMENT" -o $DEST $EXTRA << ENDSQL
SELECT
  a.attnum::integer AS field_ordinal,
  a.attname::varchar AS column_name,
  format_type(a.atttypid, a.atttypmod)::varchar AS data_type,
  a.attnotnull::boolean AS not_null,
  COALESCE(pg_get_expr(ad.adbin, ad.adrelid), '')::varchar AS default_value,
  COALESCE(ct.contype = 'p', false)::boolean AS is_primary_key
FROM pg_attribute a
  JOIN ONLY pg_class c ON c.oid = a.attrelid
  JOIN ONLY pg_namespace n ON n.oid = c.relnamespace
  LEFT JOIN pg_constraint ct ON ct.conrelid = c.oid AND a.attnum = ANY(ct.conkey) AND ct.contype = 'p'
  LEFT JOIN pg_attrdef ad ON ad.adrelid = c.oid AND ad.adnum = a.attnum
WHERE a.attisdropped = false AND n.nspname = %%schema string%% AND c.relname = %%table string%% AND (%%sys bool%% OR a.attnum > 0)
ORDER BY a.attnum
ENDSQL

# postgres table foreign key list query
COMMENT='ForeignKey represents a foreign key.'
$XOBIN $PGDB -N -M -B -T ForeignKey -F PgTableForeignKeys --query-type-comment "$COMMENT" -o $DEST $EXTRA << ENDSQL
SELECT
  r.conname::varchar AS foreign_key_name,
  b.attname::varchar AS column_name,
  i.relname::varchar AS ref_index_name,
  c.relname::varchar AS ref_table_name,
  d.attname::varchar AS ref_column_name,
  0::integer AS key_id,
  0::integer AS seq_no,
  ''::varchar AS on_update,
  ''::varchar AS on_delete,
  ''::varchar AS match
FROM pg_constraint r
  JOIN ONLY pg_class a ON a.oid = r.conrelid
  JOIN ONLY pg_attribute b ON b.attisdropped = false AND b.attnum = ANY(r.conkey) AND b.attrelid = r.conrelid
  JOIN ONLY pg_class i on i.oid = r.conindid
  JOIN ONLY pg_class c on c.oid = r.confrelid
  JOIN ONLY pg_attribute d ON d.attisdropped = false AND d.attnum = ANY(r.confkey) AND d.attrelid = r.confrelid
  JOIN ONLY pg_namespace n ON n.oid = r.connamespace
WHERE r.contype = 'f' AND n.nspname = %%schema string%% AND a.relname = %%table string%%
ORDER BY r.conname, b.attname
ENDSQL

# postgres table index list query
COMMENT='Index represents an index.'
$XOBIN $PGDB -N -M -B -T Index -F PgTableIndexes --query-type-comment "$COMMENT" -o $DEST $EXTRA << ENDSQL
SELECT
  DISTINCT ic.relname::varchar AS index_name,
  i.indisunique::boolean AS is_unique,
  i.indisprimary::boolean AS is_primary,
  0::integer AS seq_no,
  ''::varchar AS origin,
  false::boolean AS is_partial
FROM pg_index i
  JOIN ONLY pg_class c ON c.oid = i.indrelid
  JOIN ONLY pg_namespace n ON n.oid = c.relnamespace
  JOIN ONLY pg_class ic ON ic.oid = i.indexrelid
WHERE i.indkey <> '0' AND n.nspname = %%schema string%% AND c.relname = %%table string%%
ENDSQL

# postgres index column list query
COMMENT='IndexColumn represents index column info.'
$XOBIN $PGDB -N -M -B -T IndexColumn -F PgIndexColumns --query-type-comment "$COMMENT" -o $DEST $EXTRA << ENDSQL
SELECT
  (row_number() over())::integer AS seq_no,
  a.attnum::integer AS cid,
  a.attname::varchar AS column_name
FROM pg_index i
  JOIN ONLY pg_class c ON c.oid = i.indrelid
  JOIN ONLY pg_namespace n ON n.oid = c.relnamespace
  JOIN ONLY pg_class ic ON ic.oid = i.indexrelid
  LEFT JOIN pg_attribute a ON i.indrelid = a.attrelid AND a.attnum = ANY(i.indkey) AND a.attisdropped = false
WHERE i.indkey <> '0' AND n.nspname = %%schema string%% AND ic.relname = %%index string%%
ENDSQL

# postgres index column order query
COMMENT='PgColOrder represents index column order.'
$XOBIN $PGDB -N -M -B -1 -T PgColOrder -F PgGetColOrder --query-type-comment "$COMMENT" -o $DEST $EXTRA << ENDSQL
SELECT
  i.indkey::varchar AS ord
FROM pg_index i
  JOIN ONLY pg_class c ON c.oid = i.indrelid
  JOIN ONLY pg_namespace n ON n.oid = c.relnamespace
  JOIN ONLY pg_class ic ON ic.oid = i.indexrelid
WHERE n.nspname = %%schema string%% AND ic.relname = %%index string%%
ENDSQL