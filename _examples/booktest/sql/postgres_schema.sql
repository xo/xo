DROP TABLE IF EXISTS books CASCADE;
DROP TYPE IF EXISTS book_type CASCADE;
DROP TABLE IF EXISTS authors CASCADE;
DROP FUNCTION IF EXISTS say_hello(text) CASCADE;

CREATE TABLE authors (
  author_id SERIAL PRIMARY KEY,
  name TEXT NOT NULL DEFAULT ''
);

CREATE INDEX authors_name_idx ON authors(name);

CREATE TYPE book_type AS ENUM (
  'FICTION',
  'NONFICTION'
);

CREATE TABLE books (
  book_id SERIAL PRIMARY KEY,
  author_id INTEGER NOT NULL REFERENCES authors(author_id),
  isbn TEXT NOT NULL DEFAULT '' UNIQUE,
  booktype book_type NOT NULL DEFAULT 'FICTION',
  title TEXT NOT NULL DEFAULT '',
  year INTEGER NOT NULL DEFAULT 2000,
  available TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT 'NOW()',
  tags VARCHAR[] NOT NULL DEFAULT '{}'
);

CREATE INDEX books_title_idx ON books(title, year);

CREATE FUNCTION say_hello(text) RETURNS TEXT AS $$
BEGIN
  RETURN CONCAT('hello ', $1);
END;
$$ LANGUAGE plpgsql;

CREATE INDEX books_title_lower_idx ON books(title);
