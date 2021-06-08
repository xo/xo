# About

This `booktest` directory contains the canonical `xo` example demonstrate an
end-to-end use of `xo`. Generates code from a simple schema and custom query
for each database. Additionally, showcases a practical use of generated Go
code.

This examples are also used by the `xo` developers to compare generated code
for/between databases and template revisions.

Contained in this directory is a subdirectory for each supported `<database>`
by `xo`:

| Database             | Generated Code          |
|----------------------|-------------------------|
| Microsoft SQL Server | [sqlserver](sqlserver/) |
| MySQL                | [mysql](mysql/)         |
| Oracle               | [oracle](oracle/)       |
| PostgreSQL           | [postgres](postgres/)   |
| SQLite3              | [sqlite3](sqlite3/)     |

Each database has a `sql/<name>_schema.sql` and `sql/<name>_query.sql`
containing a basic `authors` and `books` schema, and a custom retrieval query
the database.

See [`gen.sh`](gen.sh) to see how the various database model code was
generated.
