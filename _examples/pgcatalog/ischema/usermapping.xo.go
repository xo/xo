package ischema

// Code generated by xo. DO NOT EDIT.

import (
	"github.com/mmmcorp/xo/_examples/pgcatalog/pgtypes"
)

// UserMapping represents a row from 'information_schema.user_mappings'.
type UserMapping struct {
	AuthorizationIdentifier pgtypes.SQLIdentifier `json:"authorization_identifier"` // authorization_identifier
	ForeignServerCatalog    pgtypes.SQLIdentifier `json:"foreign_server_catalog"`   // foreign_server_catalog
	ForeignServerName       pgtypes.SQLIdentifier `json:"foreign_server_name"`      // foreign_server_name
}
