WITH pk_names AS (
  SELECT
    name AS index_name,
    OBJECT_NAME(object_id) AS table_name,
    OBJECT_SCHEMA_NAME(object_id) AS schema_name,
    (SELECT
      CONCAT('', c.name)
    FROM sys.index_columns ic
    INNER JOIN sys.columns c ON ic.object_id = c.object_id
      AND ic.index_column_id = c.column_id
    WHERE ic.object_id = i.object_id
      AND ic.index_id = i.index_id
      AND ic.is_included_column = 0
    ORDER BY ic.key_ordinal
    FOR XML PATH('')) COLLATE SQL_Latin1_General_CP1_CI_AS AS column_names
  FROM sys.indexes i
  WHERE i.is_primary_key = 1
)
SELECT
  CONCAT(
    'EXEC sp_rename ''',
    QUOTENAME(schema_name),
    '.',
    QUOTENAME(index_name),
    ''', ''',
    table_name,
    '_',
    column_names,
    '_pkey'''
  )
FROM pk_names
WHERE index_name LIKE 'PK_%'
\gexec
