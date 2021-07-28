#!/bin/bash

xo query -o . -S model.xo.go -M -B -T PgType pg:// << END
SELECT
  DISTINCT typname::varchar AS name
FROM pg_type
WHERE typname NOT LIKE '\_%'
  AND typname <> 'pg_type'
END
