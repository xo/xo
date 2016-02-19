#!/bin/bash

PGSQL=pgsql://xodb:xodb@localhost/xodb
MYSQL=mysql://xodb:xodb@localhost/xodb
SQLTE=file:xodb.sqlite3
ORCLE=oracle://xodb:xodb@localhost/xodb

DEST=$1

if [ -z "$DEST" ]; then
  DEST=x
fi

XOBIN=$(which xo)
if [ -e ./xo ]; then
  XOBIN=./xo
fi

set -ex

mkdir -p $DEST
rm -f *.sqlite3
rm -rf $DEST/*.xo.go

# postgresql enum query
cat << ENDSQL | $XOBIN $PGSQL -v -N -M -B -T Enum -F PgEnumsBySchema --query-type-comment='Enum represents a enum value.' -o $DEST
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

# postgresql column query
cat << ENDSQL | $XOBIN $PGSQL -v -N -M -B -T Column -F PgColumnsByRelkindSchema --query-type-comment='Column represents class (ie, table, view, etc) attributes.' -o $DEST
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
WHERE a.attisdropped = false AND c.relkind = %%relkind string%% AND a.attnum > 0 AND n.nspname = %%schema string%%
ORDER BY c.relname, a.attnum
ENDSQL

# postgresql proc query
cat << ENDSQL | $XOBIN $PGSQL -v -N -M -B -T Proc -F PgProcsBySchema --query-type-comment='Proc represents a stored procedure.' -o $DEST
SELECT
  p.proname::varchar AS proc_name,
  oidvectortypes(p.proargtypes)::varchar AS parameter_types,
  pg_get_function_result(p.oid)::varchar AS return_type,
  ''::varchar AS comment
FROM pg_proc p
  INNER JOIN pg_namespace n ON p.pronamespace = n.oid
WHERE n.nspname = %%schema string%%
ORDER BY p.proname
ENDSQL

# postgresql foreign key query
cat << ENDSQL | $XOBIN $PGSQL -v -N -M -B -T ForeignKey -F PgForeignKeysBySchema --query-type-comment='ForeignKey represents a foreign key.' -o $DEST
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
  JOIN ONLY pg_attribute b ON b.attisdropped = false AND b.attnum = ANY(r.conkey) AND b.attrelid = r.conrelid
  JOIN ONLY pg_class i on i.oid = r.conindid
  JOIN ONLY pg_class c on c.oid = r.confrelid
  JOIN ONLY pg_attribute d ON d.attisdropped = false AND d.attnum = ANY(r.confkey) AND d.attrelid = r.confrelid
  JOIN ONLY pg_namespace n ON n.oid = r.connamespace
WHERE r.contype = 'f' AND n.nspname = %%schema string%%
ORDER BY r.conname, a.relname, b.attname
ENDSQL

# mysql enum query
cat << ENDSQL | $XOBIN $MYSQL -v -N -M -B -T MyEnum --query-type-comment='MyEnum represents a MySQL enum.' -o $DEST
SELECT
  table_name AS table_name,
  column_name AS enum_type,
  SUBSTRING(column_type, 6, CHAR_LENGTH(column_type) - 6) AS enum_values
FROM information_schema.columns
WHERE data_type = 'enum' AND table_schema = %%schema string%%
ORDER BY table_name, column_name
ENDSQL

# mysql column query
cat << ENDSQL | $XOBIN $MYSQL -a -v -N -M -B -T Column -F MyColumnsByRelkindSchema -o $DEST
SELECT
  c.column_name,
  c.table_name,
  IF(c.data_type = 'enum', c.column_name, c.column_type) AS data_type,
  c.ordinal_position AS field_ordinal,
  IF(c.is_nullable, true, false) AS is_nullable,
  IF(c.column_key <> '', true, false) AS is_index,
  IF(c.column_key IN('PRI', 'UNI'), true, false) AS is_unique,
  IF(c.column_key = 'PRI', true, false) AS is_primary_key,
  COALESCE((SELECT s.index_name
    FROM information_schema.statistics s
    WHERE s.table_schema = c.table_schema AND s.table_name = c.table_name AND s.column_name = c.column_name), '') AS index_name,
  COALESCE((SELECT x.constraint_name
    FROM information_schema.key_column_usage x
    WHERE x.table_name = c.table_name AND x.column_name = c.column_name AND NOT x.referenced_table_name IS NULL), '') AS foreign_index_name,
  COALESCE(IF(c.column_default IS NULL, true, false), false) AS has_default,
  COALESCE(c.column_default, '') AS default_value,
  COALESCE(c.column_comment, '') AS comment
FROM information_schema.columns c
LEFT JOIN information_schema.tables t ON t.table_schema = c.table_schema AND t.table_name = c.table_name
WHERE t.table_type = %%relkind string%% AND c.table_schema = %%schema string%%
ORDER BY c.table_name, c.ordinal_position
ENDSQL

# mysql proc query
cat << ENDSQL | $XOBIN $MYSQL -a -v -N -M -B -T Proc -F MyProcsBySchema -o $DEST
SELECT
  r.routine_name AS proc_name,
  (SELECT GROUP_CONCAT(l.dtd_identifier SEPARATOR ', ')
    FROM information_schema.parameters l
    WHERE l.specific_schema = r.routine_schema AND l.specific_name = r.routine_name AND l.ordinal_position > 0
    ORDER BY l.ordinal_position) AS parameter_types,
  p.dtd_identifier AS return_type
FROM information_schema.routines r
INNER JOIN information_schema.parameters p ON
  p.specific_schema = r.routine_schema AND p.specific_name = r.routine_name AND p.ordinal_position = 0
WHERE r.routine_schema = %%schema string%%
ORDER BY r.specific_name
ENDSQL

# mysql foreign key query
cat << ENDSQL | $XOBIN $MYSQL -a -v -N -M -B -T ForeignKey -F MyForeignKeysBySchema -o $DEST
SELECT
  constraint_name AS foreign_key_name,
  table_name AS table_name,
  column_name AS column_name,
  '' AS ref_index_name,
  referenced_table_name AS ref_table_name,
  referenced_column_name AS ref_column_name
FROM information_schema.key_column_usage
WHERE referenced_table_name IS NOT NULL AND table_schema = %%schema string%%
ORDER BY table_name, constraint_name
ENDSQL

# sqlite table list query
cat << ENDSQL | $XOBIN $SQLTE -v -N -M -B -T SqTableinfo -o $DEST
SELECT
  type,
  name,
  tbl_name AS table_name
FROM sqlite_master
WHERE type = %%relkind string%%
ENDSQL

# sqlite table info query
cat << ENDSQL | $XOBIN $SQLTE -v -I -N -M -B -T SqColumn -Z 'FieldOrdinal int,ColumnName string,DataType string,NotNull bool,DefaultValue sql.NullString,IsPrimaryKey bool' -o $DEST
PRAGMA table_info(%%table string,interpolate%%)
ENDSQL

# sqlite foreign key list query
cat << ENDSQL | $XOBIN $SQLTE -v -I -N -M -B -T SqForeignKey -Z 'ID int,Seq int,RefTableName string,ColumnName string,RefColumnName string,OnUpdate string,OnDelete string,Match string' -o $DEST
PRAGMA foreign_key_list(%%table string,interpolate%%)
ENDSQL

# sqlite index list query
cat << ENDSQL | $XOBIN $SQLTE -v -I -N -M -B -T SqIndex -Z 'Seq int,IndexName string,IsUnique bool,Origin string,IsPartial bool' -o $DEST
PRAGMA index_list(%%table string,interpolate%%)
ENDSQL

# sqlite index info query
cat << ENDSQL | $XOBIN $SQLTE -v -I -N -M -B -T SqIndexinfo -Z 'SeqNo int,Cid int,ColumnName string' -o $DEST
PRAGMA index_info(%%index string,interpolate%%)
ENDSQL
