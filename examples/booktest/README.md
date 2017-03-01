# About booktest examples

This directory contains the `xo` "booktest" examples which are used not only as
examples, but also in order to compare the generated code for the various
database loaders.

Each subdirectory contains a single, almost identical example for each
supported database loader, demonstrating how to write and use a schema for
each database, and the resulting, generated "model" code (contained in each
subdirectory's `models/` folder).

## Available Examples

* [Microsoft SQL Server](mssql/)
* [MySQL](mysql/)
* [Oracle](oracle/)
* [PostgreSQL](postgres/)
* [SQLite3](sqlite3/)

## About booktest

Each `booktest` example tries to showcase all the generated types and funcs for
each of the above supported databases, and the feature set that they/`xo`
supports.

The `booktest` example schemas try to show a standard "author / book" starter
schema, wherein there are multiple books per author. Additionally, each
database might have schema definitions for things like stored functions /
procedures.

These examples are meant to show how to hit the ground running with `xo`, and
are not meant to be an exhaustive demonstration of each database's schema
support.

### Running gen.sh

Each example has a `gen.sh` that looks in the root repository path for a built
`xo` executable. If it is not present, then it looks on `$PATH` for `xo`. The
`gen.sh` script uses that version of `xo` for generating the `models/`
subdirectory.

#### Command line options for gen.sh

`gen.sh` will pass any options to `xo`, as well as the resulting, built
`booktest` executable. At the moment, the most useful command line option is
passing `-v`, which will enable verbose output both for `xo` and the built
executable.
