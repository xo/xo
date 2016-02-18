#!/bin/bash

DBNAME=booktest.sqlite3

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
rm -f $DEST/$DBNAME

sqlite3 $DB << 'ENDSQL'
PRAGMA foreign_keys = 1;

CREATE TABLE authors (
  author_id integer PRIMARY KEY,
  name text NOT NULL DEFAULT ''
);

CREATE INDEX authors_name_idx ON authors(name);

CREATE TABLE books (
  book_id integer PRIMARY KEY,
  author_id integer NOT NULL REFERENCES authors(author_id),
  isbn text NOT NULL DEFAULT '' UNIQUE,
  title text NOT NULL DEFAULT '',
  year integer NOT NULL DEFAULT 2000,
  available text NOT NULL DEFAULT '',
  tags text NOT NULL DEFAULT '{}'
);

CREATE INDEX books_title_idx ON books(title, year);

ENDSQL


$XOBIN $DB -v -o $SRC/models

cat << ENDSQL | $XOBIN $DB -v -N -M -B -T AuthorBookResult --query-type-comment='AuthorBookResult is the result of a search.' -o $SRC/models
SELECT
  a.author_id,
  a.name AS author_name,
  b.book_id,
  b.isbn AS book_isbn,
  b.title AS book_title,
  b.tags AS book_tags
FROM books b
JOIN authors a ON a.author_id = b.author_id
WHERE LIKE(b.tags, '%' || %%tag StringSlice%% || '%')
ENDSQL

pushd $SRC &> /dev/null

go build
./sqlite

popd &> /dev/null

sqlite3 $DB <<< 'select * from books;'
