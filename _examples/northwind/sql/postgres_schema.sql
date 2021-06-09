DROP TABLE IF EXISTS customer_customer_demo;
DROP TABLE IF EXISTS customer_demographics;
DROP TABLE IF EXISTS employee_territories;
DROP TABLE IF EXISTS order_details;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS customers;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS shippers;
DROP TABLE IF EXISTS suppliers;
DROP TABLE IF EXISTS territories;
DROP TABLE IF EXISTS us_states;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS region;
DROP TABLE IF EXISTS employees;

CREATE TABLE categories (
  category_id SMALLINT NOT NULL PRIMARY KEY,
  category_name CHARACTER VARYING(15) NOT NULL,
  description TEXT,
  picture BYTEA
);

CREATE TABLE customer_demographics (
  customer_type_id BPCHAR NOT NULL PRIMARY KEY,
  customer_desc TEXT
);

CREATE TABLE customers (
  customer_id BPCHAR NOT NULL PRIMARY KEY,
  company_name CHARACTER VARYING(40) NOT NULL,
  contact_name CHARACTER VARYING(30),
  contact_title CHARACTER VARYING(30),
  address CHARACTER VARYING(60),
  city CHARACTER VARYING(15),
  region CHARACTER VARYING(15),
  postal_code CHARACTER VARYING(10),
  country CHARACTER VARYING(15),
  phone CHARACTER VARYING(24),
  fax CHARACTER VARYING(24)
);

CREATE TABLE customer_customer_demo (
  customer_id BPCHAR NOT NULL REFERENCES customers,
  customer_type_id BPCHAR NOT NULL REFERENCES customer_demographics,
  PRIMARY KEY (customer_id, customer_type_id)
);

CREATE TABLE employees (
  employee_id SMALLINT NOT NULL PRIMARY KEY,
  last_name CHARACTER VARYING(20) NOT NULL,
  first_name CHARACTER VARYING(10) NOT NULL,
  title CHARACTER VARYING(30),
  title_of_courtesy CHARACTER VARYING(25),
  birth_date DATE,
  hire_date DATE,
  address CHARACTER VARYING(60),
  city CHARACTER VARYING(15),
  region CHARACTER VARYING(15),
  postal_code CHARACTER VARYING(10),
  country CHARACTER VARYING(15),
  home_phone CHARACTER VARYING(24),
  extension CHARACTER VARYING(4),
  photo BYTEA,
  notes TEXT,
  reports_to SMALLINT REFERENCES employees,
  photo_path CHARACTER VARYING(255)
);

CREATE TABLE suppliers (
  supplier_id SMALLINT NOT NULL PRIMARY KEY,
  company_name CHARACTER VARYING(40) NOT NULL,
  contact_name CHARACTER VARYING(30),
  contact_title CHARACTER VARYING(30),
  address CHARACTER VARYING(60),
  city CHARACTER VARYING(15),
  region CHARACTER VARYING(15),
  postal_code CHARACTER VARYING(10),
  country CHARACTER VARYING(15),
  phone CHARACTER VARYING(24),
  fax CHARACTER VARYING(24),
  homepage TEXT
);

CREATE TABLE products (
  product_id SMALLINT NOT NULL PRIMARY KEY,
  product_name CHARACTER VARYING(40) NOT NULL,
  supplier_id SMALLINT REFERENCES suppliers,
  category_id SMALLINT REFERENCES categories,
  quantity_per_unit CHARACTER VARYING(20),
  unit_price REAL,
  units_in_stock SMALLINT,
  units_on_order SMALLINT,
  reorder_level SMALLINT,
  discontinued integer NOT NULL
);

CREATE TABLE region (
  region_id SMALLINT NOT NULL PRIMARY KEY,
  region_description BPCHAR NOT NULL
);

CREATE TABLE shippers (
  shipper_id SMALLINT NOT NULL PRIMARY KEY,
  company_name CHARACTER VARYING(40) NOT NULL,
  phone CHARACTER VARYING(24)
);

CREATE TABLE orders (
  order_id SMALLINT NOT NULL PRIMARY KEY,
  customer_id BPCHAR REFERENCES customers,
  employee_id SMALLINT REFERENCES employees,
  order_date DATE,
  required_date DATE,
  shipped_date DATE,
  ship_via SMALLINT REFERENCES shippers,
  freight REAL,
  ship_name CHARACTER VARYING(40),
  ship_address CHARACTER VARYING(60),
  ship_city CHARACTER VARYING(15),
  ship_region CHARACTER VARYING(15),
  ship_postal_code CHARACTER VARYING(10),
  ship_country CHARACTER VARYING(15)
);

CREATE TABLE territories (
  territory_id CHARACTER VARYING(20) NOT NULL PRIMARY KEY,
  territory_description BPCHAR NOT NULL,
  region_id SMALLINT NOT NULL REFERENCES region
);

CREATE TABLE employee_territories (
  employee_id SMALLINT NOT NULL REFERENCES employees,
  territory_id CHARACTER VARYING(20) NOT NULL REFERENCES territories,
  PRIMARY KEY (employee_id, territory_id)
);

CREATE TABLE order_details (
  order_id SMALLINT NOT NULL REFERENCES orders,
  product_id SMALLINT NOT NULL REFERENCES products,
  unit_price REAL NOT NULL,
  quantity SMALLINT NOT NULL,
  discount REAL NOT NULL,
  PRIMARY KEY (order_id, product_id)
);

CREATE TABLE us_states (
  state_id SMALLINT NOT NULL PRIMARY KEY,
  state_name CHARACTER VARYING(100),
  state_abbr CHARACTER VARYING(2),
  state_region CHARACTER VARYING(50)
);
