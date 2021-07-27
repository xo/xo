CREATE TABLE authors (
  author_id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  name VARCHAR(255) DEFAULT '' NOT NULL
);

CREATE INDEX authors_name_idx ON authors (name);

CREATE TABLE books (
  book_id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  author_id INTEGER NOT NULL REFERENCES authors (author_id),
  isbn VARCHAR(255) DEFAULT '' NOT NULL UNIQUE,
  title VARCHAR(255) DEFAULT '' NOT NULL,
  year INTEGER DEFAULT 2000 NOT NULL,
  available TIMESTAMP DEFAULT (STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'NOW')) NOT NULL,
  description TEXT DEFAULT '' NOT NULL,
  tags TEXT DEFAULT '{}' NOT NULL
);

CREATE INDEX books_title_idx ON books (title, year);
