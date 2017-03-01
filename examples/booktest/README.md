# About booktest examples #

This directory contains the `xo` "booktest" examples which are used not only as
examples, but also in order to compare the generated code for the various
database loaders.

Each subdirectory contains a single, almost identical example for each
supported database loader, demonstrating how to write and use a schema for
each database, and the resulting, generated "model" code (contained in each
subdirectory's `models/` folder).

## Available Database Examples ##

* [Microsoft SQL Server](mssql/)
* [MySQL](mysql/)
* [Oracle](oracle/)
* [PostgreSQL](postgres/)
* [SQLite3](sqlite3/)

## About booktest ##

Each `booktest` example tries to showcase all the generated types and funcs for
each of the above supported databases, and the feature set that each database
and `xo` supports, as well as being an example of how one would use `xo` in an
end-to-end way from database to generated Go code.

The examples schemas used for the `booktest` try to show a how to do the
standard "author / book" schema example, where there are multiple books per
author. Additionally, each database schema may have definitions unique to it,
such as stored functions / procedures.

These examples are meant to show how to hit the ground running with `xo`, and
are not meant to be an exhaustive demonstration of each database's schema
support.

### Running gen.sh

Each example has a `gen.sh` that looks in the root repository path for a built
`xo` executable. If it is not present, then it looks on `$PATH` for `xo`. The
`gen.sh` script uses that version of `xo` for generating the `models/`
subdirectory.

#### Example Output

The following is the output from running the PostgreSQL booktest:

```sh
$ ./gen.sh
+ mkdir -p /home/ken/src/go/src/github.com/knq/xo/examples/booktest/postgres/models
+ rm -f /home/ken/src/go/src/github.com/knq/xo/examples/booktest/postgres/models/authorbookresult.xo.go /home/ken/src/go/src/github.com/knq/xo/examples/booktest/postgres/models/author.xo.go /home/ken/src/go/src/github.com/knq/xo/examples/booktest/postgres/models/booktype.xo.go /home/ken/src/go/src/github.com/knq/xo/examples/booktest/postgres/models/book.xo.go /home/ken/src/go/src/github.com/knq/xo/examples/booktest/postgres/models/sp_sayhello.xo.go /home/ken/src/go/src/github.com/knq/xo/examples/booktest/postgres/models/xo_db.xo.go
+ rm -f /home/ken/src/go/src/github.com/knq/xo/examples/booktest/postgres/postgres
+ psql -U postgres -c 'create user booktest password '\''booktest'\'';'
ERROR:  role "booktest" already exists
+ psql -U postgres -c 'drop database booktest;'
DROP DATABASE
+ psql -U postgres -c 'create database booktest owner booktest;'
CREATE DATABASE
+ psql -U booktest
CREATE TABLE
CREATE INDEX
CREATE TYPE
CREATE TABLE
CREATE INDEX
CREATE FUNCTION
CREATE INDEX
+ /home/ken/src/go/bin/xo postgres://booktest:booktest@localhost/booktest -o /home/ken/src/go/src/github.com/knq/xo/examples/booktest/postgres/models -j
+ /home/ken/src/go/bin/xo postgres://booktest:booktest@localhost/booktest -N -M -B -T AuthorBookResult '--query-type-comment=AuthorBookResult is the result of a search.' -o /home/ken/src/go/src/github.com/knq/xo/examples/booktest/postgres/models
+ pushd /home/ken/src/go/src/github.com/knq/xo/examples/booktest/postgres
+ go build
+ ./postgres
Book 1 (FICTION): my book title available: 01 Mar 17 10:54 +0700
Book 1 author: Unknown Master
---------
Tag search results:
Book 4: 'never ever gonna finish, a quatrain', Author: 'Unknown Master', ISBN: 'NEW ISBN' Tags: '[someother]'
Book 3: 'the third book', Author: 'Unknown Master', ISBN: '3' Tags: '[cool]'
Book 2: 'changed second title', Author: 'Unknown Master', ISBN: '2' Tags: '[cool disastor]'
SayHello response: hello john
+ popd
+ psql -U booktest
 book_id | author_id | isbn | booktype |        title         | year |           available           |      tags
---------+-----------+------+----------+----------------------+------+-------------------------------+-----------------
       1 |         1 | 1    | FICTION  | my book title        | 2016 | 2017-03-01 10:54:51.888826+07 | {}
       2 |         1 | 2    | FICTION  | changed second title | 2016 | 2017-03-01 10:54:51.888826+07 | {cool,disastor}
       3 |         1 | 3    | FICTION  | the third book       | 2001 | 2017-03-01 10:54:51.888826+07 | {cool}
(3 rows)
```

**NOTE:** The output of each database's `gen.sh` should have similar output,
but might vary slightly.

#### Command line options for gen.sh

`gen.sh` will pass any options to `xo`, as well as the resulting, built
`booktest` executable. At the moment, the most useful command line option is
passing `-v`, which will enable verbose output both for `xo` and the built
executable.
