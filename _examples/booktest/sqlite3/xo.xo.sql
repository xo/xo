-- SQL Schema template for the booktest.db schema.
-- Generated on Sun Jul 25 07:08:05 WIB 2021 by xo.

CREATE TABLE authors (
    author_id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255) DEFAULT '' NOT NULL
);

CREATE INDEX authors_name_idx ON authors (name);

CREATE TABLE books (
    book_id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    author_id INTEGER NOT NULL REFERENCES authors (author_id),
    isbn VARCHAR(255) DEFAULT '' NOT NULL,
    title VARCHAR(255) DEFAULT '' NOT NULL,
    year INTEGER DEFAULT 2000 NOT NULL,
    available TIMESTAMP DEFAULT STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'NOW') NOT NULL,
    description TEXT DEFAULT '' NOT NULL,
    tags TEXT DEFAULT '{}' NOT NULL,
    UNIQUE (isbn)
);

CREATE INDEX books_title_idx ON books (title, year);

