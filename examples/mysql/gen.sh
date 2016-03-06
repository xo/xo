#!/bin/bash

DBUSER=booktest
DBPASS=booktest
DBHOST=localhost
DBNAME=booktest

DB=mysql://$DBUSER:$DBPASS@$DBHOST/$DBNAME

EXTRA=$1

SRC=$(realpath $(cd -P "$( dirname "${BASH_SOURCE[0]}" )" && pwd ))

XOBIN=$(which xo)
if [ -e $SRC/../../xo ]; then
  XOBIN=$SRC/../../xo
fi

DEST=$SRC/models

set -x

mkdir -p $DEST
rm -f $DEST/*.go
rm -f $SRC/mysql

mysql -u $DBUSER -p$DBPASS $DBNAME << 'ENDSQL'
SET FOREIGN_KEY_CHECKS=0;
DROP TABLE IF EXISTS authors;
DROP TABLE IF EXISTS books;
DROP FUNCTION IF EXISTS say_hello;
SET FOREIGN_KEY_CHECKS=1;

CREATE TABLE authors (
  author_id integer NOT NULL AUTO_INCREMENT PRIMARY KEY,
  name text NOT NULL DEFAULT ''
) ENGINE=InnoDB;

CREATE INDEX authors_name_idx ON authors(name(255));

CREATE TABLE books (
  book_id integer NOT NULL AUTO_INCREMENT PRIMARY KEY,
  author_id integer NOT NULL,
  isbn varchar(255) NOT NULL DEFAULT '' UNIQUE,
  book_type ENUM('FICTION', 'NONFICTION') NOT NULL DEFAULT 'FICTION',
  title text NOT NULL DEFAULT '',
  year integer NOT NULL DEFAULT 2000,
  available timestamp NOT NULL DEFAULT NOW(),
  tags text NOT NULL DEFAULT '',
  CONSTRAINT FOREIGN KEY (author_id) REFERENCES authors(author_id)
) ENGINE=InnoDB;

CREATE INDEX books_title_idx ON books(title(255), year);

CREATE FUNCTION say_hello(s text) RETURNS text
  DETERMINISTIC
  RETURN CONCAT('hello ', s);
ENDSQL

$XOBIN $DB -o $SRC/models -j $EXTRA

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
WHERE b.tags LIKE CONCAT('%', %%tag string%%, '%')
ENDSQL

pushd $SRC &> /dev/null

go build
./mysql $EXTRA

popd &> /dev/null

mysql -u $DBUSER -p$DBPASS $DBNAME -e 'select * from books;'
