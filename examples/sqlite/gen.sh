#!/bin/bash

SRC=$(realpath $(cd -P "$( dirname "${BASH_SOURCE[0]}" )" && pwd ))
DBFILE=$SRC/booktest.sqlite3

XOBIN=$(which xo)
if [ -e $SRC/../../xo ]; then
  XOBIN=$SRC/../../xo
fi

set -x

rm -f $DBFILE

sqlite3 $DBFILE << 'ENDSQL'
PRAGMA foreign_keys = ON;

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
    year integer NOT NULL DEFAULT 2000
);

CREATE INDEX books_title_idx ON books(title, year);

ENDSQL

exit

$XOBIN sqlite:/$DBFILE -o $SRC/models

cat << ENDSQL | $XOBIN sqlite:/$DBFILE -N -M -B -T AuthorBookResult --query-type-comment='BookTag is the result of a search.' -o $SRC/models
SELECT
  CAST(a.author_id AS integer) AS author_id,
  CAST(a.name AS text) AS author_name,
  CAST(b.book_id AS integer) AS book_id,
  CAST(b.isbn AS text) AS book_isbn,
  CAST(b.title AS text) AS book_title
FROM books b
JOIN authors a ON a.author_id = b.author_id
WHERE b.title LIKE %%title string%%
ENDSQL

pushd $SRC &> /dev/null
go build
popd &> /dev/null
