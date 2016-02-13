#!/bin/bash

SRC=$(realpath $(cd -P "$( dirname "${BASH_SOURCE[0]}" )" && pwd ))

XOBIN=$(which xo)
if [ -e $SRC/../../xo ]; then
  XOBIN=$SRC/../../xo
fi

set -ex

mysql -u booktest -pbooktest booktest << 'ENDSQL'
SET FOREIGN_KEY_CHECKS=0;
DROP TABLE IF EXISTS authors;
DROP TABLE IF EXISTS books;
DROP FUNCTION IF EXISTS say_hello;
SET FOREIGN_KEY_CHECKS=1;

CREATE TABLE authors (
  author_id integer NOT NULL AUTO_INCREMENT PRIMARY KEY,
  name varchar(255) NOT NULL DEFAULT ''
) ENGINE=InnoDB;

CREATE INDEX authors_name_idx ON authors(name);

CREATE TABLE books (
  book_id integer NOT NULL AUTO_INCREMENT PRIMARY KEY,
  author_id integer NOT NULL,
  isbn varchar(255) NOT NULL DEFAULT '' UNIQUE,
  booktype ENUM('FICTION', 'NONFICTION') NOT NULL DEFAULT 'FICTION',
  title varchar(255) DEFAULT '',
  year integer NOT NULL DEFAULT 2000,
  available timestamp NOT NULL DEFAULT NOW(),
  tags varchar(255) NOT NULL DEFAULT '',
  CONSTRAINT FOREIGN KEY (author_id) REFERENCES authors(author_id)
) ENGINE=InnoDB;

CREATE INDEX books_title_idx ON books(title, year);

CREATE FUNCTION say_hello(s varchar(255)) RETURNS varchar(255)
  DETERMINISTIC
  RETURN CONCAT('hello ', s);
ENDSQL

$XOBIN mysql://booktest:booktest@localhost/booktest -o $SRC/models

cat << ENDSQL | $XOBIN mysql://booktest:booktest@localhost/booktest -N -M -B -T AuthorBookResult --query-type-comment='AuthorBookResult is the result of a search.' -o $SRC/models
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
popd &> /dev/null
