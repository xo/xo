// Package ischema contains the types for schema 'information_schema'.
package ischema

import "github.com/bannzai/xox/examples/pgcatalog/pgtypes"

// Code generated by xo. DO NOT EDIT.

// Collation represents a row from 'information_schema.collations'.
type Collation struct {
	CollationCatalog pgtypes.SQLIdentifier `json:"collation_catalog"` // collation_catalog
	CollationSchema  pgtypes.SQLIdentifier `json:"collation_schema"`  // collation_schema
	CollationName    pgtypes.SQLIdentifier `json:"collation_name"`    // collation_name
	PadAttribute     pgtypes.CharacterData `json:"pad_attribute"`     // pad_attribute
}
