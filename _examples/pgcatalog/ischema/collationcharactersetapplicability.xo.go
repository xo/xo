package ischema

// Code generated by xo. DO NOT EDIT.

// CollationCharacterSetApplicability represents a row from 'information_schema.collation_character_set_applicability'.
type CollationCharacterSetApplicability struct {
	CollationCatalog    SQLIdentifier `json:"collation_catalog"`     // collation_catalog
	CollationSchema     SQLIdentifier `json:"collation_schema"`      // collation_schema
	CollationName       SQLIdentifier `json:"collation_name"`        // collation_name
	CharacterSetCatalog SQLIdentifier `json:"character_set_catalog"` // character_set_catalog
	CharacterSetSchema  SQLIdentifier `json:"character_set_schema"`  // character_set_schema
	CharacterSetName    SQLIdentifier `json:"character_set_name"`    // character_set_name
}
