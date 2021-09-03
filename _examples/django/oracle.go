package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	models "github.com/xo/xo/_examples/django/oracle"
)

func runOracle(ctx context.Context, db *sql.DB) error {
	now := time.Now()
	a := &models.Author{
		Name: sql.NullString{"author 1", true},
	}
	if err := a.Insert(ctx, db); err != nil {
		return err
	}
	t := &models.Tag{
		Tag: sql.NullString{"tag 1", true},
	}
	if err := t.Insert(ctx, db); err != nil {
		return err
	}
	b := &models.Book{
		ISBN:              sql.NullString{"1", true},
		BookType:          1,
		Title:             sql.NullString{"book 1", true},
		Year:              int64(now.Year()),
		Available:         now,
		BooksAuthorIDFkey: a.AuthorID,
	}
	if err := b.Insert(ctx, db); err != nil {
		return err
	}
	bt := &models.BooksTag{
		BookID: b.BookID,
		TagID:  t.TagID,
	}
	if err := bt.Insert(ctx, db); err != nil {
		return err
	}
	books, err := models.BooksByBooksAuthorIDFkey(ctx, db, a.AuthorID)
	if err != nil {
		return err
	}
	for _, book := range books {
		fmt.Fprintf(os.Stdout, "book %d: %q (%s) %d\n", book.BookID, book.Title.String, book.ISBN.String, book.Year)
	}
	return nil
}
