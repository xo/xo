CREATE TABLE authors (
  author_id SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE INDEX authors_name_idx ON authors(name);

CREATE TYPE book_type AS ENUM (
  'FICTION',
  'NONFICTION'
);

CREATE TABLE books (
  book_id SERIAL PRIMARY KEY,
  author_id INTEGER NOT NULL CONSTRAINT books_author_id_fkey REFERENCES authors(author_id),
  isbn VARCHAR(255) NOT NULL DEFAULT '' CONSTRAINT books_isbn_key UNIQUE,
  book_type book_type DEFAULT 'FICTION' NOT NULL,
  title VARCHAR(255) DEFAULT '' NOT NULL,
  year INTEGER DEFAULT 2000 NOT NULL,
  available TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
  description TEXT DEFAULT '' NOT NULL,
  tags VARCHAR[] DEFAULT '{}' NOT NULL
);

CREATE INDEX books_title_idx ON books(title, year);

CREATE FUNCTION say_hello(name VARCHAR(255)) RETURNS VARCHAR(255) AS $$
BEGIN
  RETURN 'hello ' || name;
END;
$$ LANGUAGE plpgsql;
