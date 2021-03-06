package ischema

// Code generated by xo. DO NOT EDIT.

// SQLPart represents a row from 'information_schema.sql_parts'.
type SQLPart struct {
	Tableoid     Oid           `json:"tableoid"`       // tableoid
	Cmax         Cid           `json:"cmax"`           // cmax
	Xmax         Xid           `json:"xmax"`           // xmax
	Cmin         Cid           `json:"cmin"`           // cmin
	Xmin         Xid           `json:"xmin"`           // xmin
	Ctid         Tid           `json:"ctid"`           // ctid
	FeatureID    CharacterData `json:"feature_id"`     // feature_id
	FeatureName  CharacterData `json:"feature_name"`   // feature_name
	IsSupported  YesOrNo       `json:"is_supported"`   // is_supported
	IsVerifiedBy CharacterData `json:"is_verified_by"` // is_verified_by
	Comments     CharacterData `json:"comments"`       // comments
}
