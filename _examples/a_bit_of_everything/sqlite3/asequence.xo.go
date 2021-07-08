package sqlite3

// Code generated by xo. DO NOT EDIT.

import (
	"context"
)

// ASequence represents a row from 'a_sequence'.
type ASequence struct {
	ASeq int `json:"a_seq"` // a_seq
	// xo fields
	_exists, _deleted bool
}

// Exists returns true when the ASequence exists in the database.
func (as *ASequence) Exists() bool {
	return as._exists
}

// Deleted returns true when the ASequence has been marked for deletion from
// the database.
func (as *ASequence) Deleted() bool {
	return as._deleted
}

// Insert inserts the ASequence to the database.
func (as *ASequence) Insert(ctx context.Context, db DB) error {
	switch {
	case as._exists: // already exists
		return logerror(&ErrInsertFailed{ErrAlreadyExists})
	case as._deleted: // deleted
		return logerror(&ErrInsertFailed{ErrMarkedForDeletion})
	}
	// insert (primary key generated and returned by database)
	const sqlstr = `INSERT INTO a_sequence (` +
		`` +
		`) VALUES (` +
		`` +
		`)`
	// run
	logf(sqlstr)
	res, err := db.ExecContext(ctx, sqlstr)
	if err != nil {
		return err
	}
	// retrieve id
	id, err := res.LastInsertId()
	if err != nil {
		return err
	} // set primary key
	as.ASeq = int(id)
	// set exists
	as._exists = true
	return nil
}

// ------ NOTE: Update statements omitted due to lack of fields other than primary key ------

// Delete deletes the ASequence from the database.
func (as *ASequence) Delete(ctx context.Context, db DB) error {
	switch {
	case !as._exists: // doesn't exist
		return nil
	case as._deleted: // deleted
		return nil
	}
	// delete with single primary key
	const sqlstr = `DELETE FROM a_sequence ` +
		`WHERE a_seq = $1`
	// run
	logf(sqlstr, as.ASeq)
	if _, err := db.ExecContext(ctx, sqlstr, as.ASeq); err != nil {
		return logerror(err)
	}
	// set deleted
	as._deleted = true
	return nil
}

// ASequenceByASeq retrieves a row from 'a_sequence' as a ASequence.
//
// Generated from index 'a_sequence_a_seq_pkey'.
func ASequenceByASeq(ctx context.Context, db DB, aSeq int) (*ASequence, error) {
	// query
	const sqlstr = `SELECT ` +
		`a_seq ` +
		`FROM a_sequence ` +
		`WHERE a_seq = $1`
	// run
	logf(sqlstr, aSeq)
	as := ASequence{
		_exists: true,
	}
	if err := db.QueryRowContext(ctx, sqlstr, aSeq).Scan(&as.ASeq); err != nil {
		return nil, logerror(err)
	}
	return &as, nil
}
