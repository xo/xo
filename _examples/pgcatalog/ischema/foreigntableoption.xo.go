package ischema

// Code generated by xo. DO NOT EDIT.

import (
	"github.com/mmmcorp/xo/_examples/pgcatalog/pgtypes"
)

// ForeignTableOption represents a row from 'information_schema.foreign_table_options'.
type ForeignTableOption struct {
	ForeignTableCatalog pgtypes.SQLIdentifier `json:"foreign_table_catalog"` // foreign_table_catalog
	ForeignTableSchema  pgtypes.SQLIdentifier `json:"foreign_table_schema"`  // foreign_table_schema
	ForeignTableName    pgtypes.SQLIdentifier `json:"foreign_table_name"`    // foreign_table_name
	OptionName          pgtypes.SQLIdentifier `json:"option_name"`           // option_name
	OptionValue         pgtypes.CharacterData `json:"option_value"`          // option_value
}
