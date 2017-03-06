DROP TABLE IF EXISTS books;
DROP TABLE IF EXISTS authors;

CREATE TABLE authors (
  author_id integer NOT NULL IDENTITY(1,1) PRIMARY KEY,
  name varchar(255) NOT NULL DEFAULT ''
);

CREATE INDEX authors_name_idx ON authors(name);

CREATE TABLE books (
  book_id integer NOT NULL IDENTITY(1,1) PRIMARY KEY,
  author_id integer NOT NULL FOREIGN KEY REFERENCES authors(author_id),
  isbn varchar(255) NOT NULL DEFAULT '' UNIQUE,
  title varchar(255) NOT NULL DEFAULT '',
  year integer NOT NULL DEFAULT 2000,
  available datetime2 NOT NULL DEFAULT CURRENT_TIMESTAMP,
  tags varchar(255) NOT NULL DEFAULT ''
);

CREATE INDEX books_title_idx ON books(title, year);
