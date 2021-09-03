package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	models "github.com/xo/xo/_examples/django/sqlite3"
)

func runSqlite3(ctx context.Context, db *sql.DB) error {
	now := time.Now()
	a := &models.Author{
		Name: "author 1",
	}
	if err := a.Insert(ctx, db); err != nil {
		return err
	}
	t := &models.Tag{
		Tag: "tag 1",
	}
	if err := t.Insert(ctx, db); err != nil {
		return err
	}
	b := &models.Book{
		ISBN:              "1",
		BookType:          1,
		Title:             "book 1",
		Year:              now.Year(),
		Available:         models.NewTime(now),
		BooksAuthorIDFkey: int64(a.AuthorID),
	}
	if err := b.Insert(ctx, db); err != nil {
		return err
	}
	bt := &models.BooksTag{
		BookID: int64(b.BookID),
		TagID:  int64(t.TagID),
	}
	if err := bt.Insert(ctx, db); err != nil {
		return err
	}
	books, err := models.BooksByBooksAuthorIDFkey(ctx, db, int64(a.AuthorID))
	if err != nil {
		return err
	}
	for _, book := range books {
		fmt.Fprintf(os.Stdout, "book %d: %q (%s) %d\n", book.BookID, book.Title, book.ISBN, book.Year)
	}
	return nil
}
