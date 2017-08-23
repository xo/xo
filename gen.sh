#!/bin/bash

PGDB=pg://xodb:xodb@localhost/xodb
MYDB=my://xodb:xodb@localhost/xodb
MSDB=ms://xodb:xodb@localhost/xodb
SQDB=sq:xodb.sqlite3
ORDB=or://xodb:xodb@localhost/xe

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

# mysql enum list query
$XOBIN $MYDB -a -N -M -B -T Enum -F MyEnums -o $DEST $EXTRA << ENDSQL
SELECT
  DISTINCT column_name AS enum_name
FROM information_schema.columns
WHERE data_type = 'enum' AND table_schema = %%schema string%%
ENDSQL

# mysql enum value list query
$XOBIN $MYDB -N -M -B -1 -T MyEnumValue -F MyEnumValues -o $DEST $EXTRA << ENDSQL
SELECT
  SUBSTRING(column_type, 6, CHAR_LENGTH(column_type) - 6) AS enum_values
FROM information_schema.columns
WHERE data_type = 'enum' AND table_schema = %%schema string%% AND column_name = %%enum string%%
ENDSQL

# mysql autoincrement list query
$XOBIN $MYDB -N -M -B -T MyAutoIncrement -F MyAutoIncrements -o $DEST $EXTRA << ENDSQL
SELECT
  table_name
FROM information_schema.tables
WHERE auto_increment IS NOT null AND table_schema = %%schema string%%
ENDSQL

# mysql proc list query
$XOBIN $MYDB -a -N -M -B -T Proc -F MyProcs -o $DEST $EXTRA << ENDSQL
SELECT
  r.routine_name AS proc_name,
  p.dtd_identifier AS return_type
FROM information_schema.routines r
INNER JOIN information_schema.parameters p
  ON p.specific_schema = r.routine_schema AND p.specific_name = r.routine_name AND p.ordinal_position = 0
WHERE r.routine_schema = %%schema string%%
ENDSQL

# mysql proc parameter list query
$XOBIN $MYDB -a -N -M -B -T ProcParam -F MyProcParams -o $DEST $EXTRA << ENDSQL
SELECT
  dtd_identifier AS param_type
FROM information_schema.parameters
WHERE ordinal_position > 0 AND specific_schema = %%schema string%% AND specific_name = %%proc string%%
ORDER BY ordinal_position
ENDSQL

# mysql table list query
$XOBIN $MYDB -a -N -M -B -T Table -F MyTables -o $DEST $EXTRA << ENDSQL
SELECT
  table_name
FROM information_schema.tables
WHERE table_schema = %%schema string%% AND table_type = %%relkind string%%
ENDSQL

# mysql table column list query
$XOBIN $MYDB -a -N -M -B -T Column -F MyTableColumns -o $DEST $EXTRA << ENDSQL
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
$XOBIN $MYDB -a -N -M -B -T ForeignKey -F MyTableForeignKeys -o $DEST $EXTRA << ENDSQL
SELECT
  constraint_name AS foreign_key_name,
  column_name AS column_name,
  referenced_table_name AS ref_table_name,
  referenced_column_name AS ref_column_name
FROM information_schema.key_column_usage
WHERE referenced_table_name IS NOT NULL AND table_schema = %%schema string%% AND table_name = %%table string%%
ENDSQL

# mysql table index list query
$XOBIN $MYDB -a -N -M -B -T Index -F MyTableIndexes -o $DEST $EXTRA << ENDSQL
SELECT
  DISTINCT index_name,
  NOT non_unique AS is_unique
FROM information_schema.statistics
WHERE index_name <> 'PRIMARY' AND index_schema = %%schema string%% AND table_name = %%table string%%
ENDSQL

# mysql index column list query
$XOBIN $MYDB -a -N -M -B -T IndexColumn -F MyIndexColumns -o $DEST $EXTRA << ENDSQL
SELECT
  seq_in_index AS seq_no,
  column_name
FROM information_schema.statistics
WHERE index_schema = %%schema string%% AND table_name = %%table string%% AND index_name = %%index string%%
ORDER BY seq_in_index
ENDSQL

# sqlite autoincrement query
$XOBIN $SQDB -N -M -B -T SqAutoIncrement -F SqAutoIncrements -o $DEST $EXTRA << ENDSQL
SELECT
  name as table_name, sql
FROM sqlite_master
WHERE type='table'
ORDER BY name
ENDSQL

# sqlite table list query
$XOBIN $SQDB -a -N -M -B -T Table -F SqTables -o $DEST $EXTRA << ENDSQL
SELECT
  tbl_name AS table_name
FROM sqlite_master
WHERE tbl_name NOT LIKE 'sqlite_%' AND type = %%relkind string%%
ENDSQL

# sqlite table column list query
FIELDS='FieldOrdinal int,ColumnName string,DataType string,NotNull bool,DefaultValue sql.NullString,PkColIndex int'
$XOBIN $SQDB -I -N -M -B -T SqColumn -F SqTableColumns -Z "$FIELDS" -o $DEST $EXTRA << ENDSQL
PRAGMA table_info(%%table string,interpolate%%)
ENDSQL

# sqlite table foreign key list query
FIELDS='KeyID int,SeqNo int,RefTableName string,ColumnName string,RefColumnName string,OnUpdate string,OnDelete string,Match string'
$XOBIN $SQDB -a -I -N -M -B -T ForeignKey -F SqTableForeignKeys -Z "$FIELDS" -o $DEST $EXTRA << ENDSQL
PRAGMA foreign_key_list(%%table string,interpolate%%)
ENDSQL

# sqlite table index list query
FIELDS='SeqNo int,IndexName string,IsUnique bool,Origin string,IsPartial bool'
$XOBIN $SQDB -a -I -N -M -B -T Index -F SqTableIndexes -Z "$FIELDS" -o $DEST $EXTRA << ENDSQL
PRAGMA index_list(%%table string,interpolate%%)
ENDSQL

# sqlite index column list query
FIELDS='SeqNo int,Cid int,ColumnName string'
$XOBIN $SQDB -a -I -N -M -B -T IndexColumn -F SqIndexColumns -Z "$FIELDS" -o $DEST $EXTRA << ENDSQL
PRAGMA index_info(%%index string,interpolate%%)
ENDSQL

# mssql identity table list query
$XOBIN $MSDB -N -M -B -T MsIdentity -F MsIdentities -o $DEST $EXTRA << ENDSQL
SELECT o.name as table_name
FROM sys.objects o inner join sys.columns c on o.object_id = c.object_id
WHERE c.is_identity = 1
AND schema_name(o.schema_id) = %%schema string%% AND o.type = 'U'
ENDSQL

# mssql table list query
$XOBIN $MSDB -a -N -M -B -T Table -F MsTables -o $DEST $EXTRA << ENDSQL
SELECT
  xtype AS type,
  name AS table_name
FROM sysobjects
WHERE SCHEMA_NAME(uid) = %%schema string%% AND xtype = %%relkind string%%
ENDSQL

# mssql table column list query
$XOBIN $MSDB -a -N -M -B -T Column -F MsTableColumns -o $DEST $EXTRA << ENDSQL
SELECT
  c.colid AS field_ordinal,
  c.name AS column_name,
  TYPE_NAME(c.xtype)+IIF(c.prec > 0, '('+CAST(c.prec AS varchar)+IIF(c.scale > 0,','+CAST(c.scale AS varchar),'')+')', '') as data_type,
  IIF(c.isnullable=1, 0, 1) AS not_null,
  x.text AS default_value,
  IIF(COALESCE((
    SELECT count(z.colid)
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

# mssql table foreign key list query
$XOBIN $MSDB -a -N -M -B -T ForeignKey -F MsTableForeignKeys -o $DEST $EXTRA << ENDSQL
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

# mssql table index list query
$XOBIN $MSDB -a -N -M -B -T Index -F MsTableIndexes -o $DEST $EXTRA << ENDSQL
SELECT
  i.name AS index_name,
  i.is_primary_key AS is_primary,
  i.is_unique
FROM sys.indexes i
  INNER JOIN sysobjects o ON i.object_id = o.id
WHERE i.name IS NOT NULL AND o.type = 'U' AND SCHEMA_NAME(o.uid) = %%schema string%% AND o.name = %%table string%%
ENDSQL

# mssql index column list query
$XOBIN $MSDB -a -N -M -B -T IndexColumn -F MsIndexColumns -o $DEST $EXTRA << ENDSQL
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

# oracle proc list query
#$XOBIN $ORDB -a -N -M -B -T Proc -F OrProcs -o $DEST $EXTRA << ENDSQL
#SELECT
#  object_name
#FROM
#ENDSQL

# oracle proc parameter list query
#$XOBIN $ORDB -a -N -M -B -T ProcParam -F OrProcParams -o $DEST $EXTRA << ENDSQL
#SELECT
#ENDSQL

# oracle table list query
$XOBIN $ORDB -a -N -M -B -T Table -F OrTables -o $DEST $EXTRA << ENDSQL
SELECT
  LOWER(object_name) AS table_name
FROM all_objects
WHERE owner = UPPER(%%schema string%%) AND object_type = UPPER(%%relkind string%%)
  AND object_name NOT LIKE '%$%'
  AND object_name NOT LIKE 'LOGMNR%_%'
  AND object_name NOT LIKE 'REDO_%'
  AND object_name NOT LIKE 'SCHEDULER_%_TBL'
  AND object_name NOT LIKE 'SQLPLUS_%'
ENDSQL

# oracle table column list query
$XOBIN $ORDB -a -N -M -B -T Column -F OrTableColumns -o $DEST $EXTRA << ENDSQL
SELECT
  c.column_id AS field_ordinal,
  LOWER(c.column_name) AS column_name,
  LOWER(CASE c.data_type
          WHEN 'CHAR' THEN 'CHAR('||c.data_length||')'
          WHEN 'VARCHAR2' THEN 'VARCHAR2('||data_length||')'
          WHEN 'NUMBER' THEN
		    (CASE WHEN c.data_precision IS NULL AND c.data_scale IS NULL THEN 'NUMBER'
               ELSE 'NUMBER('||NVL(c.data_precision, 38)||','||NVL(c.data_scale, 0)||')' END)
          ELSE c.data_type END) AS data_type,
  CASE WHEN c.nullable = 'N' THEN '1' ELSE '0' END AS not_null,
  COALESCE((SELECT CASE WHEN r.constraint_type = 'P' THEN '1' ELSE '0' END
    FROM all_cons_columns l, all_constraints r
    WHERE r.constraint_type = 'P' AND r.owner = c.owner AND r.table_name = c.table_name AND r.constraint_name = l.constraint_name
    AND l.owner = c.owner AND l.table_name = c.table_name AND l.column_name = c.column_name), '0') AS is_primary_key
FROM all_tab_columns c
WHERE c.owner = UPPER(%%schema string%%) AND c.table_name = UPPER(%%table string%%)
ORDER BY c.column_id
ENDSQL

# oracle table foreign key list query
$XOBIN $ORDB -a -N -M -B -T ForeignKey -F OrTableForeignKeys -o $DEST $EXTRA << ENDSQL
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
$XOBIN $ORDB -a -N -M -B -T Index -F OrTableIndexes -o $DEST $EXTRA << ENDSQL
SELECT
  LOWER(index_name) AS index_name,
  CASE WHEN uniqueness = 'UNIQUE' THEN '1' ELSE '0' END AS is_unique
FROM all_indexes
WHERE owner = UPPER(%%schema string%%) AND table_name = UPPER(%%table string%%)
ENDSQL

# oracle index column list query
$XOBIN $ORDB -a -N -M -B -T IndexColumn -F OrIndexColumns -o $DEST $EXTRA << ENDSQL
SELECT
  column_position AS seq_no,
  LOWER(column_name) AS column_name
FROM all_ind_columns
WHERE index_owner = UPPER(%%schema string%%) AND table_name = UPPER(%%table string%%) AND index_name = UPPER(%%index string%%)
ORDER BY column_position
ENDSQL
