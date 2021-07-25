-- table with manual primary key
-- generate insert only (no update, save, upsert, delete)
CREATE TABLE a_manual_table (
  a_text NVARCHAR2(255)
);

-- table with auto increment
CREATE TABLE a_sequence (
  a_seq NUMBER GENERATED ALWAYS AS IDENTITY,
  CONSTRAINT a_sequence_pkey PRIMARY KEY (a_seq)
);

CREATE TABLE a_sequence_multi (
  a_seq INTEGER GENERATED ALWAYS AS IDENTITY,
  a_text NVARCHAR2(255),
  CONSTRAINT a_sequence_multi_pkey PRIMARY KEY (a_seq)
);

-- table with primary key
CREATE TABLE a_primary (
  a_key INTEGER NOT NULL,
  CONSTRAINT a_primary_pkey PRIMARY KEY (a_key)
);

CREATE TABLE a_primary_multi (
  a_key INTEGER NOT NULL CONSTRAINT a_primary_multi_pkey PRIMARY KEY,
  a_text NVARCHAR2(255)
);

-- table with composite primary key
CREATE TABLE a_primary_composite (
  a_key1 INTEGER,
  a_key2 INTEGER,
  CONSTRAINT a_primary_composite_pkey PRIMARY KEY (a_key1, a_key2)
);

-- table with foreign key
CREATE TABLE a_foreign_key (
  a_key INTEGER CONSTRAINT a_key_fkey REFERENCES a_primary (a_key)
);

-- table with composite foreign key
CREATE TABLE a_foreign_key_composite (
  a_key1 INTEGER,
  a_key2 INTEGER,
  CONSTRAINT a_foreign_key_composite_fkey FOREIGN KEY(a_key1, a_key2) REFERENCES a_primary_composite (a_key1, a_key2)
);

-- table with index
CREATE TABLE a_index (
  a_key INTEGER
);

CREATE INDEX a_index_idx ON a_index (a_key);

-- table with composite index
CREATE TABLE a_index_composite (
  a_key1 INTEGER,
  a_key2 INTEGER
);

CREATE INDEX a_index_composite_idx ON a_index_composite (a_key1, a_key2);

-- table with unique index
CREATE TABLE a_unique_index (
  a_key INTEGER CONSTRAINT a_unique_index_idx UNIQUE
);

-- table with composite unique index
CREATE TABLE a_unique_index_composite (
  a_key1 INTEGER,
  a_key2 INTEGER,
  CONSTRAINT a_unique_index_composite_idx UNIQUE (a_key1, a_key2)
);

/*

bool
blob
char
clob
date
double precision
decimal
float
int
integer
long raw
nchar
nclob
number
nvarchar2
raw
real
rowid
smallint
timestamp
timestamp with local time zone
timestamp with time zone
varchar
varchar2
xmltype

*/
CREATE TABLE a_bit_of_everything (
  a_bool NUMBER(1) NOT NULL,
  a_bool_nullable NUMBER(1),
  a_blob BLOB NOT NULL,
  a_blob_nullable BLOB,
  a_char CHAR(255) NOT NULL,
  a_char_nullable CHAR(255),
  a_character CHARACTER(255) NOT NULL,
  a_character_nullable CHARACTER(255),
  a_clob CLOB NOT NULL,
  a_clob_nullable CLOB,
  a_date DATE NOT NULL,
  a_date_nullable DATE,
  a_double_precision DOUBLE PRECISION NOT NULL,
  a_double_precision_nullable DOUBLE PRECISION,
  a_decimal DECIMAL NOT NULL,
  a_decimal_nullable DECIMAL,
  a_float FLOAT NOT NULL,
  a_float_nullable FLOAT,
  a_int INT NOT NULL,
  a_int_nullable INT,
  a_integer INT NOT NULL,
  a_integer_nullable INT,
  a_long_raw LONG RAW NOT NULL,
  a_nchar NCHAR(255) NOT NULL,
  a_nchar_nullable NCHAR(255),
  a_nclob NCLOB NOT NULL,
  a_nclob_nullable NCLOB,
  a_number NUMBER(6) NOT NULL,
  a_number_nullable NUMBER(6),
  a_numeric NUMERIC NOT NULL,
  a_numeric_nullable NUMERIC,
  a_nvarchar2 NVARCHAR2(255) NOT NULL,
  a_nvarchar2_nullable NVARCHAR2(255),
  a_raw RAW(2000) NOT NULL,
  a_raw_nullable RAW(2000),
  a_real REAL NOT NULL,
  a_real_nullable REAL,
  a_rowid ROWID NOT NULL,
  a_rowid_nullable ROWID,
  a_smallint SMALLINT NOT NULL,
  a_smallint_nullable SMALLINT,
  a_timestamp TIMESTAMP NOT NULL,
  a_timestamp_nullable TIMESTAMP,
  a_timestamp_with_local_time_zone TIMESTAMP WITH LOCAL TIME ZONE NOT NULL,
  a_timestamp_with_local_time_zone_nullable TIMESTAMP WITH LOCAL TIME ZONE,
  a_timestamp_with_time_zone TIMESTAMP WITH TIME ZONE NOT NULL,
  a_timestamp_with_time_zone_nullable TIMESTAMP WITH TIME ZONE,
  a_varchar VARCHAR(255) NOT NULL,
  a_varchar_nullable VARCHAR(255),
  a_varchar2 VARCHAR2(255) NOT NULL,
  a_varchar2_nullable VARCHAR2(255),
  a_xmltype XMLTYPE NOT NULL,
  a_xmltype_nullable XMLTYPE
);

-- views
CREATE VIEW a_view_of_everything AS
  SELECT * FROM a_bit_of_everything;

CREATE VIEW a_view_of_everything_some AS
  SELECT a_bool, a_nclob FROM a_bit_of_everything;

-- procs
CREATE PROCEDURE a_0_in_0_out AS
BEGIN
  null\;
END\;;

CREATE PROCEDURE a_1_in_0_out(a_param IN INTEGER) AS
BEGIN
  null\;
END\;;

CREATE PROCEDURE a_0_in_1_out(a_return OUT INTEGER) AS
BEGIN
  SELECT 10 INTO a_return FROM dual\;
END\;;

CREATE PROCEDURE a_1_in_1_out(a_param IN INTEGER, a_return OUT INTEGER) AS
BEGIN
  SELECT a_param INTO a_return FROM dual\;
END\;;

CREATE PROCEDURE a_2_in_2_out(param_one IN INTEGER, param_two IN INTEGER, return_one OUT INTEGER, return_two OUT INTEGER) AS
BEGIN
  SELECT param_one, param_two INTO return_one, return_two FROM dual\;
END\;;

CREATE FUNCTION a_func_0_in RETURN INTEGER AS
BEGIN
  RETURN 10\;
end\;;

CREATE FUNCTION a_func_1_in(a_param IN INTEGER) RETURN INTEGER AS
BEGIN
  RETURN a_param\;
END\;;

CREATE FUNCTION a_func_2_in(param_one IN INTEGER, param_two IN INTEGER) RETURN INTEGER AS
BEGIN
  RETURN param_one + param_two\;
END\;;
