// Package ischema contains the types for schema 'information_schema'.
package ischema

import "github.com/mccolljr/xo/examples/pgcatalog/pgtypes"

// GENERATED BY XO. DO NOT EDIT.

// SQLSizingProfile represents a row from information_schema.sql_sizing_profiles.
type SQLSizingProfile struct {
	Tableoid      pgtypes.Oid            // tableoid
	Cmax          pgtypes.Cid            // cmax
	Xmax          pgtypes.Xid            // xmax
	Cmin          pgtypes.Cid            // cmin
	Xmin          pgtypes.Xid            // xmin
	Ctid          pgtypes.Tid            // ctid
	SizingID      pgtypes.CardinalNumber // sizing_id
	SizingName    pgtypes.CharacterData  // sizing_name
	ProfileID     pgtypes.CharacterData  // profile_id
	RequiredValue pgtypes.CardinalNumber // required_value
	Comments      pgtypes.CharacterData  // comments
}
