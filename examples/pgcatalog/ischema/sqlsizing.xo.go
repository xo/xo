// Package ischema contains the types for schema 'information_schema'.
package ischema

import "github.com/mccolljr/xo/examples/pgcatalog/pgtypes"

// GENERATED BY XO. DO NOT EDIT.

// SQLSizing represents a row from information_schema.sql_sizing.
type SQLSizing struct {
	Tableoid       pgtypes.Oid            // tableoid
	Cmax           pgtypes.Cid            // cmax
	Xmax           pgtypes.Xid            // xmax
	Cmin           pgtypes.Cid            // cmin
	Xmin           pgtypes.Xid            // xmin
	Ctid           pgtypes.Tid            // ctid
	SizingID       pgtypes.CardinalNumber // sizing_id
	SizingName     pgtypes.CharacterData  // sizing_name
	SupportedValue pgtypes.CardinalNumber // supported_value
	Comments       pgtypes.CharacterData  // comments
}
