CREATE TABLE authors (
  author_id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL DEFAULT ''
);

CREATE INDEX authors_name_idx ON authors(name);

CREATE TABLE books (
  book_id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  author_id INTEGER NOT NULL REFERENCES authors(author_id),
  isbn TEXT NOT NULL DEFAULT '' UNIQUE,
  title TEXT NOT NULL DEFAULT '',
  year INTEGER NOT NULL DEFAULT 2000,
  available TIMESTAMP NOT NULL DEFAULT '',
  tags TEXT NOT NULL DEFAULT '{}'
);

CREATE INDEX books_title_idx ON books(title, year);
