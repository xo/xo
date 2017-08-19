#!/bin/bash

# SETUP THE CLUSTER AND INITIALIZE DB:
cockroach start --insecure --store=booktest --host=localhost --background

cockroach user set booktest --insecure

cockroach sql --insecure -e 'CREATE DATABASE booktest'

cockroach sql --insecure -e 'GRANT ALL ON DATABASE booktest TO booktest'

#SEED THE DB WITH TABLES:
cockroach sql --url="postgresql://booktest@localhost:26257/booktest?sslmode=disable" << 'ENDSQL'

DROP TABLE IF EXISTS books CASCADE;
DROP TABLE IF EXISTS authors CASCADE;

CREATE TABLE authors (
  author_id SERIAL PRIMARY KEY,
  name text NOT NULL DEFAULT ''
);

CREATE INDEX authors_name_idx ON authors(name);

CREATE TABLE books (
  book_id SERIAL PRIMARY KEY,
  author_id integer NOT NULL REFERENCES authors(author_id),
  isbn text NOT NULL DEFAULT '' UNIQUE,
  booktype varchar(25) NOT NULL DEFAULT 'FICTION',
  title text NOT NULL DEFAULT '',
  year integer NOT NULL DEFAULT 2000,
  available timestamp with time zone NOT NULL,
  tags varchar(25) NOT NULL DEFAULT '{}'
);

CREATE INDEX books_title_year_idx ON books(title, year);
CREATE INDEX books_title_lower_idx ON books(title);
ENDSQL

#RUN XO TO GENERATE THE MODELS:

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
rm -f $SRC/postgres

$XOBIN CR:booktest:booktest@localhost:26257/booktest?sslmode=disable $EXTRA -o $DEST -s booktest

pushd $SRC &> /dev/null

go build
./cockroachdb $EXTRA

cockroach sql --insecure -e 'set database = booktest; select * from books;'
