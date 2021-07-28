package postgres

// Code generated by xo. DO NOT EDIT.

import (
	"context"
	"time"
)

// Book represents a row from 'public.books'.
type Book struct {
	BookID      int         `json:"book_id"`     // book_id
	AuthorID    int         `json:"author_id"`   // author_id
	ISBN        string      `json:"isbn"`        // isbn
	BookType    BookType    `json:"book_type"`   // book_type
	Title       string      `json:"title"`       // title
	Year        int         `json:"year"`        // year
	Available   time.Time   `json:"available"`   // available
	Description string      `json:"description"` // description
	Tags        StringSlice `json:"tags"`        // tags
	// xo fields
	_exists, _deleted bool
}

// Exists returns true when the Book exists in the database.
func (b *Book) Exists() bool {
	return b._exists
}

// Deleted returns true when the Book has been marked for deletion from
// the database.
func (b *Book) Deleted() bool {
	return b._deleted
}

// Insert inserts the Book to the database.
func (b *Book) Insert(ctx context.Context, db DB) error {
	switch {
	case b._exists: // already exists
		return logerror(&ErrInsertFailed{ErrAlreadyExists})
	case b._deleted: // deleted
		return logerror(&ErrInsertFailed{ErrMarkedForDeletion})
	}
	// insert (primary key generated and returned by database)
	const sqlstr = `INSERT INTO public.books (` +
		`author_id, isbn, book_type, title, year, available, description, tags` +
		`) VALUES (` +
		`$1, $2, $3, $4, $5, $6, $7, $8` +
		`) RETURNING book_id`
	// run
	logf(sqlstr, b.AuthorID, b.ISBN, b.BookType, b.Title, b.Year, b.Available, b.Description, b.Tags)
	if err := db.QueryRowContext(ctx, sqlstr, b.AuthorID, b.ISBN, b.BookType, b.Title, b.Year, b.Available, b.Description, b.Tags).Scan(&b.BookID); err != nil {
		return logerror(err)
	}
	// set exists
	b._exists = true
	return nil
}

// Update updates a Book in the database.
func (b *Book) Update(ctx context.Context, db DB) error {
	switch {
	case !b._exists: // doesn't exist
		return logerror(&ErrUpdateFailed{ErrDoesNotExist})
	case b._deleted: // deleted
		return logerror(&ErrUpdateFailed{ErrMarkedForDeletion})
	}
	// update with composite primary key
	const sqlstr = `UPDATE public.books SET ` +
		`author_id = $1, isbn = $2, book_type = $3, title = $4, year = $5, available = $6, description = $7, tags = $8 ` +
		`WHERE book_id = $9`
	// run
	logf(sqlstr, b.AuthorID, b.ISBN, b.BookType, b.Title, b.Year, b.Available, b.Description, b.Tags, b.BookID)
	if _, err := db.ExecContext(ctx, sqlstr, b.AuthorID, b.ISBN, b.BookType, b.Title, b.Year, b.Available, b.Description, b.Tags, b.BookID); err != nil {
		return logerror(err)
	}
	return nil
}

// Save saves the Book to the database.
func (b *Book) Save(ctx context.Context, db DB) error {
	if b.Exists() {
		return b.Update(ctx, db)
	}
	return b.Insert(ctx, db)
}

// Upsert performs an upsert for Book.
func (b *Book) Upsert(ctx context.Context, db DB) error {
	switch {
	case b._deleted: // deleted
		return logerror(&ErrUpsertFailed{ErrMarkedForDeletion})
	}
	// upsert
	const sqlstr = `INSERT INTO public.books (` +
		`book_id, author_id, isbn, book_type, title, year, available, description, tags` +
		`) VALUES (` +
		`$1, $2, $3, $4, $5, $6, $7, $8, $9` +
		`)` +
		` ON CONFLICT (book_id) DO ` +
		`UPDATE SET ` +
		`author_id = EXCLUDED.author_id, isbn = EXCLUDED.isbn, book_type = EXCLUDED.book_type, title = EXCLUDED.title, year = EXCLUDED.year, available = EXCLUDED.available, description = EXCLUDED.description, tags = EXCLUDED.tags `
	// run
	logf(sqlstr, b.BookID, b.AuthorID, b.ISBN, b.BookType, b.Title, b.Year, b.Available, b.Description, b.Tags)
	if _, err := db.ExecContext(ctx, sqlstr, b.BookID, b.AuthorID, b.ISBN, b.BookType, b.Title, b.Year, b.Available, b.Description, b.Tags); err != nil {
		return err
	}
	// set exists
	b._exists = true
	return nil
}

// Delete deletes the Book from the database.
func (b *Book) Delete(ctx context.Context, db DB) error {
	switch {
	case !b._exists: // doesn't exist
		return nil
	case b._deleted: // deleted
		return nil
	}
	// delete with single primary key
	const sqlstr = `DELETE FROM public.books ` +
		`WHERE book_id = $1`
	// run
	logf(sqlstr, b.BookID)
	if _, err := db.ExecContext(ctx, sqlstr, b.BookID); err != nil {
		return logerror(err)
	}
	// set deleted
	b._deleted = true
	return nil
}

// BookByISBN retrieves a row from 'public.books' as a Book.
//
// Generated from index 'books_isbn_key'.
func BookByISBN(ctx context.Context, db DB, isbn string) (*Book, error) {
	// query
	const sqlstr = `SELECT ` +
		`book_id, author_id, isbn, book_type, title, year, available, description, tags ` +
		`FROM public.books ` +
		`WHERE isbn = $1`
	// run
	logf(sqlstr, isbn)
	b := Book{
		_exists: true,
	}
	if err := db.QueryRowContext(ctx, sqlstr, isbn).Scan(&b.BookID, &b.AuthorID, &b.ISBN, &b.BookType, &b.Title, &b.Year, &b.Available, &b.Description, &b.Tags); err != nil {
		return nil, logerror(err)
	}
	return &b, nil
}

// BookByBookID retrieves a row from 'public.books' as a Book.
//
// Generated from index 'books_pkey'.
func BookByBookID(ctx context.Context, db DB, bookID int) (*Book, error) {
	// query
	const sqlstr = `SELECT ` +
		`book_id, author_id, isbn, book_type, title, year, available, description, tags ` +
		`FROM public.books ` +
		`WHERE book_id = $1`
	// run
	logf(sqlstr, bookID)
	b := Book{
		_exists: true,
	}
	if err := db.QueryRowContext(ctx, sqlstr, bookID).Scan(&b.BookID, &b.AuthorID, &b.ISBN, &b.BookType, &b.Title, &b.Year, &b.Available, &b.Description, &b.Tags); err != nil {
		return nil, logerror(err)
	}
	return &b, nil
}

// BooksByTitleYear retrieves a row from 'public.books' as a Book.
//
// Generated from index 'books_title_idx'.
func BooksByTitleYear(ctx context.Context, db DB, title string, year int) ([]*Book, error) {
	// query
	const sqlstr = `SELECT ` +
		`book_id, author_id, isbn, book_type, title, year, available, description, tags ` +
		`FROM public.books ` +
		`WHERE title = $1 AND year = $2`
	// run
	logf(sqlstr, title, year)
	rows, err := db.QueryContext(ctx, sqlstr, title, year)
	if err != nil {
		return nil, logerror(err)
	}
	defer rows.Close()
	// process
	var res []*Book
	for rows.Next() {
		b := Book{
			_exists: true,
		}
		// scan
		if err := rows.Scan(&b.BookID, &b.AuthorID, &b.ISBN, &b.BookType, &b.Title, &b.Year, &b.Available, &b.Description, &b.Tags); err != nil {
			return nil, logerror(err)
		}
		res = append(res, &b)
	}
	if err := rows.Err(); err != nil {
		return nil, logerror(err)
	}
	return res, nil
}

// Author returns the Author associated with the Book's (AuthorID).
//
// Generated from foreign key 'books_author_id_fkey'.
func (b *Book) Author(ctx context.Context, db DB) (*Author, error) {
	return AuthorByAuthorID(ctx, db, b.AuthorID)
}
