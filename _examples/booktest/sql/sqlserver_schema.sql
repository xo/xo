CREATE TABLE authors (
  author_id INTEGER NOT NULL IDENTITY(1,1) CONSTRAINT authors_pkey PRIMARY KEY,
  name VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE INDEX authors_name_idx ON authors(name);

CREATE TABLE books (
  book_id INTEGER NOT NULL IDENTITY CONSTRAINT books_pkey PRIMARY KEY,
  author_id INTEGER NOT NULL CONSTRAINT books_author_id_fkey FOREIGN KEY REFERENCES authors(author_id),
  isbn NVARCHAR(255) NOT NULL DEFAULT '' CONSTRAINT books_isbn_key UNIQUE,
  title NVARCHAR(255) NOT NULL DEFAULT '',
  year INTEGER NOT NULL DEFAULT 2000,
  available DATETIME2 NOT NULL DEFAULT CURRENT_TIMESTAMP,
  description NTEXT DEFAULT '' NOT NULL,
  tags TEXT NOT NULL DEFAULT ''
);

CREATE INDEX books_title_idx ON books(title, year);

CREATE PROCEDURE say_hello @name NVARCHAR(255), @result NVARCHAR(255) OUTPUT AS
BEGIN
  SELECT @result = CONCAT('hello ', @name)\;
END;
