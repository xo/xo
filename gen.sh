#!/bin/bash

PGDB=pg://
MYDB=my://localhost/mysql
MSDB=ms://
SQDB=sq:xo.db
ORDB=or://localhost:1521/orasid

DEST=$1
if [ -z "$DEST" ]; then
  echo "usage: $0 <DEST>"
  exit 1
fi
shift

XOBIN=$(which xo)
if [ -e ./xo ]; then
  XOBIN=./xo
fi
XOBIN=$(realpath $XOBIN)

set -ex

mkdir -p $DEST
rm -f *.db
rm -rf $DEST/*.xo.go

# postgres view create query
COMMENT='{{ . }} creates a view for introspection.'
$XOBIN query $PGDB -M -B -X -F PostgresViewCreate --func-comment "$COMMENT" --single=models.xo.go -I -o $DEST $@ << ENDSQL
/* %%schema string,interpolate%% */
CREATE TEMPORARY VIEW %%id string,interpolate%% AS %%query []string,interpolate,join%%
ENDSQL

# postgres view schema query
COMMENT='{{ . }} retrieves the schema for a view created for introspection.'
$XOBIN query $PGDB -M -B -l -F PostgresViewSchema --func-comment "$COMMENT" --single=models.xo.go -I -a -o $DEST $@ << ENDSQL
SELECT
  n.nspname::varchar AS schema_name
FROM pg_class c
  JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname LIKE 'pg_temp%'
  AND c.relname = %%id string%%
ENDSQL

# postgres view drop query
COMMENT='{{ . }} drops a view created for introspection.'
$XOBIN query $PGDB -M -B -X -F PostgresViewDrop --func-comment "$COMMENT" --single=models.xo.go -I -a -o $DEST $@ << ENDSQL
/* %%schema string,interpolate%% */
DROP VIEW %%id string,interpolate%%
ENDSQL

# postgres schema query
COMMENT='{{ . }} retrieves the schema.'
$XOBIN query $PGDB -M -B -l -F PostgresSchema --func-comment "$COMMENT" --single=models.xo.go -a -o $DEST $@ << ENDSQL
SELECT
  CURRENT_SCHEMA()::varchar AS schema_name
ENDSQL

# postgres enum list query
COMMENT='{{ . }} is a enum.'
$XOBIN query $PGDB -M -B -2 -T Enum -F PostgresEnums --type-comment "$COMMENT" -o $DEST $@ << ENDSQL
SELECT
  DISTINCT t.typname::varchar AS enum_name
FROM pg_type t
  JOIN ONLY pg_namespace n ON n.oid = t.typnamespace
  JOIN ONLY pg_enum e ON t.oid = e.enumtypid
WHERE n.nspname = %%schema string%%
ENDSQL

# postgres enum value list query
COMMENT='{{ . }} is a enum value.'
$XOBIN query $PGDB -M -B -2 -T EnumValue -F PostgresEnumValues --type-comment "$COMMENT" -o $DEST $@ << ENDSQL
SELECT
  e.enumlabel::varchar AS enum_value,
  e.enumsortorder::integer AS const_value
FROM pg_type t
  JOIN ONLY pg_namespace n ON n.oid = t.typnamespace
  LEFT JOIN pg_enum e ON t.oid = e.enumtypid
WHERE n.nspname = %%schema string%%
  AND t.typname = %%enum string%%
ENDSQL

# postgres proc list query
COMMENT='{{ . }} is a stored procedure.'
$XOBIN query $PGDB -M -B -2 -T Proc -F PostgresProcs --type-comment "$COMMENT" -o $DEST $@ << ENDSQL
SELECT
  p.oid::varchar AS proc_id,
  p.proname::varchar AS proc_name,
  pp.proc_type::varchar AS proc_type,
  format_type(pp.return_type, NULL)::varchar AS return_type,
  pp.return_name::varchar AS return_name,
  p.prosrc::varchar AS proc_def
FROM pg_catalog.pg_proc p
  JOIN pg_catalog.pg_namespace n ON (p.pronamespace = n.oid)
  JOIN (
    SELECT
      p.oid,
      (CASE WHEN EXISTS(
          SELECT
            column_name
          FROM information_schema.columns
          WHERE table_name = 'pg_proc'
            AND column_name = 'prokind'
        )
        THEN (CASE p.prokind
          WHEN 'p' THEN 'procedure'
          WHEN 'f' THEN 'function'
        END)
        ELSE ''
      END) AS proc_type,
      UNNEST(COALESCE(p.proallargtypes, ARRAY[p.prorettype])) AS return_type,
      UNNEST(CASE
        WHEN p.proargmodes IS NULL THEN ARRAY['']
        ELSE p.proargnames
      END) AS return_name,
      UNNEST(COALESCE(p.proargmodes, ARRAY['o'])) AS param_type
    FROM pg_catalog.pg_proc p
  ) AS pp ON p.oid = pp.oid
WHERE p.prorettype <> 'pg_catalog.cstring'::pg_catalog.regtype
  AND (p.proargtypes[0] IS NULL
    OR p.proargtypes[0] <> 'pg_catalog.cstring'::pg_catalog.regtype)
  AND (pp.proc_type = 'function'
    OR pp.proc_type = 'procedure')
  AND pp.param_type = 'o'
  AND n.nspname = %%schema string%%
ENDSQL

# postgres proc parameter list query
COMMENT='{{ . }} is a stored procedure param.'
$XOBIN query $PGDB -M -B -2 -T ProcParam -F PostgresProcParams --type-comment "$COMMENT" -o $DEST $@ << ENDSQL
SELECT
  COALESCE(pp.param_name, '')::varchar AS param_name,
  pp.param_type::varchar AS param_type
FROM pg_proc p
  JOIN ONLY pg_namespace n ON p.pronamespace = n.oid
  JOIN (
    SELECT
      p.oid,
      UNNEST(p.proargnames) AS param_name,
      format_type(UNNEST(p.proargtypes), NULL) AS param_type
    FROM pg_proc p
  ) AS pp ON p.oid = pp.oid
WHERE n.nspname = %%schema string%%
  AND p.oid::varchar = %%id string%%
  AND pp.param_type IS NOT NULL
ENDSQL

# postgres table list query
COMMENT='{{ . }} is a table.'
$XOBIN query $PGDB -M -B -2 -T Table -F PostgresTables --type-comment "$COMMENT" -o $DEST $@ << ENDSQL
SELECT
  (CASE c.relkind
    WHEN 'r' THEN 'table'
    WHEN 'v' THEN 'view'
  END)::varchar AS type,
  c.relname::varchar AS table_name,
  false::boolean AS manual_pk,
  CASE c.relkind
    WHEN 'r' THEN ''
    WHEN 'v' THEN v.definition
  END AS view_def
FROM pg_class c
  JOIN ONLY pg_namespace n ON n.oid = c.relnamespace
  LEFT JOIN pg_views v ON n.nspname = v.schemaname
    AND v.viewname = c.relname
WHERE n.nspname = %%schema string%%
  AND (CASE c.relkind
    WHEN 'r' THEN 'table'
    WHEN 'v' THEN 'view'
  END) = LOWER(%%typ string%%)
ENDSQL

# postgres table column list query
FIELDS='FieldOrdinal int,ColumnName string,DataType string,NotNull bool,DefaultValue sql.NullString,IsPrimaryKey bool,Comment sql.NullString'
COMMENT='{{ . }} is a column.'
$XOBIN query $PGDB -M -B -2 -T Column -F PostgresTableColumns -Z "$FIELDS" --type-comment "$COMMENT" -o $DEST $@ << ENDSQL
SELECT
  a.attnum::integer AS field_ordinal,
  a.attname::varchar AS column_name,
  format_type(a.atttypid, a.atttypmod)::varchar AS data_type,
  a.attnotnull::boolean AS not_null,
  COALESCE(pg_get_expr(ad.adbin, ad.adrelid), '')::varchar AS default_value,
  COALESCE(ct.contype = 'p', false)::boolean AS is_primary_key,
  d.description::varchar as comment
FROM pg_attribute a
  JOIN ONLY pg_class c ON c.oid = a.attrelid
  JOIN ONLY pg_namespace n ON n.oid = c.relnamespace
  LEFT JOIN pg_constraint ct ON ct.conrelid = c.oid
    AND a.attnum = ANY(ct.conkey)
    AND ct.contype = 'p'
  LEFT JOIN pg_attrdef ad ON ad.adrelid = c.oid
    AND ad.adnum = a.attnum
  LEFT JOIN pg_description d on d.objoid = c.oid
		AND d.objsubid = a.attnum
WHERE a.attisdropped = false
  AND n.nspname = %%schema string%%
  AND c.relname = %%table string%%
  AND (%%sys bool%% OR a.attnum > 0)
ORDER BY a.attnum
ENDSQL

# postgres sequence list query
COMMENT='{{ . }} is a sequence.'
$XOBIN query $PGDB -M -B -2 -T Sequence -F PostgresTableSequences --type-comment "$COMMENT" -o $DEST $@ << ENDSQL
SELECT
  a.attname::varchar as column_name
FROM pg_class s
  JOIN pg_depend d ON d.objid = s.oid
  JOIN pg_class t ON d.objid = s.oid AND d.refobjid = t.oid
  JOIN pg_attribute a ON (d.refobjid, d.refobjsubid) = (a.attrelid, a.attnum)
  JOIN pg_namespace n ON n.oid = s.relnamespace
WHERE s.relkind = 'S'
  AND n.nspname = %%schema string%%
  AND t.relname = %%table string%%
ENDSQL

# postgres table foreign key list query
COMMENT='{{ . }} is a foreign key.'
$XOBIN query $PGDB -M -B -2 -T ForeignKey -F PostgresTableForeignKeys --type-comment "$COMMENT" -o $DEST $@ << ENDSQL
SELECT
  tc.constraint_name::varchar AS foreign_key_name,
  kcu.column_name::varchar AS column_name,
  ccu.table_name::varchar AS ref_table_name,
  ccu.column_name::varchar AS ref_column_name,
  0::integer AS key_id
FROM information_schema.table_constraints tc
  JOIN information_schema.key_column_usage AS kcu ON tc.constraint_name = kcu.constraint_name
    AND tc.table_schema = kcu.table_schema
  JOIN (
    SELECT
      ROW_NUMBER() OVER (
        PARTITION BY
          table_schema,
          table_name,
          constraint_name
        ORDER BY row_num
      ) AS ordinal_position,
      table_schema,
      table_name,
      column_name,
      constraint_name
    FROM (
      SELECT
        ROW_NUMBER() OVER (ORDER BY 1) AS row_num,
        table_schema,
        table_name,
        column_name,
        constraint_name
      FROM information_schema.constraint_column_usage
    ) t
  ) AS ccu ON ccu.constraint_name = tc.constraint_name
    AND ccu.table_schema = tc.table_schema
    AND ccu.ordinal_position = kcu.ordinal_position
WHERE tc.constraint_type = 'FOREIGN KEY'
  AND tc.table_schema = %%schema string%%
  AND tc.table_name = %%table string%%
ENDSQL

# postgres table index list query
COMMENT='{{ . }} is a index.'
$XOBIN query $PGDB -M -B -2 -T Index -F PostgresTableIndexes --type-comment "$COMMENT" -o $DEST $@ << ENDSQL
SELECT
  DISTINCT ic.relname::varchar AS index_name,
  i.indisunique::boolean AS is_unique,
  i.indisprimary::boolean AS is_primary
FROM pg_index i
  JOIN ONLY pg_class c ON c.oid = i.indrelid
  JOIN ONLY pg_namespace n ON n.oid = c.relnamespace
  JOIN ONLY pg_class ic ON ic.oid = i.indexrelid
WHERE i.indkey <> '0'
  AND n.nspname = %%schema string%%
  AND c.relname = %%table string%%
ENDSQL

# postgres index column list query
COMMENT='{{ . }} is a index column.'
$XOBIN query $PGDB -M -B -2 -T IndexColumn -F PostgresIndexColumns --type-comment "$COMMENT" -o $DEST $@ << ENDSQL
SELECT
  (row_number() over())::integer AS seq_no,
  a.attnum::integer AS cid,
  a.attname::varchar AS column_name
FROM pg_index i
  JOIN ONLY pg_class c ON c.oid = i.indrelid
  JOIN ONLY pg_namespace n ON n.oid = c.relnamespace
  JOIN ONLY pg_class ic ON ic.oid = i.indexrelid
  LEFT JOIN pg_attribute a ON i.indrelid = a.attrelid
    AND a.attnum = ANY(i.indkey)
    AND a.attisdropped = false
WHERE i.indkey <> '0'
  AND n.nspname = %%schema string%%
  AND ic.relname = %%index string%%
ENDSQL

# postgres index column order query
COMMENT='{{ . }} is a index column order.'
$XOBIN query $PGDB -M -B -1 -2 -T PostgresColOrder -F PostgresGetColOrder --type-comment "$COMMENT" -o $DEST $@ << ENDSQL
SELECT
  i.indkey::varchar AS ord
FROM pg_index i
  JOIN ONLY pg_class c ON c.oid = i.indrelid
  JOIN ONLY pg_namespace n ON n.oid = c.relnamespace
  JOIN ONLY pg_class ic ON ic.oid = i.indexrelid
WHERE n.nspname = %%schema string%%
  AND ic.relname = %%index string%%
ENDSQL

# mysql view create query
COMMENT='{{ . }} creates a view for introspection.'
$XOBIN query $MYDB -M -B -X -F MysqlViewCreate --func-comment "$COMMENT" --single=models.xo.go -I -a -o $DEST $@ << ENDSQL
/* %%schema string,interpolate%% */
CREATE VIEW %%id string,interpolate%% AS %%query []string,interpolate,join%%
ENDSQL

# mysql view drop query
COMMENT='{{ . }} drops a view created for introspection.'
$XOBIN query $MYDB -M -B -X -F MysqlViewDrop --func-comment "$COMMENT" --single=models.xo.go -I -a -o $DEST $@ << ENDSQL
/* %%schema string,interpolate%% */
DROP VIEW %%id string,interpolate%%
ENDSQL

# mysql schema query
COMMENT='{{ . }} retrieves the schema.'
$XOBIN query $MYDB -M -B -l -F MysqlSchema --func-comment "$COMMENT" --single=models.xo.go -a -o $DEST $@ << ENDSQL
SELECT
  SCHEMA() AS schema_name
ENDSQL

# mysql enum list query
$XOBIN query $MYDB -M -B -2 -T Enum -F MysqlEnums -a -o $DEST $@ << ENDSQL
SELECT
  DISTINCT column_name AS enum_name
FROM information_schema.columns
WHERE data_type = 'enum'
  AND table_schema = %%schema string%%
ENDSQL

# mysql enum value list query
$XOBIN query $MYDB -M -B -1 -2 -T MysqlEnumValue -F MysqlEnumValues -o $DEST $@ << ENDSQL
SELECT
  SUBSTRING(column_type, 6, CHAR_LENGTH(column_type) - 6) AS enum_values
FROM information_schema.columns
WHERE data_type = 'enum'
  AND table_schema = %%schema string%%
  AND column_name = %%enum string%%
ENDSQL

# mysql proc list query
$XOBIN query $MYDB -M -B -2 -T Proc -F MysqlProcs -a -o $DEST $@ << ENDSQL
SELECT
  r.routine_name AS proc_id,
  r.routine_name AS proc_name,
  LOWER(r.routine_type) AS proc_type,
  COALESCE(p.dtd_identifier, 'void') AS return_type,
  COALESCE(p.parameter_name, '') AS return_name,
  r.routine_definition AS proc_def
FROM information_schema.routines r
  LEFT JOIN information_schema.parameters p ON p.specific_schema = r.routine_schema
    AND p.specific_name = r.routine_name
    AND (p.parameter_mode = 'OUT' OR p.ordinal_position = 0)
WHERE r.routine_schema = %%schema string%%
ORDER BY r.routine_name, p.ordinal_position
ENDSQL

# mysql proc parameter list query
$XOBIN query $MYDB -M -B -2 -T ProcParam -F MysqlProcParams -a -o $DEST $@ << ENDSQL
SELECT
  p.parameter_name AS param_name,
  p.dtd_identifier AS param_type
FROM information_schema.routines r
  INNER JOIN information_schema.parameters p ON p.specific_schema = r.routine_schema
    AND p.specific_name = r.routine_name
    AND p.parameter_mode = 'IN'
WHERE r.routine_schema = %%schema string%%
  AND r.routine_name = %%id string%%
ORDER BY p.ordinal_position
ENDSQL

# mysql table list query
$XOBIN query $MYDB -M -B -2 -T Table -F MysqlTables -a -o $DEST $@ << ENDSQL
SELECT
  (CASE t.table_type
    WHEN 'BASE TABLE' THEN 'table'
    WHEN 'VIEW' THEN 'view'
  END) AS type,
  t.table_name,
  CASE t.table_type
    WHEN 'BASE TABLE' THEN ''
    WHEN 'VIEW' then v.view_definition
  END AS view_def
FROM information_schema.tables t
  LEFT JOIN information_schema.views v ON t.table_schema = v.table_schema
    AND t.table_name = v.table_name
WHERE t.table_schema = %%schema string%%
  AND (CASE t.table_type
    WHEN 'BASE TABLE' THEN 'table'
    WHEN 'VIEW' THEN 'view'
  END) = LOWER(%%typ string%%)
ENDSQL

# mysql table column list query
$XOBIN query $MYDB -M -B -2 -T Column -F MysqlTableColumns -a -o $DEST $@ << ENDSQL
SELECT
  ordinal_position AS field_ordinal,
  column_name,
  IF(data_type = 'enum', column_name, column_type) AS data_type,
  IF(is_nullable = 'YES', false, true) AS not_null,
  column_default AS default_value,
  IF(column_key = 'PRI', true, false) AS is_primary_key
FROM information_schema.columns
WHERE table_schema = %%schema string%%
  AND table_name = %%table string%%
ORDER BY ordinal_position
ENDSQL

# mysql sequence list query
$XOBIN query $MYDB -M -B -2 -T Sequence -F MysqlTableSequences -a -o $DEST $@ << ENDSQL
SELECT
  column_name
FROM information_schema.columns c
WHERE c.extra = 'auto_increment'
  AND c.table_schema = %%schema string%%
  AND c.table_name = %%table string%%
ENDSQL

# mysql table foreign key list query
$XOBIN query $MYDB -M -B -2 -T ForeignKey -F MysqlTableForeignKeys -a -o $DEST $@ << ENDSQL
SELECT
  constraint_name AS foreign_key_name,
  column_name AS column_name,
  referenced_table_name AS ref_table_name,
  referenced_column_name AS ref_column_name
FROM information_schema.key_column_usage
WHERE referenced_table_name IS NOT NULL
  AND table_schema = %%schema string%%
  AND table_name = %%table string%%
ENDSQL

# mysql table index list query
$XOBIN query $MYDB -M -B -2 -T Index -F MysqlTableIndexes -a -o $DEST $@ << ENDSQL
SELECT
  DISTINCT index_name,
  NOT non_unique AS is_unique
FROM information_schema.statistics
WHERE index_name <> 'PRIMARY'
  AND index_schema = %%schema string%%
  AND table_name = %%table string%%
ENDSQL

# mysql index column list query
$XOBIN query $MYDB -M -B -2 -T IndexColumn -F MysqlIndexColumns -a -o $DEST $@ << ENDSQL
SELECT
  seq_in_index AS seq_no,
  column_name
FROM information_schema.statistics
WHERE index_schema = %%schema string%%
  AND table_name = %%table string%%
  AND index_name = %%index string%%
ORDER BY seq_in_index
ENDSQL

# sqlite3 view create query
COMMENT='{{ . }} creates a view for introspection.'
$XOBIN query $SQDB -M -B -X -F Sqlite3ViewCreate --func-comment "$COMMENT" --single=models.xo.go -I -a -o $DEST $@ << ENDSQL
/* %%schema string,interpolate%% */
CREATE TEMPORARY VIEW %%id string,interpolate%% AS %%query []string,interpolate,join%%
ENDSQL

# sqlite3 view drop query
COMMENT='{{ . }} drops a view created for introspection.'
$XOBIN query $SQDB -M -B -X -F Sqlite3ViewDrop --func-comment "$COMMENT" --single=models.xo.go -I -a -o $DEST $@ << ENDSQL
/* %%schema string,interpolate%% */
DROP VIEW %%id string,interpolate%%
ENDSQL

# sqlite3 schema query
COMMENT='{{ . }} retrieves the schema.'
$XOBIN query $SQDB -M -B -l -F Sqlite3Schema --func-comment "$COMMENT" --single=models.xo.go -a -o $DEST $@ << ENDSQL
SELECT
  REPLACE(file, RTRIM(file, REPLACE(file, '/', '')), '') AS schema_name
FROM pragma_database_list()
ENDSQL

# sqlite3 table list query
$XOBIN query $SQDB -M -B -2 -T Table -F Sqlite3Tables -I -a -o $DEST $@ << ENDSQL
/* %%schema string,interpolate%% */
SELECT
  type,
  tbl_name AS table_name,
  CASE LOWER(type)
    WHEN 'table' THEN ''
    WHEN 'view' THEN sql
  END AS view_def
FROM sqlite_master
WHERE tbl_name NOT LIKE 'sqlite_%'
  AND LOWER(type) = LOWER(%%typ string%%)
ENDSQL

# sqlite3 table column list query
$XOBIN query $SQDB -M -B -2 -T Column -F Sqlite3TableColumns -I -a -o $DEST $@ << ENDSQL
/* %%schema string,interpolate%% */
SELECT
  cid AS field_ordinal,
  name AS column_name,
  type AS data_type,
  "notnull" AS not_null,
  dflt_value AS default_value,
  CAST(pk <> 0 AS boolean) AS is_primary_key
FROM pragma_table_info(%%table string%%)
ENDSQL

# sqlite3 sequence list query
$XOBIN query $SQDB -M -B -2 -T Sequence -F Sqlite3TableSequences -I -a -o $DEST $@ << ENDSQL
/* %%schema string,interpolate%% */
WITH RECURSIVE
  a AS (
    SELECT name, lower(replace(replace(sql, char(13), ' '), char(10), ' ')) AS sql
    FROM sqlite_master
    WHERE lower(sql) LIKE '%integer% autoincrement%'
  ),
  b AS (
    SELECT name, trim(substr(sql, instr(sql, '(') + 1)) AS sql
    FROM a
  ),
  c AS (
    SELECT b.name, sql, '' AS col
    FROM b
    UNION ALL
    SELECT
      c.name,
      trim(substr(c.sql, ifnull(nullif(instr(c.sql, ','), 0), instr(c.sql, ')')) + 1)) AS sql,
      trim(substr(c.sql, 1, ifnull(nullif(instr(c.sql, ','), 0), instr(c.sql, ')')) - 1)) AS col
    FROM c JOIN b ON c.name = b.name
    WHERE c.sql != ''
  ),
  d AS (
    SELECT name, substr(col, 1, instr(col, ' ') - 1) AS col
    FROM c
    WHERE col LIKE '%autoincrement%'
  )
SELECT col AS column_name
FROM d
WHERE name = %%table string%%
ENDSQL

# sqlite3 table foreign key list query
$XOBIN query $SQDB -M -B -2 -T ForeignKey -F Sqlite3TableForeignKeys -I -a -o $DEST $@ << ENDSQL
/* %%schema string,interpolate%% */
SELECT
  id AS key_id,
  "table" AS ref_table_name,
  "from" AS column_name,
  "to" AS ref_column_name
FROM pragma_foreign_key_list(%%table string%%)
ENDSQL

# sqlite3 table index list query
$XOBIN query $SQDB -M -B -2 -T Index -F Sqlite3TableIndexes -I -a -o $DEST $@ << ENDSQL
/* %%schema string,interpolate%% */
SELECT
  name AS index_name,
  "unique" AS is_unique,
  CAST(origin = 'pk' AS boolean) AS is_primary
FROM pragma_index_list(%%table string%%)
ENDSQL

# sqlite3 index column list query
$XOBIN query $SQDB -M -B -2 -T IndexColumn -F Sqlite3IndexColumns -I -a -o $DEST $@ << ENDSQL
/* %%schema string,interpolate%% %%table string,interpolate%% */
SELECT
  seqno AS seq_no,
  cid,
  name AS column_name
FROM pragma_index_info(%%index string%%)
ENDSQL

# sqlserver view create query
COMMENT='{{ . }} creates a view for introspection.'
$XOBIN query $MSDB -M -B -X -F SqlserverViewCreate --func-comment "$COMMENT" --single=models.xo.go -I -a -o $DEST $@ << ENDSQL
/* %%schema string,interpolate%% */
CREATE VIEW %%id string,interpolate%% AS %%query []string,interpolate,join%%
ENDSQL

# sqlserver view drop query
COMMENT='{{ . }} drops a view created for introspection.'
$XOBIN query $MSDB -M -B -X -F SqlserverViewDrop --func-comment "$COMMENT" --single=models.xo.go -I -a -o $DEST $@ << ENDSQL
/* %%schema string,interpolate%% */
DROP VIEW %%id string,interpolate%%
ENDSQL

# sqlserver schema query
COMMENT='{{ . }} retrieves the schema.'
$XOBIN query $MSDB -M -B -l -F SqlserverSchema --func-comment "$COMMENT" --single=models.xo.go -a -o $DEST $@ << ENDSQL
SELECT
  SCHEMA_NAME() AS schema_name
ENDSQL

# sqlserver proc list query
$XOBIN query $MSDB -M -B -2 -T Proc -F SqlserverProcs -a -o $DEST $@ << ENDSQL
SELECT
  STR(o.object_id) AS proc_id,
  o.name AS proc_name,
  (CASE o.type
    WHEN 'P' THEN 'procedure'
    WHEN 'FN' THEN 'function'
  END) AS proc_type,
  CASE
    WHEN p.object_id IS NOT NULL
      THEN TYPE_NAME(p.system_type_id)+IIF(p.precision > 0, '('+CAST(p.precision AS varchar)+IIF(p.scale > 0,','+CAST(p.scale AS varchar),'')+')', '')
    ELSE 'void'
  END AS return_type,
  CASE
  WHEN p.object_id IS NOT NULL AND p.name <> ''
      THEN SUBSTRING(p.name, 2, LEN(p.name)-1)
    ELSE ''
  END AS return_name,
  OBJECT_DEFINITION(o.object_id) AS proc_def
FROM sys.objects o
  LEFT JOIN sys.parameters p ON o.object_id = p.object_id
    AND (p.object_id IS NULL OR p.is_output = 'true')
WHERE o.type = 'P'
   OR o.type = 'FN'
  AND SCHEMA_NAME(o.schema_id) = %%schema string%%
ORDER BY o.object_id
ENDSQL

# sqlserver proc parameter list query
$XOBIN query $MSDB -M -B -2 -T ProcParam -F SqlserverProcParams -a -o $DEST $@ << ENDSQL
SELECT
  SUBSTRING(p.name, 2, LEN(p.name)-1) AS param_name,
  TYPE_NAME(p.user_type_id) AS param_type
FROM sys.objects o
  INNER JOIN sys.parameters p ON o.object_id = p.object_id
WHERE SCHEMA_NAME(schema_id) = %%schema string%%
  AND STR(o.object_id) = %%id string%%
  AND p.is_output = 'false'
ORDER BY p.parameter_id
ENDSQL

# sqlserver table list query
$XOBIN query $MSDB -M -B -2 -T Table -F SqlserverTables -a -o $DEST $@ << ENDSQL
SELECT
  (CASE xtype
    WHEN 'U' THEN 'table'
    WHEN 'V' THEN 'view'
  END) AS type,
  name AS table_name,
  CASE xtype
    WHEN 'U' THEN ''
    WHEN 'V' THEN OBJECT_DEFINITION(id)
  END AS view_def
FROM sysobjects
WHERE SCHEMA_NAME(uid) = %%schema string%%
  AND (CASE xtype
    WHEN 'U' THEN 'table'
    WHEN 'V' THEN 'view'
  END) = LOWER(%%typ string%%)
ENDSQL

# sqlserver table column list query
$XOBIN query $MSDB -M -B -2 -T Column -F SqlserverTableColumns -a -o $DEST $@ << ENDSQL
SELECT
  c.colid AS field_ordinal,
  c.name AS column_name,
  TYPE_NAME(c.xtype)+IIF(c.prec > 0, '('+CAST(c.prec AS varchar)+IIF(c.scale > 0,','+CAST(c.scale AS varchar),'')+')', '') AS data_type,
  IIF(c.isnullable=1, 0, 1) AS not_null,
  x.text AS default_value,
  IIF(COALESCE((
    SELECT COUNT(z.colid)
    FROM sysindexes i
      INNER JOIN sysindexkeys z ON i.id = z.id
        AND i.indid = z.indid
        AND z.colid = c.colid
    WHERE i.id = o.id
      AND i.name = k.name
  ), 0) > 0, 1, 0) AS is_primary_key
FROM syscolumns c
  JOIN sysobjects o ON o.id = c.id
  LEFT JOIN sysobjects k ON k.xtype = 'PK'
    AND k.parent_obj = o.id
  LEFT JOIN syscomments x ON x.id = c.cdefault
WHERE o.type IN('U', 'V')
  AND SCHEMA_NAME(o.uid) = %%schema string%%
  AND o.name = %%table string%%
ORDER BY c.colid
ENDSQL

# sqlserver sequence list query
$XOBIN query $MSDB -M -B -2 -T Sequence -F SqlserverTableSequences -a -o $DEST $@ << ENDSQL
SELECT
  COL_NAME(o.object_id, c.column_id) AS column_name
FROM sys.objects o
  INNER JOIN sys.columns c ON o.object_id = c.object_id
WHERE c.is_identity = 1
  AND o.type = 'U'
  AND SCHEMA_NAME(o.schema_id) = %%schema string%%
  AND o.name = %%table string%%
ENDSQL

# sqlserver table foreign key list query
$XOBIN query $MSDB -M -B -2 -T ForeignKey -F SqlserverTableForeignKeys -a -o $DEST $@ << ENDSQL
SELECT
  fk.name AS foreign_key_name,
  col.name AS column_name,
  pk_tab.name AS ref_table_name,
  pk_col.name AS ref_column_name
FROM sys.tables tab
  INNER JOIN sys.columns col ON col.object_id = tab.object_id
  LEFT OUTER JOIN sys.foreign_key_columns fk_cols ON fk_cols.parent_object_id = tab.object_id
    AND fk_cols.parent_column_id = col.column_id
  LEFT OUTER JOIN sys.foreign_keys fk ON fk.object_id = fk_cols.constraint_object_id
  LEFT OUTER JOIN sys.tables pk_tab ON pk_tab.object_id = fk_cols.referenced_object_id
  LEFT OUTER JOIN sys.columns pk_col ON pk_col.column_id = fk_cols.referenced_column_id
    AND pk_col.object_id = fk_cols.referenced_object_id
WHERE schema_name(tab.schema_id) = %%schema string%%
  AND tab.name = %%table string%%
  AND fk.object_id IS NOT NULL
ENDSQL

# sqlserver table index list query
$XOBIN query $MSDB -M -B -2 -T Index -F SqlserverTableIndexes -a -o $DEST $@ << ENDSQL
SELECT
  i.name AS index_name,
  i.is_primary_key AS is_primary,
  i.is_unique
FROM sys.indexes i
  INNER JOIN sysobjects o ON i.object_id = o.id
WHERE i.name IS NOT NULL
  AND o.type = 'U'
  AND SCHEMA_NAME(o.uid) = %%schema string%%
  AND o.name = %%table string%%
ENDSQL

# sqlserver index column list query
$XOBIN query $MSDB -M -B -2 -T IndexColumn -F SqlserverIndexColumns -a -o $DEST $@ << ENDSQL
SELECT
  k.keyno AS seq_no,
  k.colid AS cid,
  c.name AS column_name
FROM sysindexes i
  INNER JOIN sysobjects o ON i.id = o.id
  INNER JOIN sysindexkeys k ON k.id = o.id
    AND k.indid = i.indid
  INNER JOIN syscolumns c ON c.id = o.id
    AND c.colid = k.colid
WHERE o.type = 'U'
  AND SCHEMA_NAME(o.uid) = %%schema string%%
  AND o.name = %%table string%%
  AND i.name = %%index string%%
ORDER BY k.keyno
ENDSQL

# oracle view create query
COMMENT='{{ . }} creates a view for introspection.'
$XOBIN query $ORDB -M -B -X -F OracleViewCreate --func-comment "$COMMENT" --single=models.xo.go -I -a -o $DEST $@ << ENDSQL
/* %%schema string,interpolate%% */
CREATE GLOBAL TEMPORARY TABLE %%id string,interpolate%% ON COMMIT PRESERVE ROWS AS %%query []string,interpolate,join%%
ENDSQL

# oracle view truncate query
COMMENT='{{ . }} truncates a view created for introspection.'
$XOBIN query $ORDB -M -B -X -F OracleViewTruncate --func-comment "$COMMENT" --single=models.xo.go -I -a -o $DEST $@ << ENDSQL
/* %%schema string,interpolate%% */
TRUNCATE TABLE %%id string,interpolate%%
ENDSQL

# oracle view drop query
COMMENT='{{ . }} drops a view created for introspection.'
$XOBIN query $ORDB -M -B -X -F OracleViewDrop --func-comment "$COMMENT" --single=models.xo.go -I -a -o $DEST $@ << ENDSQL
/* %%schema string,interpolate%% */
DROP TABLE %%id string,interpolate%%
ENDSQL

# oracle schema query
COMMENT='{{ . }} retrieves the schema.'
$XOBIN query $ORDB -M -B -l -F OracleSchema --func-comment "$COMMENT" --single=models.xo.go -a -o $DEST $@ << ENDSQL
SELECT
  LOWER(SYS_CONTEXT('USERENV', 'CURRENT_SCHEMA')) AS schema_name
FROM dual
ENDSQL

# oracle proc list query
$XOBIN query $ORDB -M -B -2 -T Proc -F OracleProcs -a -o $DEST $@ << ENDSQL
SELECT
  CAST(o.object_id AS NVARCHAR2(255)) AS proc_id,
  LOWER(o.object_name) AS proc_name,
  LOWER(o.object_type) AS proc_type,
  LOWER(CASE
    WHEN a.data_type IS NULL THEN 'void'
    WHEN a.data_type = 'CHAR' THEN 'CHAR(' || a.data_length || ')'
    WHEN a.data_type = 'VARCHAR2' THEN 'VARCHAR2(' || a.data_length || ')'
    WHEN a.data_type = 'NUMBER' THEN 'NUMBER(' || NVL(a.data_precision, 0) || ',' || NVL(a.data_scale, 0) || ')'
    ELSE a.data_type END) AS return_type,
  LOWER(CASE
    WHEN a.argument_name IS NULL THEN '-'
    ELSE a.argument_name END) AS return_name,
  s.src AS proc_def
FROM all_objects o
  LEFT JOIN sys.all_arguments a ON a.object_id = o.object_id
    AND a.in_out = 'OUT'
  JOIN (
    SELECT
      s.owner,
      s.name,
      listagg(s.text) WITHIN GROUP (ORDER BY s.line) AS src
    FROM all_source s
    GROUP BY s.owner, s.name
  ) s ON s.owner = o.owner
  AND s.name = o.object_name
WHERE o.object_type IN ('FUNCTION','PROCEDURE')
  AND o.owner = UPPER(%%schema string%%)
ORDER BY o.object_id
ENDSQL

# oracle proc parameter list query
$XOBIN query $ORDB -M -B -2 -T ProcParam -F OracleProcParams -a -o $DEST $@ << ENDSQL
SELECT
  LOWER(a.argument_name) AS param_name,
  LOWER(CASE a.data_type
    WHEN 'CHAR' THEN 'CHAR(' || a.data_length || ')'
    WHEN 'VARCHAR2' THEN 'VARCHAR2(' || a.data_length || ')'
    WHEN 'NUMBER' THEN 'NUMBER(' || NVL(a.data_precision, 0) || ',' || NVL(a.data_scale, 0) || ')'
    ELSE a.data_type END) AS param_type
FROM all_objects o
  JOIN sys.all_arguments a ON a.object_id = o.object_id
    AND a.in_out = 'IN'
WHERE o.object_type IN ('FUNCTION','PROCEDURE')
  AND o.owner = UPPER(%%schema string%%)
  AND CAST(o.object_id AS NVARCHAR2(255)) = UPPER(%%id string%%)
ORDER BY a.position
ENDSQL

# oracle table list query
$XOBIN query $ORDB -M -B -2 -T Table -F OracleTables -a -o $DEST $@ << ENDSQL
SELECT
  LOWER(o.object_type) AS type,
  LOWER(o.object_name) AS table_name,
  CASE o.object_type
    WHEN 'TABLE' THEN ' '
    WHEN 'VIEW' THEN v.text_vc
  END AS view_def
FROM all_objects o
  LEFT JOIN all_views v ON o.owner = v.owner
    AND o.object_name = v.view_name
WHERE o.object_name NOT LIKE '%$%'
  AND o.object_name NOT LIKE 'LOGMNR%_%'
  AND o.object_name NOT LIKE 'REDO_%'
  AND o.object_name NOT LIKE 'SCHEDULER_%_TBL'
  AND o.object_name NOT LIKE 'SQLPLUS_%'
  AND o.owner = UPPER(%%schema string%%)
  AND o.object_type = UPPER(%%typ string%%)
ENDSQL

# Thanks to github.com/svenwiltink
# oracle table column list query
$XOBIN query $ORDB -M -B -2 -T Column -F OracleTableColumns -a -o $DEST $@ << ENDSQL
SELECT
  c.column_id AS field_ordinal,
  LOWER(c.column_name) AS column_name,
  LOWER(CASE c.data_type
    WHEN 'CHAR' THEN 'CHAR(' || c.char_length || ')'
    WHEN 'NCHAR' THEN 'NCHAR(' || c.char_length || ')'
    WHEN 'VARCHAR2' THEN 'VARCHAR2(' || c.char_length || ')'
    WHEN 'NVARCHAR2' THEN 'NVARCHAR2(' || c.char_length || ')'
    WHEN 'NUMBER' THEN 'NUMBER(' || NVL(c.data_precision, 0) || ',' || NVL(c.data_scale, 0) || ')'
    WHEN 'RAW' THEN 'RAW(' || c.data_length || ')'
    ELSE c.data_type END) AS data_type,
  CASE WHEN c.nullable = 'N' THEN '1' ELSE '0' END AS not_null,
  CASE WHEN p.column_id IS NOT NULL THEN '1' ELSE '0' END as is_primary_key
FROM all_tab_columns c
  LEFT JOIN (
    SELECT distinct c.column_id FROM all_tab_columns c
    JOIN all_cons_columns l ON l.owner = c.owner
      AND c.column_name = l.column_name
    JOIN all_constraints r ON r.owner = c.owner
      AND r.table_name = c.table_name
      AND r.constraint_name = l.constraint_name
      AND r.constraint_type = 'P'
    WHERE c.owner = UPPER(%%schema string%%)
    AND c.table_name = UPPER(%%table string%%)
  ) p on p.column_id = c.column_id
WHERE c.owner = UPPER(%%schema string%%)
  AND c.table_name = UPPER(%%table string%%)
ORDER BY c.column_id
ENDSQL

# oracle sequence list query
$XOBIN query $ORDB -M -B -2 -T Sequence -F OracleTableSequences -a -o $DEST $@ << ENDSQL
SELECT
  LOWER(c.column_name) AS column_name
FROM all_tab_columns c
WHERE c.identity_column='YES'
  AND c.owner = UPPER(%%schema string%%)
  AND c.table_name  = UPPER(%%table string%%)
ENDSQL

# oracle table foreign key list query
$XOBIN query $ORDB -M -B -2 -T ForeignKey -F OracleTableForeignKeys -a -o $DEST $@ << ENDSQL
SELECT
  LOWER(a.constraint_name) AS foreign_key_name,
  LOWER(a.column_name) AS column_name,
  LOWER(c_pk.table_name) AS ref_table_name,
  LOWER(b.column_name) AS ref_column_name
FROM user_cons_columns a
  JOIN user_constraints c ON a.owner = c.owner
       AND a.constraint_name = c.constraint_name
  JOIN user_constraints c_pk ON c.r_owner = c_pk.owner
       AND c.r_constraint_name = c_pk.constraint_name
  JOIN user_cons_columns b ON C_PK.owner = b.owner
       AND C_PK.CONSTRAINT_NAME = b.constraint_name AND b.POSITION = a.POSITION
WHERE c.constraint_type = 'R'
  AND a.owner = UPPER(%%schema string%%)
  AND a.table_name = UPPER(%%table string%%)
ENDSQL

# oracle table index list query
$XOBIN query $ORDB -M -B -2 -T Index -F OracleTableIndexes -a -o $DEST $@ << ENDSQL
SELECT
  LOWER(index_name) AS index_name,
  CASE WHEN uniqueness = 'UNIQUE' THEN '1' ELSE '0' END AS is_unique
FROM all_indexes
WHERE owner = UPPER(%%schema string%%)
  AND table_name = UPPER(%%table string%%)
ENDSQL

# oracle index column list query
$XOBIN query $ORDB -M -B -2 -T IndexColumn -F OracleIndexColumns -a -o $DEST $@ << ENDSQL
SELECT
  column_position AS seq_no,
  LOWER(column_name) AS column_name
FROM all_ind_columns
WHERE index_owner = UPPER(%%schema string%%)
  AND table_name = UPPER(%%table string%%)
  AND index_name = UPPER(%%index string%%)
ORDER BY column_position
ENDSQL
