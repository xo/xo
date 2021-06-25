-- table with manual primary key
-- generate insert only (no update, save, upsert, delete)
CREATE TABLE a_manual_table (
  a_text NVARCHAR(255)
);

-- table with auto increment
CREATE TABLE a_sequence (
  a_seq INTEGER IDENTITY(1, 1) NOT NULL CONSTRAINT a_sequence_pkey PRIMARY KEY
);

CREATE TABLE a_sequence_multi (
  a_seq INTEGER IDENTITY(1, 1) NOT NULL CONSTRAINT a_sequence_multi_pkey PRIMARY KEY,
  a_text NVARCHAR(255)
);

-- table with primary key
CREATE TABLE a_primary (
  a_key INTEGER NOT NULL CONSTRAINT a_primary_pkey PRIMARY KEY
);

CREATE TABLE a_primary_multi (
  a_key INTEGER NOT NULL CONSTRAINT a_primary_multi_pkey PRIMARY KEY,
  a_text NVARCHAR(255)
);

-- table with composite primary key
CREATE TABLE a_primary_composite (
  a_key1 INTEGER,
  a_key2 INTEGER,
  CONSTRAINT a_primary_composite_pkey PRIMARY KEY (a_key1, a_key2)
);

-- table with foreign key
CREATE TABLE a_foreign_key (
  a_key INTEGER CONSTRAINT a_key_fkey REFERENCES a_primary(a_key)
);

-- table with composite foreign key
CREATE TABLE a_foreign_key_composite (
  a_key1 INTEGER,
  a_key2 INTEGER,
  CONSTRAINT a_foreign_key_composite_fkey FOREIGN KEY(a_key1, a_key2) REFERENCES a_primary_composite(a_key1, a_key2)
);

-- table with index
CREATE TABLE a_index (
  a_key INTEGER
);

CREATE INDEX a_index_idx ON a_index(a_key);

-- table with composite index
CREATE TABLE a_index_composite (
  a_key1 INTEGER,
  a_key2 INTEGER
);

CREATE INDEX a_index_composite_idx ON a_index_composite(a_key1, a_key2);

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

bigint
binary
bit
char
date
datetime
datetime2
datetimeoffset
decimal
float
image
int
money
nchar
ntext
numeric
nvarchar
real
smalldatetime
smallint
smallmoney
text
time
tinyint
varbinary
varchar
xml

*/

CREATE TABLE a_bit_of_everything (
  a_bigint BIGINT NOT NULL,
  a_bigint_nullable BIGINT,
  a_binary BINARY NOT NULL,
  a_binary_nullable BINARY,
  a_bit BIT NOT NULL,
  a_bit_nullable BIT,
  a_char CHAR NOT NULL,
  a_char_nullable CHAR,
  a_date DATE NOT NULL,
  a_date_nullable DATE,
  a_datetime DATETIME NOT NULL,
  a_datetime_nullable DATETIME,
  a_datetime2 DATETIME2 NOT NULL,
  a_datetime2_nullable DATETIME2,
  a_datetimeoffset DATETIMEOFFSET NOT NULL,
  a_datetimeoffset_nullable DATETIMEOFFSET,
  a_decimal DECIMAL NOT NULL,
  a_decimal_nullable DECIMAL,
  a_float FLOAT NOT NULL,
  a_float_nullable FLOAT,
  a_image IMAGE NOT NULL,
  a_image_nullable IMAGE,
  a_int INT NOT NULL,
  a_int_nullable INT,
  a_money MONEY NOT NULL,
  a_money_nullable MONEY,
  a_nchar NCHAR NOT NULL,
  a_nchar_nullable NCHAR,
  a_ntext NTEXT NOT NULL,
  a_ntext_nullable NTEXT,
  a_numeric NUMERIC NOT NULL,
  a_numeric_nullable NUMERIC,
  a_nvarchar NVARCHAR NOT NULL,
  a_nvarchar_nullable NVARCHAR,
  a_real REAL NOT NULL,
  a_real_nullable REAL,
  a_smalldatetime SMALLDATETIME NOT NULL,
  a_smalldatetime_nullable SMALLDATETIME,
  a_smallint SMALLINT NOT NULL,
  a_smallint_nullable SMALLINT,
  a_smallmoney SMALLMONEY NOT NULL,
  a_smallmoney_nullable SMALLMONEY,
  a_text TEXT NOT NULL,
  a_text_nullable TEXT,
  a_time TIME NOT NULL,
  a_time_nullable TIME,
  a_tinyint TINYINT NOT NULL,
  a_tinyint_nullable TINYINT,
  a_varbinary VARBINARY NOT NULL,
  a_varbinary_nullable VARBINARY,
  a_varchar VARCHAR NOT NULL,
  a_varchar_nullable VARCHAR,
  a_xml XML NOT NULL,
  a_xml_nullable XML
);

-- views
CREATE VIEW a_view_of_everything AS
  SELECT * FROM a_bit_of_everything;
