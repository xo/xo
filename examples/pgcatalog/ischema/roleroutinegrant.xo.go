// Package ischema contains the types for schema 'information_schema'.
package ischema

import "github.com/hunter-io/xo/examples/pgcatalog/pgtypes"

// Code generated by xo. DO NOT EDIT.

// RoleRoutineGrant represents a row from 'information_schema.role_routine_grants'.
type RoleRoutineGrant struct {
	Grantor         pgtypes.SQLIdentifier `json:"grantor"`          // grantor
	Grantee         pgtypes.SQLIdentifier `json:"grantee"`          // grantee
	SpecificCatalog pgtypes.SQLIdentifier `json:"specific_catalog"` // specific_catalog
	SpecificSchema  pgtypes.SQLIdentifier `json:"specific_schema"`  // specific_schema
	SpecificName    pgtypes.SQLIdentifier `json:"specific_name"`    // specific_name
	RoutineCatalog  pgtypes.SQLIdentifier `json:"routine_catalog"`  // routine_catalog
	RoutineSchema   pgtypes.SQLIdentifier `json:"routine_schema"`   // routine_schema
	RoutineName     pgtypes.SQLIdentifier `json:"routine_name"`     // routine_name
	PrivilegeType   pgtypes.CharacterData `json:"privilege_type"`   // privilege_type
	IsGrantable     pgtypes.YesOrNo       `json:"is_grantable"`     // is_grantable
}
