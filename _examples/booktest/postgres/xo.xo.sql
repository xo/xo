-- SQL Schema template for the public schema.
-- Generated on Sun Jul 25 07:08:04 WIB 2021 by xo.

CREATE TYPE book_type AS ENUM (
    'FICTION',
    'NONFICTION'
);

CREATE TABLE authors (
    author_id SERIAL,
    name VARCHAR(255) DEFAULT '' NOT NULL,
    PRIMARY KEY (author_id)
);

CREATE INDEX authors_name_idx ON authors (name);

CREATE TABLE books (
    book_id SERIAL,
    author_id INTEGER NOT NULL REFERENCES authors (author_id),
    isbn VARCHAR(255) DEFAULT '' NOT NULL,
    book_type BOOK_TYPE DEFAULT 'FICTION' NOT NULL,
    title VARCHAR(255) DEFAULT '' NOT NULL,
    year INTEGER DEFAULT 2000 NOT NULL,
    available TIMESTAMPTZ DEFAULT now() NOT NULL,
    description TEXT DEFAULT '' NOT NULL,
    tags VARCHAR[] DEFAULT '{}' NOT NULL,
    UNIQUE (isbn),
    PRIMARY KEY (book_id)
);

CREATE INDEX books_title_idx ON books (title, year);

CREATE FUNCTION say_hello(name VARCHAR) RETURNS VARCHAR AS $$
BEGIN
  RETURN 'hello ' || name;
END;
$$ LANGUAGE plpgsql;

