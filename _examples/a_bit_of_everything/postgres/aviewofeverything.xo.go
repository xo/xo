package postgres

// Code generated by xo. DO NOT EDIT.

import (
	"database/sql"

	"github.com/google/uuid"
)

// SELECT a_enum,     a_enum_nullable,     a_bigint,     a_bigint_nullable,     a_bigserial,     a_bigserial_nullable,     a_bit,     a_bit_nullable,     a_bit_varying,     a_bit_varying_nullable,     a_bool,     a_bool_nullable,     a_boolean,     a_boolean_nullable,     a_bpchar,     a_bpchar_nullable,     a_bytea,     a_bytea_nullable,     a_char,     a_char_nullable,     a_character,     a_character_nullable,     a_character_varying,     a_character_varying_nullable,     a_date,     a_date_nullable,     a_decimal,     a_decimal_nullable,     a_double_precision,     a_double_precision_nullable,     a_inet,     a_inet_nullable,     a_int,     a_int_nullable,     a_integer,     a_integer_nullable,     a_interval,     a_interval_nullable,     a_json,     a_json_nullable,     a_jsonb,     a_jsonb_nullable,     a_money,     a_money_nullable,     a_numeric,     a_numeric_nullable,     a_real,     a_real_nullable,     a_serial,     a_serial_nullable,     a_smallint,     a_smallint_nullable,     a_smallserial,     a_smallserial_nullable,     a_text,     a_text_nullable,     a_time,     a_time_nullable,     a_timestamp,     a_timestamp_nullable,     a_timestamptz,     a_timestamptz_nullable,     a_timetz,     a_timetz_nullable,     a_uuid,     a_uuid_nullable,     a_varchar,     a_varchar_nullable,     a_xml,     a_xml_nullable    FROM a_bit_of_everything;
type AViewOfEverything struct {
	AEnum                     NullAEnum       `json:"a_enum"`                       // a_enum
	AEnumNullable             NullAEnum       `json:"a_enum_nullable"`              // a_enum_nullable
	ABigint                   sql.NullInt64   `json:"a_bigint"`                     // a_bigint
	ABigintNullable           sql.NullInt64   `json:"a_bigint_nullable"`            // a_bigint_nullable
	ABigserial                sql.NullInt64   `json:"a_bigserial"`                  // a_bigserial
	ABigserialNullable        sql.NullInt64   `json:"a_bigserial_nullable"`         // a_bigserial_nullable
	ABit                      *uint8          `json:"a_bit"`                        // a_bit
	ABitNullable              *uint8          `json:"a_bit_nullable"`               // a_bit_nullable
	ABitVarying               []byte          `json:"a_bit_varying"`                // a_bit_varying
	ABitVaryingNullable       []byte          `json:"a_bit_varying_nullable"`       // a_bit_varying_nullable
	ABool                     sql.NullBool    `json:"a_bool"`                       // a_bool
	ABoolNullable             sql.NullBool    `json:"a_bool_nullable"`              // a_bool_nullable
	ABoolean                  sql.NullBool    `json:"a_boolean"`                    // a_boolean
	ABooleanNullable          sql.NullBool    `json:"a_boolean_nullable"`           // a_boolean_nullable
	ABpchar                   sql.NullString  `json:"a_bpchar"`                     // a_bpchar
	ABpcharNullable           sql.NullString  `json:"a_bpchar_nullable"`            // a_bpchar_nullable
	ABytea                    []byte          `json:"a_bytea"`                      // a_bytea
	AByteaNullable            []byte          `json:"a_bytea_nullable"`             // a_bytea_nullable
	AChar                     sql.NullString  `json:"a_char"`                       // a_char
	ACharNullable             sql.NullString  `json:"a_char_nullable"`              // a_char_nullable
	ACharacter                sql.NullString  `json:"a_character"`                  // a_character
	ACharacterNullable        sql.NullString  `json:"a_character_nullable"`         // a_character_nullable
	ACharacterVarying         sql.NullString  `json:"a_character_varying"`          // a_character_varying
	ACharacterVaryingNullable sql.NullString  `json:"a_character_varying_nullable"` // a_character_varying_nullable
	ADate                     sql.NullTime    `json:"a_date"`                       // a_date
	ADateNullable             sql.NullTime    `json:"a_date_nullable"`              // a_date_nullable
	ADecimal                  sql.NullFloat64 `json:"a_decimal"`                    // a_decimal
	ADecimalNullable          sql.NullFloat64 `json:"a_decimal_nullable"`           // a_decimal_nullable
	ADoublePrecision          sql.NullFloat64 `json:"a_double_precision"`           // a_double_precision
	ADoublePrecisionNullable  sql.NullFloat64 `json:"a_double_precision_nullable"`  // a_double_precision_nullable
	AInet                     sql.NullString  `json:"a_inet"`                       // a_inet
	AInetNullable             sql.NullString  `json:"a_inet_nullable"`              // a_inet_nullable
	AInt                      sql.NullInt64   `json:"a_int"`                        // a_int
	AIntNullable              sql.NullInt64   `json:"a_int_nullable"`               // a_int_nullable
	AInteger                  sql.NullInt64   `json:"a_integer"`                    // a_integer
	AIntegerNullable          sql.NullInt64   `json:"a_integer_nullable"`           // a_integer_nullable
	AInterval                 []byte          `json:"a_interval"`                   // a_interval
	AIntervalNullable         []byte          `json:"a_interval_nullable"`          // a_interval_nullable
	AJSON                     []byte          `json:"a_json"`                       // a_json
	AJSONNullable             []byte          `json:"a_json_nullable"`              // a_json_nullable
	AJsonb                    []byte          `json:"a_jsonb"`                      // a_jsonb
	AJsonbNullable            []byte          `json:"a_jsonb_nullable"`             // a_jsonb_nullable
	AMoney                    sql.NullString  `json:"a_money"`                      // a_money
	AMoneyNullable            sql.NullString  `json:"a_money_nullable"`             // a_money_nullable
	ANumeric                  sql.NullFloat64 `json:"a_numeric"`                    // a_numeric
	ANumericNullable          sql.NullFloat64 `json:"a_numeric_nullable"`           // a_numeric_nullable
	AReal                     sql.NullFloat64 `json:"a_real"`                       // a_real
	ARealNullable             sql.NullFloat64 `json:"a_real_nullable"`              // a_real_nullable
	ASerial                   sql.NullInt64   `json:"a_serial"`                     // a_serial
	ASerialNullable           sql.NullInt64   `json:"a_serial_nullable"`            // a_serial_nullable
	ASmallint                 sql.NullInt64   `json:"a_smallint"`                   // a_smallint
	ASmallintNullable         sql.NullInt64   `json:"a_smallint_nullable"`          // a_smallint_nullable
	ASmallserial              sql.NullInt64   `json:"a_smallserial"`                // a_smallserial
	ASmallserialNullable      sql.NullInt64   `json:"a_smallserial_nullable"`       // a_smallserial_nullable
	AText                     sql.NullString  `json:"a_text"`                       // a_text
	ATextNullable             sql.NullString  `json:"a_text_nullable"`              // a_text_nullable
	ATime                     sql.NullTime    `json:"a_time"`                       // a_time
	ATimeNullable             sql.NullTime    `json:"a_time_nullable"`              // a_time_nullable
	ATimestamp                sql.NullTime    `json:"a_timestamp"`                  // a_timestamp
	ATimestampNullable        sql.NullTime    `json:"a_timestamp_nullable"`         // a_timestamp_nullable
	ATimestamptz              sql.NullTime    `json:"a_timestamptz"`                // a_timestamptz
	ATimestamptzNullable      sql.NullTime    `json:"a_timestamptz_nullable"`       // a_timestamptz_nullable
	ATimetz                   sql.NullTime    `json:"a_timetz"`                     // a_timetz
	ATimetzNullable           sql.NullTime    `json:"a_timetz_nullable"`            // a_timetz_nullable
	AUUID                     uuid.NullUUID   `json:"a_uuid"`                       // a_uuid
	AUUIDNullable             uuid.NullUUID   `json:"a_uuid_nullable"`              // a_uuid_nullable
	AVarchar                  sql.NullString  `json:"a_varchar"`                    // a_varchar
	AVarcharNullable          sql.NullString  `json:"a_varchar_nullable"`           // a_varchar_nullable
	AXML                      []byte          `json:"a_xml"`                        // a_xml
	AXMLNullable              []byte          `json:"a_xml_nullable"`               // a_xml_nullable
}
