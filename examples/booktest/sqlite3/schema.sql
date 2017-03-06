PRAGMA foreign_keys = 1;

DROP TABLE IF EXISTS books;
DROP TABLE IF EXISTS authors;

CREATE TABLE authors (
  author_id integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  name text NOT NULL DEFAULT ''
);

CREATE INDEX authors_name_idx ON authors(name);

CREATE TABLE books (
  book_id integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  author_id integer NOT NULL REFERENCES authors(author_id),
  isbn text NOT NULL DEFAULT '' UNIQUE,
  title text NOT NULL DEFAULT '',
  year integer NOT NULL DEFAULT 2000,
  available timestamp with time zone NOT NULL DEFAULT '',
  tags text NOT NULL DEFAULT '{}'
);

CREATE INDEX books_title_idx ON books(title, year);
