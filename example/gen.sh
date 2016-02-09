#!/bin/bash

SRC=$(realpath $(cd -P "$( dirname "${BASH_SOURCE[0]}" )" && pwd ))

XOBIN=$(which xo)
if [ -e $SRC/../xo ]; then
  XOBIN=$SRC/../xo
fi

set -x

psql -U postgres -c "create user booktest password 'booktest';"
psql -U postgres -c 'drop database booktest;'
psql -U postgres -c 'create database booktest owner booktest;'

psql -U booktest << 'ENDSQL'
CREATE TABLE authors (
  author_id SERIAL PRIMARY KEY,
  isbn text NOT NULL DEFAULT '' UNIQUE,
  name text NOT NULL DEFAULT '',
  subject text NOT NULL DEFAULT ''
);

CREATE INDEX authors_name_idx ON authors(name);

CREATE TYPE book_type AS ENUM (
  'FICTION',
  'NONFICTION'
);

CREATE TABLE books (
    book_id SERIAL PRIMARY KEY,
    author_id integer NOT NULL REFERENCES authors(author_id),
    title text NOT NULL DEFAULT '',
    booktype book_type NOT NULL DEFAULT 'FICTION',
    year integer NOT NULL DEFAULT 2000
);

CREATE INDEX books_title_idx ON books(title, year);

CREATE FUNCTION say_hello(text) RETURNS text AS $$
BEGIN
    RETURN CONCAT('hello ' || $1);
END;
$$ LANGUAGE plpgsql;

ENDSQL

$XOBIN pgsql://booktest:booktest@localhost/booktest -o $SRC/models

pushd $SRC &> /dev/null
go build
popd &> /dev/null
