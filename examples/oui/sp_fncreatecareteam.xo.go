// Package oui contains the types for schema 'oui'.
package oui

import "github.com/google/uuid"

// Code generated by xo. DO NOT EDIT.

// FnCreateCareTeam calls the stored procedure 'oui.fn_create_care_team(uuid) uuid' on db.
func FnCreateCareTeam(db XODB, i_patient_id uuid.UUID) (uuid.UUID, error) {
	var err error

	// sql query
	const sqlstr = `SELECT oui.fn_create_care_team($1)`

	// run query
	var ret uuid.UUID
	XOLog(sqlstr, i_patient_id)
	err = db.QueryRow(sqlstr, i_patient_id).Scan(&ret)
	if err != nil {
		return uuid.New(), err
	}

	return ret, nil
}
