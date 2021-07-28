package mysql

// Code generated by xo. DO NOT EDIT.

import (
	"context"
)

// AuthorBookResult is the result of a search.
type AuthorBookResult struct {
	AuthorID   int    `json:"author_id"`   // author_id
	AuthorName string `json:"author_name"` // author_name
	BookID     int    `json:"book_id"`     // book_id
	BookISBN   string `json:"book_isbn"`   // book_isbn
	BookTitle  string `json:"book_title"`  // book_title
	BookTags   string `json:"book_tags"`   // book_tags
}

// AuthorBookResultsByTag runs a custom query, returning results as AuthorBookResult.
func AuthorBookResultsByTag(ctx context.Context, db DB, tag string) ([]*AuthorBookResult, error) {
	// query
	const sqlstr = `SELECT ` +
		`a.author_id AS author_id, ` +
		`a.name AS author_name, ` +
		`b.book_id AS book_id, ` +
		`b.isbn AS book_isbn, ` +
		`b.title AS book_title, ` +
		`b.tags AS book_tags ` +
		`FROM books b ` +
		`JOIN authors a ON a.author_id = b.author_id ` +
		`WHERE b.tags LIKE CONCAT('%', ?, '%')`
	// run
	logf(sqlstr, tag)
	rows, err := db.QueryContext(ctx, sqlstr, tag)
	if err != nil {
		return nil, logerror(err)
	}
	defer rows.Close()
	// load results
	var res []*AuthorBookResult
	for rows.Next() {
		var abr AuthorBookResult
		// scan
		if err := rows.Scan(&abr.AuthorID, &abr.AuthorName, &abr.BookID, &abr.BookISBN, &abr.BookTitle, &abr.BookTags); err != nil {
			return nil, logerror(err)
		}
		res = append(res, &abr)
	}
	if err := rows.Err(); err != nil {
		return nil, logerror(err)
	}
	return res, nil
}
