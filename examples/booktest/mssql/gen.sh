#!/bin/bash

SAUSER=sa
SAPASS=changeit

DBUSER=booktest
DBPASS=booktest
DBHOST=localhost
DBNAME=booktest

DB=mssql://$DBUSER:$DBPASS@$DBHOST/$DBNAME

EXTRA=$1

SRC=$(realpath $(cd -P "$( dirname "${BASH_SOURCE[0]}" )" && pwd ))

XOBIN=$(which xo)
if [ -e $SRC/../../../xo ]; then
  XOBIN=$SRC/../../../xo
fi

DEST=$SRC/models

set -x

mkdir -p $DEST
rm -f $DEST/*.go
rm -f $SRC/mssql

mssql -u $SAUSER -p $SAPASS -s $DBHOST -q "drop database $DBNAME;"

$SRC/../../../contrib/mscreate.sh $DBNAME

SQLFILE=$(mktemp /tmp/mssql.XXXXXX.sql)
cat > $SQLFILE << ENDSQL
CREATE TABLE authors (
  author_id integer NOT NULL IDENTITY(1,1) PRIMARY KEY,
  name varchar(255) NOT NULL DEFAULT ''
);

CREATE INDEX authors_name_idx ON authors(name);

CREATE TABLE books (
  book_id integer NOT NULL IDENTITY(1,1) PRIMARY KEY,
  author_id integer NOT NULL FOREIGN KEY REFERENCES authors(author_id),
  isbn varchar(255) NOT NULL DEFAULT '' UNIQUE,
  title varchar(255) NOT NULL DEFAULT '',
  year integer NOT NULL DEFAULT 2000,
  available datetime2 NOT NULL,
  tags varchar(255) NOT NULL DEFAULT ''
);

CREATE INDEX books_title_idx ON books(title, year);

ENDSQL

mssql -u $DBUSER -p $DBPASS -s $DBHOST -d $DBNAME -q ".run $SQLFILE"

rm -f $SQLFILE

$XOBIN $DB -o $SRC/models $EXTRA

$XOBIN $DB -N -M -B -T AuthorBookResult --query-type-comment='AuthorBookResult is the result of a search.' -o $SRC/models $EXTRA << ENDSQL
SELECT
  a.author_id AS author_id,
  a.name AS author_name,
  b.book_id AS book_id,
  b.isbn AS book_isbn,
  b.title AS book_title,
  b.tags AS book_tags
FROM books b
JOIN authors a ON a.author_id = b.author_id
WHERE b.tags LIKE '%' + %%tags string%% + '%'
ENDSQL

pushd $SRC &> /dev/null

go build
./mssql $EXTRA

popd &> /dev/null

mssql -u $DBUSER -p $DBPASS -s $DBHOST -d $DBNAME -q 'select * from books;'
