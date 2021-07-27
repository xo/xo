WITH idx_names AS (
  SELECT
    LOWER(i.owner) AS owner,
    i.index_name AS index_name,
    LOWER(i.table_name) AS table_name,
    LISTAGG(LOWER(c.column_name), '_') AS column_names
  FROM all_indexes i
    JOIN all_ind_columns c
    ON i.index_name = c.index_name
  WHERE i.owner = 'DJANGO'
  GROUP BY i.owner, i.index_name, i.table_name
)
SELECT
  'ALTER INDEX ' || index_name || ' RENAME TO ' || table_name || '_' || column_names || '_idx' AS cmd
FROM idx_names
WHERE owner = 'django'
  AND index_name LIKE 'SYS_%'
\gexec
