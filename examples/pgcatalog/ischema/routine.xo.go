// Package ischema contains the types for schema 'information_schema'.
package ischema

import "github.com/knq/xo/examples/pgcatalog/pgtypes"

// Code generated by xo. DO NOT EDIT.

// Routine represents a row from 'information_schema.routines'.
type Routine struct {
	SpecificCatalog                 pgtypes.SQLIdentifier  `json:"specific_catalog"`                    // specific_catalog
	SpecificSchema                  pgtypes.SQLIdentifier  `json:"specific_schema"`                     // specific_schema
	SpecificName                    pgtypes.SQLIdentifier  `json:"specific_name"`                       // specific_name
	RoutineCatalog                  pgtypes.SQLIdentifier  `json:"routine_catalog"`                     // routine_catalog
	RoutineSchema                   pgtypes.SQLIdentifier  `json:"routine_schema"`                      // routine_schema
	RoutineName                     pgtypes.SQLIdentifier  `json:"routine_name"`                        // routine_name
	RoutineType                     pgtypes.CharacterData  `json:"routine_type"`                        // routine_type
	ModuleCatalog                   pgtypes.SQLIdentifier  `json:"module_catalog"`                      // module_catalog
	ModuleSchema                    pgtypes.SQLIdentifier  `json:"module_schema"`                       // module_schema
	ModuleName                      pgtypes.SQLIdentifier  `json:"module_name"`                         // module_name
	UdtCatalog                      pgtypes.SQLIdentifier  `json:"udt_catalog"`                         // udt_catalog
	UdtSchema                       pgtypes.SQLIdentifier  `json:"udt_schema"`                          // udt_schema
	UdtName                         pgtypes.SQLIdentifier  `json:"udt_name"`                            // udt_name
	DataType                        pgtypes.CharacterData  `json:"data_type"`                           // data_type
	CharacterMaximumLength          pgtypes.CardinalNumber `json:"character_maximum_length"`            // character_maximum_length
	CharacterOctetLength            pgtypes.CardinalNumber `json:"character_octet_length"`              // character_octet_length
	CharacterSetCatalog             pgtypes.SQLIdentifier  `json:"character_set_catalog"`               // character_set_catalog
	CharacterSetSchema              pgtypes.SQLIdentifier  `json:"character_set_schema"`                // character_set_schema
	CharacterSetName                pgtypes.SQLIdentifier  `json:"character_set_name"`                  // character_set_name
	CollationCatalog                pgtypes.SQLIdentifier  `json:"collation_catalog"`                   // collation_catalog
	CollationSchema                 pgtypes.SQLIdentifier  `json:"collation_schema"`                    // collation_schema
	CollationName                   pgtypes.SQLIdentifier  `json:"collation_name"`                      // collation_name
	NumericPrecision                pgtypes.CardinalNumber `json:"numeric_precision"`                   // numeric_precision
	NumericPrecisionRadix           pgtypes.CardinalNumber `json:"numeric_precision_radix"`             // numeric_precision_radix
	NumericScale                    pgtypes.CardinalNumber `json:"numeric_scale"`                       // numeric_scale
	DatetimePrecision               pgtypes.CardinalNumber `json:"datetime_precision"`                  // datetime_precision
	IntervalType                    pgtypes.CharacterData  `json:"interval_type"`                       // interval_type
	IntervalPrecision               pgtypes.CardinalNumber `json:"interval_precision"`                  // interval_precision
	TypeUdtCatalog                  pgtypes.SQLIdentifier  `json:"type_udt_catalog"`                    // type_udt_catalog
	TypeUdtSchema                   pgtypes.SQLIdentifier  `json:"type_udt_schema"`                     // type_udt_schema
	TypeUdtName                     pgtypes.SQLIdentifier  `json:"type_udt_name"`                       // type_udt_name
	ScopeCatalog                    pgtypes.SQLIdentifier  `json:"scope_catalog"`                       // scope_catalog
	ScopeSchema                     pgtypes.SQLIdentifier  `json:"scope_schema"`                        // scope_schema
	ScopeName                       pgtypes.SQLIdentifier  `json:"scope_name"`                          // scope_name
	MaximumCardinality              pgtypes.CardinalNumber `json:"maximum_cardinality"`                 // maximum_cardinality
	DtdIdentifier                   pgtypes.SQLIdentifier  `json:"dtd_identifier"`                      // dtd_identifier
	RoutineBody                     pgtypes.CharacterData  `json:"routine_body"`                        // routine_body
	RoutineDefinition               pgtypes.CharacterData  `json:"routine_definition"`                  // routine_definition
	ExternalName                    pgtypes.CharacterData  `json:"external_name"`                       // external_name
	ExternalLanguage                pgtypes.CharacterData  `json:"external_language"`                   // external_language
	ParameterStyle                  pgtypes.CharacterData  `json:"parameter_style"`                     // parameter_style
	IsDeterministic                 pgtypes.YesOrNo        `json:"is_deterministic"`                    // is_deterministic
	SQLDataAccess                   pgtypes.CharacterData  `json:"sql_data_access"`                     // sql_data_access
	IsNullCall                      pgtypes.YesOrNo        `json:"is_null_call"`                        // is_null_call
	SQLPath                         pgtypes.CharacterData  `json:"sql_path"`                            // sql_path
	SchemaLevelRoutine              pgtypes.YesOrNo        `json:"schema_level_routine"`                // schema_level_routine
	MaxDynamicResultSets            pgtypes.CardinalNumber `json:"max_dynamic_result_sets"`             // max_dynamic_result_sets
	IsUserDefinedCast               pgtypes.YesOrNo        `json:"is_user_defined_cast"`                // is_user_defined_cast
	IsImplicitlyInvocable           pgtypes.YesOrNo        `json:"is_implicitly_invocable"`             // is_implicitly_invocable
	SecurityType                    pgtypes.CharacterData  `json:"security_type"`                       // security_type
	ToSQLSpecificCatalog            pgtypes.SQLIdentifier  `json:"to_sql_specific_catalog"`             // to_sql_specific_catalog
	ToSQLSpecificSchema             pgtypes.SQLIdentifier  `json:"to_sql_specific_schema"`              // to_sql_specific_schema
	ToSQLSpecificName               pgtypes.SQLIdentifier  `json:"to_sql_specific_name"`                // to_sql_specific_name
	AsLocator                       pgtypes.YesOrNo        `json:"as_locator"`                          // as_locator
	Created                         pgtypes.TimeStamp      `json:"created"`                             // created
	LastAltered                     pgtypes.TimeStamp      `json:"last_altered"`                        // last_altered
	NewSavepointLevel               pgtypes.YesOrNo        `json:"new_savepoint_level"`                 // new_savepoint_level
	IsUdtDependent                  pgtypes.YesOrNo        `json:"is_udt_dependent"`                    // is_udt_dependent
	ResultCastFromDataType          pgtypes.CharacterData  `json:"result_cast_from_data_type"`          // result_cast_from_data_type
	ResultCastAsLocator             pgtypes.YesOrNo        `json:"result_cast_as_locator"`              // result_cast_as_locator
	ResultCastCharMaxLength         pgtypes.CardinalNumber `json:"result_cast_char_max_length"`         // result_cast_char_max_length
	ResultCastCharOctetLength       pgtypes.CardinalNumber `json:"result_cast_char_octet_length"`       // result_cast_char_octet_length
	ResultCastCharSetCatalog        pgtypes.SQLIdentifier  `json:"result_cast_char_set_catalog"`        // result_cast_char_set_catalog
	ResultCastCharSetSchema         pgtypes.SQLIdentifier  `json:"result_cast_char_set_schema"`         // result_cast_char_set_schema
	ResultCastCharSetName           pgtypes.SQLIdentifier  `json:"result_cast_char_set_name"`           // result_cast_char_set_name
	ResultCastCollationCatalog      pgtypes.SQLIdentifier  `json:"result_cast_collation_catalog"`       // result_cast_collation_catalog
	ResultCastCollationSchema       pgtypes.SQLIdentifier  `json:"result_cast_collation_schema"`        // result_cast_collation_schema
	ResultCastCollationName         pgtypes.SQLIdentifier  `json:"result_cast_collation_name"`          // result_cast_collation_name
	ResultCastNumericPrecision      pgtypes.CardinalNumber `json:"result_cast_numeric_precision"`       // result_cast_numeric_precision
	ResultCastNumericPrecisionRadix pgtypes.CardinalNumber `json:"result_cast_numeric_precision_radix"` // result_cast_numeric_precision_radix
	ResultCastNumericScale          pgtypes.CardinalNumber `json:"result_cast_numeric_scale"`           // result_cast_numeric_scale
	ResultCastDatetimePrecision     pgtypes.CardinalNumber `json:"result_cast_datetime_precision"`      // result_cast_datetime_precision
	ResultCastIntervalType          pgtypes.CharacterData  `json:"result_cast_interval_type"`           // result_cast_interval_type
	ResultCastIntervalPrecision     pgtypes.CardinalNumber `json:"result_cast_interval_precision"`      // result_cast_interval_precision
	ResultCastTypeUdtCatalog        pgtypes.SQLIdentifier  `json:"result_cast_type_udt_catalog"`        // result_cast_type_udt_catalog
	ResultCastTypeUdtSchema         pgtypes.SQLIdentifier  `json:"result_cast_type_udt_schema"`         // result_cast_type_udt_schema
	ResultCastTypeUdtName           pgtypes.SQLIdentifier  `json:"result_cast_type_udt_name"`           // result_cast_type_udt_name
	ResultCastScopeCatalog          pgtypes.SQLIdentifier  `json:"result_cast_scope_catalog"`           // result_cast_scope_catalog
	ResultCastScopeSchema           pgtypes.SQLIdentifier  `json:"result_cast_scope_schema"`            // result_cast_scope_schema
	ResultCastScopeName             pgtypes.SQLIdentifier  `json:"result_cast_scope_name"`              // result_cast_scope_name
	ResultCastMaximumCardinality    pgtypes.CardinalNumber `json:"result_cast_maximum_cardinality"`     // result_cast_maximum_cardinality
	ResultCastDtdIdentifier         pgtypes.SQLIdentifier  `json:"result_cast_dtd_identifier"`          // result_cast_dtd_identifier
}
