-- SQL Schema template for the booktest schema.
-- Generated on Sun Jul 25 07:08:06 WIB 2021 by xo.

CREATE TABLE authors (
    author_id INT(10) IDENTITY(1, 1),
    name VARCHAR(255) DEFAULT ('') NOT NULL,
    CONSTRAINT authors_pkey PRIMARY KEY (author_id)
);

CREATE INDEX authors_name_idx ON authors (name);

CREATE TABLE books (
    book_id INT(10) IDENTITY(1, 1),
    author_id INT(10) NOT NULL CONSTRAINT books_author_id_fkey REFERENCES authors (author_id),
    isbn NVARCHAR(255) DEFAULT ('') NOT NULL,
    title NVARCHAR(255) DEFAULT ('') NOT NULL,
    year INT(10) DEFAULT ((2000)) NOT NULL,
    available DATETIME2(27, 7) DEFAULT (getdate()) NOT NULL,
    description NTEXT DEFAULT ('') NOT NULL,
    tags TEXT DEFAULT ('') NOT NULL,
    CONSTRAINT books_isbn_key UNIQUE (isbn),
    CONSTRAINT books_pkey PRIMARY KEY (book_id)
);

CREATE INDEX books_title_idx ON books (title, year);

CREATE PROCEDURE say_hello @name NVARCHAR(255), @result NVARCHAR(255) OUTPUT AS
BEGIN
  SELECT @result = CONCAT('hello ', @name)\;
END;

