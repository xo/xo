// Package TESTGEN contains the types for schema 'c9'.
package TESTGEN

// GENERATED BY XO. DO NOT EDIT.

import (
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
)

// Post represents a row from 'c9.posts'.
type Post struct {
	ID      int            // id
	Tags    sql.NullString // tags
	Content string         // content
	Title   string         // title
	Created pq.NullTime    // created

	// xo fields
	_exists, _deleted bool
}

// Exists determines if the Post exists in the database.
func (p *Post) Exists() bool {
	return p._exists
}

// Deleted provides information if the Post has been deleted from the database.
func (p *Post) Deleted() bool {
	return p._deleted
}

// Insert inserts the Post to the database.
func (p *Post) Insert(db XODB) error {
	var err error

	// if already exist, bail
	if p._exists {
		return errors.New("insert failed: already exists")
	}

	// sql query
	const sqlstr = `INSERT INTO c9.posts (` +
		`tags, content, title, created` +
		`) VALUES (` +
		`?, ?, ?, ?` +
		`)`

	// run query
	XOLog(sqlstr, p.Tags, p.Content, p.Title, pq.NullTime{time.Now(), true})
	res, err := db.Exec(sqlstr, p.Tags, p.Content, p.Title, pq.NullTime{time.Now(), true})
	if err != nil {
		return err
	}

	// retrieve id
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	// set primary key and existence
	p.ID = int(id)
	p._exists = true

	return nil
}

// Update updates the Post in the database.
func (p *Post) Update(db XODB) error {
	var err error

	// if doesn't exist, bail
	if !p._exists {
		return errors.New("update failed: does not exist")
	}

	// if deleted, bail
	if p._deleted {
		return errors.New("update failed: marked for deletion")
	}

	// sql query
	const sqlstr = `UPDATE c9.posts SET ` +
		`tags = ?, content = ?, title = ?, created = ?` +
		` WHERE id = ?`

	// run query
	XOLog(sqlstr, p.Tags, p.Content, p.Title, p.Created, p.ID)
	_, err = db.Exec(sqlstr, p.Tags, p.Content, p.Title, p.Created, p.ID)
	return err
}

// Save saves the Post to the database.
func (p *Post) Save(db XODB) error {
	if p.Exists() {
		return p.Update(db)
	}

	return p.Insert(db)
}

// Delete deletes the Post from the database.
func (p *Post) Delete(db XODB) error {
	var err error

	// if doesn't exist, bail
	if !p._exists {
		return nil
	}

	// if deleted, bail
	if p._deleted {
		return nil
	}

	// sql query
	const sqlstr = `DELETE FROM c9.posts WHERE id = ?`

	// run query
	XOLog(sqlstr, p.ID)
	_, err = db.Exec(sqlstr, p.ID)
	if err != nil {
		return err
	}

	// set deleted
	p._deleted = true

	return nil
}

// PostByID retrieves a row from 'c9.posts' as a Post.
//
// Generated from index 'posts_id_pkey'.
func PostByID(db XODB, id int) (*Post, error) {
	var err error

	// sql query
	const sqlstr = `SELECT ` +
		`id, tags, content, title, created ` +
		`FROM c9.posts ` +
		`WHERE id = ?`

	// run query
	XOLog(sqlstr, id)
	p := Post{
		_exists: true,
	}

	err = db.QueryRow(sqlstr, id).Scan(&p.ID, &p.Tags, &p.Content, &p.Title, &p.Created)
	if err != nil {
		return nil, err
	}

	return &p, nil
}
