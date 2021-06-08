#!/bin/bash

PGDB=pg://
MYDB=my://localhost/mysql
MSDB=ms://
SQDB=sq:xo.db
ORDB=or://localhost/orasid

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

# postgres schema query
COMMENT='PostgresSchema retrieves the current schema.'
$XOBIN query $PGDB -M -B -l -F PostgresSchema --func-comment "$COMMENT" --single=models.xo.go -o $DEST $@ << ENDSQL
SELECT CURRENT_SCHEMA()::text AS schema_name
ENDSQL

# postgres enum list query
COMMENT='Enum represents a enum.'
$XOBIN query $PGDB -M -B -2 -T Enum -F PostgresEnums --type-comment "$COMMENT" -o $DEST $@ << ENDSQL
SELECT
  t.typname::varchar AS enum_name
FROM pg_type t
  JOIN ONLY pg_namespace n ON n.oid = t.typnamespace
  JOIN ONLY pg_enum e ON t.oid = e.enumtypid
WHERE n.nspname = %%schema string%%
ENDSQL

# postgres enum value list query
COMMENT='EnumValue represents a enum value.'
$XOBIN query $PGDB -M -B -2 -T EnumValue -F PostgresEnumValues --type-comment "$COMMENT" -o $DEST $@ << ENDSQL
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
$XOBIN query $PGDB -M -B -2 -T Sequence -F PostgresSequences -o $DEST $@ << ENDSQL
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
$XOBIN query $PGDB -M -B -2 -T Proc -F PostgresProcs --type-comment "$COMMENT" -o $DEST $@ << ENDSQL
SELECT
  p.proname::varchar AS proc_name,
  pg_get_function_result(p.oid)::varchar AS return_type
FROM pg_proc p
  JOIN ONLY pg_namespace n ON p.pronamespace = n.oid
WHERE n.nspname = %%schema string%%
ENDSQL

# postgres proc parameter list query
COMMENT='ProcParam represents a stored procedure param.'
$XOBIN query $PGDB -M -B -2 -T ProcParam -F PostgresProcParams --type-comment "$COMMENT" -o $DEST $@ << ENDSQL
SELECT
  UNNEST(STRING_TO_ARRAY(oidvectortypes(p.proargtypes), ', '))::varchar AS param_type
FROM pg_proc p
  JOIN ONLY pg_namespace n ON p.pronamespace = n.oid
WHERE n.nspname = %%schema string%% AND p.proname = %%proc string%%
ENDSQL

# postgres table list query
COMMENT='Table represents table info.'
$XOBIN query $PGDB -M -B -2 -T Table -F PostgresTables --type-comment "$COMMENT" -o $DEST $@ << ENDSQL
SELECT
  c.relkind::varchar AS type,
  c.relname::varchar AS table_name,
  false::boolean AS manual_pk
FROM pg_class c
  JOIN ONLY pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname = %%schema string%% AND c.relkind = %%kind string%%
ENDSQL

# postgres table column list query
FIELDS='FieldOrdinal int,ColumnName string,DataType string,NotNull bool,DefaultValue sql.NullString,IsPrimaryKey bool'
COMMENT='Column represents column info.'
$XOBIN query $PGDB -M -B -2 -T Column -F PostgresTableColumns -Z "$FIELDS" --type-comment "$COMMENT" -o $DEST $@ << ENDSQL
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
$XOBIN query $PGDB -M -B -2 -T ForeignKey -F PostgresTableForeignKeys --type-comment "$COMMENT" -o $DEST $@ << ENDSQL
SELECT
  r.conname::varchar AS foreign_key_name,
  b.attname::varchar AS column_name,
  i.relname::varchar AS ref_index_name,
  c.relname::varchar AS ref_table_name,
  d.attname::varchar AS ref_column_name,
  0::integer AS key_id,
  0::integer AS seq_no
FROM pg_constraint r
  JOIN ONLY pg_class a ON a.oid = r.conrelid
  JOIN ONLY pg_attribute b ON b.attisdropped = false AND b.attnum = ANY(r.conkey) AND b.attrelid = r.conrelid
  JOIN ONLY pg_class i ON i.oid = r.conindid
  JOIN ONLY pg_class c ON c.oid = r.confrelid
  JOIN ONLY pg_attribute d ON d.attisdropped = false AND d.attnum = ANY(r.confkey) AND d.attrelid = r.confrelid
  JOIN ONLY pg_namespace n ON n.oid = r.connamespace
WHERE r.contype = 'f' AND n.nspname = %%schema string%% AND a.relname = %%table string%%
ORDER BY r.conname, b.attname
ENDSQL

# postgres table index list query
COMMENT='Index represents an index.'
$XOBIN query $PGDB -M -B -2 -T Index -F PostgresTableIndexes --type-comment "$COMMENT" -o $DEST $@ << ENDSQL
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
$XOBIN query $PGDB -M -B -2 -T IndexColumn -F PostgresIndexColumns --type-comment "$COMMENT" -o $DEST $@ << ENDSQL
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
COMMENT='PostgresColOrder represents index column order.'
$XOBIN query $PGDB -M -B -1 -2 -T PostgresColOrder -F PostgresGetColOrder --type-comment "$COMMENT" -o $DEST $@ << ENDSQL
SELECT
  i.indkey::varchar AS ord
FROM pg_index i
  JOIN ONLY pg_class c ON c.oid = i.indrelid
  JOIN ONLY pg_namespace n ON n.oid = c.relnamespace
  JOIN ONLY pg_class ic ON ic.oid = i.indexrelid
WHERE n.nspname = %%schema string%% AND ic.relname = %%index string%%
ENDSQL

# mysql schema query
COMMENT='MysqlSchema retrieves the current schema.'
$XOBIN query $MYDB -M -B -l -F MysqlSchema --func-comment "$COMMENT" --single=models.xo.go -a -o $DEST $@ << ENDSQL
SELECT SCHEMA() AS schema_name
ENDSQL

# mysql enum list query
$XOBIN query $MYDB -M -B -2 -T Enum -F MysqlEnums -a -o $DEST $@ << ENDSQL
SELECT
  DISTINCT column_name AS enum_name
FROM information_schema.columns
WHERE data_type = 'enum' AND table_schema = %%schema string%%
ENDSQL

# mysql enum value list query
$XOBIN query $MYDB -M -B -1 -2 -T MysqlEnumValue -F MysqlEnumValues -o $DEST $@ << ENDSQL
SELECT
  SUBSTRING(column_type, 6, CHAR_LENGTH(column_type) - 6) AS enum_values
FROM information_schema.columns
WHERE data_type = 'enum' AND table_schema = %%schema string%% AND column_name = %%enum string%%
ENDSQL

# mysql sequence list query
$XOBIN query $MYDB -M -B -2 -T Sequence -F MysqlSequences -a -o $DEST $@ << ENDSQL
SELECT
  table_name
FROM information_schema.tables
WHERE auto_increment IS NOT NULL AND table_schema = %%schema string%%
ENDSQL

# mysql proc list query
$XOBIN query $MYDB -M -B -2 -T Proc -F MysqlProcs -a -o $DEST $@ << ENDSQL
SELECT
  r.routine_name AS proc_name,
  p.dtd_identifier AS return_type
FROM information_schema.routines r
INNER JOIN information_schema.parameters p
  ON p.specific_schema = r.routine_schema AND p.specific_name = r.routine_name AND p.ordinal_position = 0
WHERE r.routine_schema = %%schema string%%
ENDSQL

# mysql proc parameter list query
$XOBIN query $MYDB -M -B -2 -T ProcParam -F MysqlProcParams -a -o $DEST $@ << ENDSQL
SELECT
  dtd_identifier AS param_type
FROM information_schema.parameters
WHERE ordinal_position > 0 AND specific_schema = %%schema string%% AND specific_name = %%proc string%%
ORDER BY ordinal_position
ENDSQL

# mysql table list query
$XOBIN query $MYDB -M -B -2 -T Table -F MysqlTables -a -o $DEST $@ << ENDSQL
SELECT
  table_name
FROM information_schema.tables
WHERE table_schema = %%schema string%% AND table_type = %%kind string%%
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
WHERE table_schema = %%schema string%% AND table_name = %%table string%%
ORDER BY ordinal_position
ENDSQL

# mysql table foreign key list query
$XOBIN query $MYDB -M -B -2 -T ForeignKey -F MysqlTableForeignKeys -a -o $DEST $@ << ENDSQL
SELECT
  constraint_name AS foreign_key_name,
  column_name AS column_name,
  referenced_table_name AS ref_table_name,
  referenced_column_name AS ref_column_name
FROM information_schema.key_column_usage
WHERE referenced_table_name IS NOT NULL AND table_schema = %%schema string%% AND table_name = %%table string%%
ENDSQL

# mysql table index list query
$XOBIN query $MYDB -M -B -2 -T Index -F MysqlTableIndexes -a -o $DEST $@ << ENDSQL
SELECT
  DISTINCT index_name,
  NOT non_unique AS is_unique
FROM information_schema.statistics
WHERE index_name <> 'PRIMARY' AND index_schema = %%schema string%% AND table_name = %%table string%%
ENDSQL

# mysql index column list query
$XOBIN query $MYDB -M -B -2 -T IndexColumn -F MysqlIndexColumns -a -o $DEST $@ << ENDSQL
SELECT
  seq_in_index AS seq_no,
  column_name
FROM information_schema.statistics
WHERE index_schema = %%schema string%% AND table_name = %%table string%% AND index_name = %%index string%%
ORDER BY seq_in_index
ENDSQL

# sqlite3 sequence list query
$XOBIN query $SQDB -M -B -2 -T Sequence -F Sqlite3Sequences -a -o $DEST $@ << ENDSQL
SELECT
  name AS table_name
FROM sqlite_master
WHERE type='table' AND LOWER(sql) LIKE '%autoincrement%'
ORDER BY name
ENDSQL

# sqlite3 table list query
$XOBIN query $SQDB -M -B -2 -T Table -F Sqlite3Tables -a -o $DEST $@ << ENDSQL
SELECT
  tbl_name AS table_name
FROM sqlite_master
WHERE tbl_name NOT LIKE 'sqlite_%' AND type = %%kind string%%
ENDSQL

# sqlite3 table column list query
$XOBIN query $SQDB -M -B -2 -T Column -F Sqlite3TableColumns -a -o $DEST $@ << ENDSQL
SELECT
  cid AS field_ordinal,
  name AS column_name,
  type AS data_type,
  "notnull" AS not_null,
  dflt_value AS default_value,
  CAST(pk <> 0 AS boolean) AS is_primary_key
FROM pragma_table_info(%%table string%%)
ENDSQL

# sqlite3 table foreign key list query
$XOBIN query $SQDB -M -B -2 -T ForeignKey -F Sqlite3TableForeignKeys -a -o $DEST $@ << ENDSQL
SELECT
  id AS key_id,
  seq AS seq_no,
  "table" AS ref_table_name,
  "from" AS column_name,
  "to" AS ref_column_name
FROM pragma_foreign_key_list(%%table string%%)
ENDSQL

# sqlite3 table index list query
$XOBIN query $SQDB -M -B -2 -T Index -F Sqlite3TableIndexes -a -o $DEST $@ << ENDSQL
SELECT
  seq AS seq_no,
  name AS index_name,
  "unique" AS is_unique,
  origin,
  partial AS is_partial
FROM pragma_index_list(%%table string%%)
ENDSQL

# sqlite3 index column list query
$XOBIN query $SQDB -M -B -2 -T IndexColumn -F Sqlite3IndexColumns -a -o $DEST $@ << ENDSQL
SELECT
  seqno AS seq_no,
  cid,
  name AS column_name
FROM pragma_index_info(%%index string%%)
ENDSQL

# sqlserver schema query
COMMENT='SqlserverSchema retrieves the current schema.'
$XOBIN query $MSDB -M -B -l -F SqlserverSchema --func-comment "$COMMENT" --single=models.xo.go -a -o $DEST $@ << ENDSQL
SELECT SCHEMA_NAME() AS schema_name
ENDSQL

# sqlserver identity table list query
$XOBIN query $MSDB -M -B -2 -T SqlserverIdentity -F SqlserverIdentities -o $DEST $@ << ENDSQL
SELECT o.name AS table_name
FROM sys.objects o
  INNER JOIN sys.columns c ON o.object_id = c.object_id
WHERE c.is_identity = 1 AND SCHEMA_NAME(o.schema_id) = %%schema string%% AND o.type = 'U'
ENDSQL

# sqlserver table list query
$XOBIN query $MSDB -M -B -2 -T Table -F SqlserverTables -a -o $DEST $@ << ENDSQL
SELECT
  xtype AS type,
  name AS table_name
FROM sysobjects
WHERE SCHEMA_NAME(uid) = %%schema string%% AND xtype = %%kind string%%
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
      INNER JOIN sysindexkeys z ON i.id = z.id AND i.indid = z.indid AND z.colid = c.colid
    WHERE i.id = o.id AND i.name = k.name
  ), 0) > 0, 1, 0) AS is_primary_key
FROM syscolumns c
  JOIN sysobjects o ON o.id = c.id
  LEFT JOIN sysobjects k ON k.xtype='PK' AND k.parent_obj = o.id
  LEFT JOIN syscomments x ON x.id = c.cdefault
WHERE o.type IN('U', 'V') AND SCHEMA_NAME(o.uid) = %%schema string%% AND o.name = %%table string%%
ORDER BY c.colid
ENDSQL

# sqlserver table foreign key list query
$XOBIN query $MSDB -M -B -2 -T ForeignKey -F SqlserverTableForeignKeys -a -o $DEST $@ << ENDSQL
SELECT
  f.name AS foreign_key_name,
  c.name AS column_name,
  o.name AS ref_table_name,
  x.name AS ref_column_name
FROM sysobjects f
  INNER JOIN sysobjects t ON f.parent_obj = t.id
  INNER JOIN sysreferences r ON f.id = r.constid
  INNER JOIN sysobjects o ON r.rkeyid = o.id
  INNER JOIN syscolumns c ON r.rkeyid = c.id AND r.rkey1 = c.colid
  INNER JOIN syscolumns x ON r.fkeyid = x.id AND r.fkey1 = x.colid
WHERE f.type = 'F' AND t.type = 'U' AND SCHEMA_NAME(t.uid) = %%schema string%% AND t.name = %%table string%%
ENDSQL

# sqlserver table index list query
$XOBIN query $MSDB -M -B -2 -T Index -F SqlserverTableIndexes -a -o $DEST $@ << ENDSQL
SELECT
  i.name AS index_name,
  i.is_primary_key AS is_primary,
  i.is_unique
FROM sys.indexes i
  INNER JOIN sysobjects o ON i.object_id = o.id
WHERE i.name IS NOT NULL AND o.type = 'U' AND SCHEMA_NAME(o.uid) = %%schema string%% AND o.name = %%table string%%
ENDSQL

# sqlserver index column list query
$XOBIN query $MSDB -M -B -2 -T IndexColumn -F SqlserverIndexColumns -a -o $DEST $@ << ENDSQL
SELECT
  k.keyno AS seq_no,
  k.colid AS cid,
  c.name AS column_name
FROM sysindexes i
  INNER JOIN sysobjects o ON i.id = o.id
  INNER JOIN sysindexkeys k ON k.id = o.id AND k.indid = i.indid
  INNER JOIN syscolumns c ON c.id = o.id AND c.colid = k.colid
WHERE o.type = 'U' AND SCHEMA_NAME(o.uid) = %%schema string%% AND o.name = %%table string%% AND i.name = %%index string%%
ORDER BY k.keyno
ENDSQL

# oracle schema query
COMMENT='OracleSchema retrieves the current schema.'
$XOBIN query $ORDB -M -B -l -F OracleSchema --func-comment "$COMMENT" --single=models.xo.go -a -o $DEST $@ << ENDSQL
SELECT LOWER(SYS_CONTEXT('USERENV', 'CURRENT_SCHEMA')) AS schema_name FROM dual
ENDSQL

# oracle proc list query
#$XOBIN query $ORDB -M -B -2 -T Proc -F OracleProcs -a -o $DEST $@ << ENDSQL
#SELECT
#  object_name
#FROM
#ENDSQL

# oracle proc parameter list query
#$XOBIN query $ORDB -M -B -2 -T ProcParam -F OracleProcParams -a -o $DEST $@ << ENDSQL
#SELECT
#ENDSQL

# oracle table list query
$XOBIN query $ORDB -M -B -2 -T Table -F OracleTables -a -o $DEST $@ << ENDSQL
SELECT
  LOWER(object_name) AS table_name
FROM all_objects
WHERE
  owner = UPPER(%%schema string%%)
  AND object_type = UPPER(%%kind string%%)
  AND object_name NOT LIKE '%$%'
  AND object_name NOT LIKE 'LOGMNR%_%'
  AND object_name NOT LIKE 'REDO_%'
  AND object_name NOT LIKE 'SCHEDULER_%_TBL'
  AND object_name NOT LIKE 'SQLPLUS_%'
ENDSQL

# oracle table column list query
$XOBIN query $ORDB -M -B -2 -T Column -F OracleTableColumns -a -o $DEST $@ << ENDSQL
SELECT
  c.column_id AS field_ordinal,
  LOWER(c.column_name) AS column_name,
  LOWER(CASE c.data_type
    WHEN 'CHAR' THEN 'CHAR(' || c.data_length || ')'
    WHEN 'VARCHAR2' THEN 'VARCHAR2(' || data_length || ')'
    WHEN 'NUMBER' THEN
      (CASE WHEN c.data_precision IS NULL AND c.data_scale IS NULL THEN 'NUMBER'
        ELSE 'NUMBER(' || NVL(c.data_precision, 38) || ',' || NVL(c.data_scale, 0) || ')' END)
    ELSE c.data_type END) AS data_type,
  CASE WHEN c.nullable = 'N' THEN '1' ELSE '0' END AS not_null,
  COALESCE((
    SELECT CASE WHEN r.constraint_type = 'P' THEN '1' ELSE '0' END
    FROM all_cons_columns l, all_constraints r
    WHERE r.constraint_type = 'P'
      AND r.owner = c.owner AND r.table_name = c.table_name AND r.constraint_name = l.constraint_name
      AND l.owner = c.owner AND l.table_name = c.table_name AND l.column_name = c.column_name
  ), '0') AS is_primary_key
FROM all_tab_columns c
WHERE c.owner = UPPER(%%schema string%%) AND c.table_name = UPPER(%%table string%%)
ORDER BY c.column_id
ENDSQL

# oracle table foreign key list query
$XOBIN query $ORDB -M -B -2 -T ForeignKey -F OracleTableForeignKeys -a -o $DEST $@ << ENDSQL
SELECT
  LOWER(a.constraint_name) AS foreign_key_name,
  LOWER(a.column_name) AS column_name,
  LOWER(r.constraint_name) AS ref_index_name,
  LOWER(r.table_name) AS ref_table_name
FROM all_cons_columns a
  JOIN all_constraints c ON a.owner = c.owner AND a.constraint_name = c.constraint_name
  JOIN all_constraints r ON c.r_owner = r.owner AND c.r_constraint_name = r.constraint_name
  WHERE c.constraint_type = 'R' AND a.owner = UPPER(%%schema string%%) AND a.table_name = UPPER(%%table string%%)
ENDSQL

# oracle table index list query
$XOBIN query $ORDB -M -B -2 -T Index -F OracleTableIndexes -a -o $DEST $@ << ENDSQL
SELECT
  LOWER(index_name) AS index_name,
  CASE WHEN uniqueness = 'UNIQUE' THEN '1' ELSE '0' END AS is_unique
FROM all_indexes
WHERE owner = UPPER(%%schema string%%) AND table_name = UPPER(%%table string%%)
ENDSQL

# oracle index column list query
$XOBIN query $ORDB -M -B -2 -T IndexColumn -F OracleIndexColumns -a -o $DEST $@ << ENDSQL
SELECT
  column_position AS seq_no,
  LOWER(column_name) AS column_name
FROM all_ind_columns
WHERE index_owner = UPPER(%%schema string%%) AND table_name = UPPER(%%table string%%) AND index_name = UPPER(%%index string%%)
ORDER BY column_position
ENDSQL
