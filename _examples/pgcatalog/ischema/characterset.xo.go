package ischema

// Code generated by xo. DO NOT EDIT.

import (
	"github.com/mmmcorp/xo/_examples/pgcatalog/pgtypes"
)

// CharacterSet represents a row from 'information_schema.character_sets'.
type CharacterSet struct {
	CharacterSetCatalog   pgtypes.SQLIdentifier `json:"character_set_catalog"`   // character_set_catalog
	CharacterSetSchema    pgtypes.SQLIdentifier `json:"character_set_schema"`    // character_set_schema
	CharacterSetName      pgtypes.SQLIdentifier `json:"character_set_name"`      // character_set_name
	CharacterRepertoire   pgtypes.SQLIdentifier `json:"character_repertoire"`    // character_repertoire
	FormOfUse             pgtypes.SQLIdentifier `json:"form_of_use"`             // form_of_use
	DefaultCollateCatalog pgtypes.SQLIdentifier `json:"default_collate_catalog"` // default_collate_catalog
	DefaultCollateSchema  pgtypes.SQLIdentifier `json:"default_collate_schema"`  // default_collate_schema
	DefaultCollateName    pgtypes.SQLIdentifier `json:"default_collate_name"`    // default_collate_name
}
