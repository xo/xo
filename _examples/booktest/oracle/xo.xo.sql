-- SQL Schema template for the booktest schema.
-- Generated on Sun Jul 25 07:08:01 WIB 2021 by xo.

CREATE TABLE authors (
    author_id NUMBER GENERATED ALWAYS AS IDENTITY,
    name NVARCHAR2 NOT NULL,
    CONSTRAINT authors_pkey UNIQUE (author_id)
);

CREATE INDEX authors_name_idx ON authors (name);

CREATE TABLE books (
    book_id NUMBER GENERATED ALWAYS AS IDENTITY,
    author_id NUMBER NOT NULL CONSTRAINT books_author_id_fkey REFERENCES authors (author_id),
    isbn NVARCHAR2 NOT NULL,
    title NVARCHAR2 NOT NULL,
    year NUMBER NOT NULL,
    available TIMESTAMP WITH TIME ZONE(6) NOT NULL,
    description NCLOB,
    tags NCLOB NOT NULL,
    CONSTRAINT books_isbn_key UNIQUE (isbn),
    CONSTRAINT books_pkey UNIQUE (book_id)
);

CREATE INDEX books_title_idx ON books (title, year);

CREATE FUNCTION say_hello(name IN NVARCHAR2) RETURN NVARCHAR2 AS
BEGIN
  RETURN 'hello ' || name\;
END\;;

