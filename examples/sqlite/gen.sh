#!/bin/bash

DBNAME=booktest.sqlite3

EXTRA=$1

SRC=$(realpath $(cd -P "$( dirname "${BASH_SOURCE[0]}" )" && pwd ))

DB=file:$SRC/$DBNAME

XOBIN=$(which xo)
if [ -e $SRC/../../xo ]; then
  XOBIN=$SRC/../../xo
fi

DEST=$SRC/models

set -x

mkdir -p $DEST
rm -f $DEST/*.go
rm -f $SRC/sqlite
rm -f $SRC/$DBNAME

sqlite3 $DB << 'ENDSQL'
PRAGMA foreign_keys = 1;

CREATE TABLE authors (
  author_id integer NOT NULL PRIMARY KEY,
  name text NOT NULL DEFAULT ''
);

CREATE INDEX authors_name_idx ON authors(name);

CREATE TABLE books (
  book_id integer NOT NULL PRIMARY KEY,
  author_id integer NOT NULL REFERENCES authors(author_id),
  isbn text NOT NULL DEFAULT '' UNIQUE,
  title text NOT NULL DEFAULT '',
  year integer NOT NULL DEFAULT 2000,
  available text NOT NULL DEFAULT '',
  tags text NOT NULL DEFAULT '{}'
);

CREATE INDEX books_title_idx ON books(title, year);

ENDSQL

$XOBIN $DB -o $SRC/models $EXTRA

cat << ENDSQL | $XOBIN $DB -N -M -B -T AuthorBookResult --query-type-comment='AuthorBookResult is the result of a search.' -o $SRC/models $EXTRA
SELECT
  a.author_id,
  a.name AS author_name,
  b.book_id,
  b.isbn AS book_isbn,
  b.title AS book_title,
  b.tags AS book_tags
FROM books b
JOIN authors a ON a.author_id = b.author_id
WHERE b.tags LIKE '%' || %%tag string%% || '%'
ENDSQL

pushd $SRC &> /dev/null

go build
./sqlite $EXTRA

popd &> /dev/null

sqlite3 $DB << ENDSQL
.headers on
.mode column
.width 7 9 4 20 4 21 15
select * from books;
ENDSQL
