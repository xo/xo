package sqlite3

// Code generated by xo. DO NOT EDIT.

import (
	"database/sql"
)

// AViewOfEverything represents a row from 'a_view_of_everything'.
type AViewOfEverything struct {
	ABigint                           sql.NullInt64   `json:"a_bigint"`                              // a_bigint
	ABigintNullable                   sql.NullInt64   `json:"a_bigint_nullable"`                     // a_bigint_nullable
	ABlob                             []byte          `json:"a_blob"`                                // a_blob
	ABlobNullable                     []byte          `json:"a_blob_nullable"`                       // a_blob_nullable
	ABool                             sql.NullBool    `json:"a_bool"`                                // a_bool
	ABoolNullable                     sql.NullBool    `json:"a_bool_nullable"`                       // a_bool_nullable
	ABoolean                          sql.NullBool    `json:"a_boolean"`                             // a_boolean
	ABooleanNullable                  sql.NullBool    `json:"a_boolean_nullable"`                    // a_boolean_nullable
	ACharacter                        sql.NullString  `json:"a_character"`                           // a_character
	ACharacterNullable                sql.NullString  `json:"a_character_nullable"`                  // a_character_nullable
	AClob                             sql.NullString  `json:"a_clob"`                                // a_clob
	AClobNullable                     sql.NullString  `json:"a_clob_nullable"`                       // a_clob_nullable
	ADate                             *Time           `json:"a_date"`                                // a_date
	ADateNullable                     *Time           `json:"a_date_nullable"`                       // a_date_nullable
	ADatetime                         *Time           `json:"a_datetime"`                            // a_datetime
	ADatetimeNullable                 *Time           `json:"a_datetime_nullable"`                   // a_datetime_nullable
	ADecimal                          sql.NullFloat64 `json:"a_decimal"`                             // a_decimal
	ADecimalNullable                  sql.NullFloat64 `json:"a_decimal_nullable"`                    // a_decimal_nullable
	ADouble                           sql.NullFloat64 `json:"a_double"`                              // a_double
	ADoubleNullable                   sql.NullFloat64 `json:"a_double_nullable"`                     // a_double_nullable
	AFloat                            sql.NullFloat64 `json:"a_float"`                               // a_float
	AFloatNullable                    sql.NullFloat64 `json:"a_float_nullable"`                      // a_float_nullable
	AInt                              sql.NullInt64   `json:"a_int"`                                 // a_int
	AIntNullable                      sql.NullInt64   `json:"a_int_nullable"`                        // a_int_nullable
	AInteger                          sql.NullInt64   `json:"a_integer"`                             // a_integer
	AIntegerNullable                  sql.NullInt64   `json:"a_integer_nullable"`                    // a_integer_nullable
	AMediumint                        sql.NullInt64   `json:"a_mediumint"`                           // a_mediumint
	AMediumintNullable                sql.NullInt64   `json:"a_mediumint_nullable"`                  // a_mediumint_nullable
	ANativeCharacter                  sql.NullString  `json:"a_native_character"`                    // a_native_character
	ANativeCharacterNullable          sql.NullString  `json:"a_native_character_nullable"`           // a_native_character_nullable
	ANchar                            sql.NullString  `json:"a_nchar"`                               // a_nchar
	ANcharNullable                    sql.NullString  `json:"a_nchar_nullable"`                      // a_nchar_nullable
	ANumeric                          sql.NullFloat64 `json:"a_numeric"`                             // a_numeric
	ANumericNullable                  sql.NullFloat64 `json:"a_numeric_nullable"`                    // a_numeric_nullable
	ANvarchar                         sql.NullString  `json:"a_nvarchar"`                            // a_nvarchar
	ANvarcharNullable                 sql.NullString  `json:"a_nvarchar_nullable"`                   // a_nvarchar_nullable
	AReal                             sql.NullFloat64 `json:"a_real"`                                // a_real
	ARealNullable                     sql.NullFloat64 `json:"a_real_nullable"`                       // a_real_nullable
	ASmallint                         sql.NullInt64   `json:"a_smallint"`                            // a_smallint
	ASmallintNullable                 sql.NullInt64   `json:"a_smallint_nullable"`                   // a_smallint_nullable
	AText                             sql.NullString  `json:"a_text"`                                // a_text
	ATextNullable                     sql.NullString  `json:"a_text_nullable"`                       // a_text_nullable
	ATime                             sql.NullString  `json:"a_time"`                                // a_time
	ATimeNullable                     sql.NullString  `json:"a_time_nullable"`                       // a_time_nullable
	ATimestamp                        *Time           `json:"a_timestamp"`                           // a_timestamp
	ATimestampNullable                *Time           `json:"a_timestamp_nullable"`                  // a_timestamp_nullable
	ATimestampWithoutTimezone         *Time           `json:"a_timestamp_without_timezone"`          // a_timestamp_without_timezone
	ATimestampWithoutTimezoneNullable *Time           `json:"a_timestamp_without_timezone_nullable"` // a_timestamp_without_timezone_nullable
	ATimestampWithTimezone            *Time           `json:"a_timestamp_with_timezone"`             // a_timestamp_with_timezone
	ATimestampWithTimezoneNullable    *Time           `json:"a_timestamp_with_timezone_nullable"`    // a_timestamp_with_timezone_nullable
	ATimeWithoutTimezone              *Time           `json:"a_time_without_timezone"`               // a_time_without_timezone
	ATimeWithoutTimezoneNullable      *Time           `json:"a_time_without_timezone_nullable"`      // a_time_without_timezone_nullable
	ATimeWithTimezone                 *Time           `json:"a_time_with_timezone"`                  // a_time_with_timezone
	ATimeWithTimezoneNullable         *Time           `json:"a_time_with_timezone_nullable"`         // a_time_with_timezone_nullable
	ATinyint                          sql.NullInt64   `json:"a_tinyint"`                             // a_tinyint
	ATinyintNullable                  sql.NullInt64   `json:"a_tinyint_nullable"`                    // a_tinyint_nullable
	AVarchar                          sql.NullString  `json:"a_varchar"`                             // a_varchar
	AVarcharNullable                  sql.NullString  `json:"a_varchar_nullable"`                    // a_varchar_nullable
	AVaryingCharacter                 sql.NullString  `json:"a_varying_character"`                   // a_varying_character
	AVaryingCharacterNullable         sql.NullString  `json:"a_varying_character_nullable"`          // a_varying_character_nullable
}