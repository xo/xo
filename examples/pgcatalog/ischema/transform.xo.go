// Package ischema contains the types for schema 'information_schema'.
package ischema

import "github.com/codercom/xo/examples/pgcatalog/pgtypes"

// Code generated by xo. DO NOT EDIT.

// Transform represents a row from 'information_schema.transforms'.
type Transform struct {
	UdtCatalog      pgtypes.SQLIdentifier `json:"udt_catalog"`      // udt_catalog
	UdtSchema       pgtypes.SQLIdentifier `json:"udt_schema"`       // udt_schema
	UdtName         pgtypes.SQLIdentifier `json:"udt_name"`         // udt_name
	SpecificCatalog pgtypes.SQLIdentifier `json:"specific_catalog"` // specific_catalog
	SpecificSchema  pgtypes.SQLIdentifier `json:"specific_schema"`  // specific_schema
	SpecificName    pgtypes.SQLIdentifier `json:"specific_name"`    // specific_name
	GroupName       pgtypes.SQLIdentifier `json:"group_name"`       // group_name
	TransformType   pgtypes.CharacterData `json:"transform_type"`   // transform_type
}
