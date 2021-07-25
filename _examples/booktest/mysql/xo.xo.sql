-- SQL Schema template for the booktest schema.
-- Generated on Sun Jul 25 07:07:57 WIB 2021 by xo.

CREATE TABLE authors (
    author_id INT(11) AUTO_INCREMENT,
    name VARCHAR(255) DEFAULT '' NOT NULL,
    PRIMARY KEY (author_id)
) ENGINE=InnoDB;

CREATE INDEX authors_name_idx ON authors (name);

CREATE TABLE books (
    book_id INT(11) AUTO_INCREMENT,
    author_id INT(11) NOT NULL REFERENCES authors (author_id),
    isbn VARCHAR(255) DEFAULT '' NOT NULL,
    book_type ENUM('FICTION', 'NONFICTION') DEFAULT 'FICTION' NOT NULL,
    title VARCHAR(255) DEFAULT '' NOT NULL,
    year INT(11) DEFAULT 2000 NOT NULL,
    available DATETIME DEFAULT current_timestamp() NOT NULL,
    description TEXT DEFAULT '' NOT NULL,
    tags TEXT DEFAULT '' NOT NULL,
    UNIQUE (isbn),
    PRIMARY KEY (book_id)
) ENGINE=InnoDB;

CREATE INDEX author_id ON books (author_id);
CREATE INDEX books_title_idx ON books (title, year);

CREATE FUNCTION say_hello(name VARCHAR(255)) RETURNS VARCHAR(255)
BEGIN
  RETURN CONCAT('hello ', name)\;
END;

