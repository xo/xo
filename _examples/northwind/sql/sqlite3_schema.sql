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
  category_name VARCHAR(15) NOT NULL,
  description TEXT,
  picture BLOB
);

CREATE TABLE customer_demographics (
  customer_type_id CHAR(255) NOT NULL PRIMARY KEY,
  customer_desc TEXT
);

CREATE TABLE customers (
  customer_id CHAR(255) NOT NULL PRIMARY KEY,
  company_name VARCHAR(40) NOT NULL,
  contact_name VARCHAR(30),
  contact_title VARCHAR(30),
  address VARCHAR(60),
  city VARCHAR(15),
  region VARCHAR(15),
  postal_code VARCHAR(10),
  country VARCHAR(15),
  phone VARCHAR(24),
  fax VARCHAR(24)
);

CREATE TABLE customer_customer_demo (
  customer_id CHAR(255) NOT NULL REFERENCES customers(customer_id),
  customer_type_id CHAR(255) NOT NULL REFERENCES customer_demographics(customer_type_id),
  PRIMARY KEY (customer_id, customer_type_id)
);

CREATE TABLE employees (
  employee_id SMALLINT NOT NULL PRIMARY KEY,
  last_name VARCHAR(20) NOT NULL,
  first_name VARCHAR(10) NOT NULL,
  title VARCHAR(30),
  title_of_courtesy VARCHAR(25),
  birth_date DATE,
  hire_date DATE,
  address VARCHAR(60),
  city VARCHAR(15),
  region VARCHAR(15),
  postal_code VARCHAR(10),
  country VARCHAR(15),
  home_phone VARCHAR(24),
  extension VARCHAR(4),
  photo BLOB,
  notes TEXT,
  reports_to SMALLINT REFERENCES employees(employee_id),
  photo_path VARCHAR(255)
);

CREATE TABLE suppliers (
  supplier_id SMALLINT NOT NULL PRIMARY KEY,
  company_name VARCHAR(40) NOT NULL,
  contact_name VARCHAR(30),
  contact_title VARCHAR(30),
  address VARCHAR(60),
  city VARCHAR(15),
  region VARCHAR(15),
  postal_code VARCHAR(10),
  country VARCHAR(15),
  phone VARCHAR(24),
  fax VARCHAR(24),
  homepage TEXT
);

CREATE TABLE products (
  product_id SMALLINT NOT NULL PRIMARY KEY,
  product_name VARCHAR(40) NOT NULL,
  supplier_id SMALLINT REFERENCES suppliers(supplier_id),
  category_id SMALLINT REFERENCES categories(category_id),
  quantity_per_unit VARCHAR(20),
  unit_price REAL,
  units_in_stock SMALLINT,
  units_on_order SMALLINT,
  reorder_level SMALLINT,
  discontinued integer NOT NULL
);

CREATE TABLE region (
  region_id SMALLINT NOT NULL PRIMARY KEY,
  region_description CHAR(255) NOT NULL
);

CREATE TABLE shippers (
  shipper_id SMALLINT NOT NULL PRIMARY KEY,
  company_name VARCHAR(40) NOT NULL,
  phone VARCHAR(24)
);

CREATE TABLE orders (
  order_id SMALLINT NOT NULL PRIMARY KEY,
  customer_id CHAR(255) REFERENCES customers(customer_id),
  employee_id SMALLINT REFERENCES employees(employee_id),
  order_date DATE,
  required_date DATE,
  shipped_date DATE,
  ship_via SMALLINT REFERENCES shippers(shipper_id),
  freight REAL,
  ship_name VARCHAR(40),
  ship_address VARCHAR(60),
  ship_city VARCHAR(15),
  ship_region VARCHAR(15),
  ship_postal_code VARCHAR(10),
  ship_country VARCHAR(15)
);

CREATE TABLE territories (
  territory_id VARCHAR(20) NOT NULL PRIMARY KEY,
  territory_description CHAR(255) NOT NULL,
  region_id SMALLINT NOT NULL REFERENCES region(region_id)
);

CREATE TABLE employee_territories (
  employee_id SMALLINT NOT NULL REFERENCES employees(employee_id),
  territory_id VARCHAR(20) NOT NULL REFERENCES territories(territory_id),
  PRIMARY KEY (employee_id, territory_id)
);

CREATE TABLE order_details (
  order_id SMALLINT NOT NULL REFERENCES orders(order_id),
  product_id SMALLINT NOT NULL REFERENCES products(product_id),
  unit_price REAL NOT NULL,
  quantity SMALLINT NOT NULL,
  discount REAL NOT NULL,
  PRIMARY KEY (order_id, product_id)
);

CREATE TABLE us_states (
  state_id SMALLINT NOT NULL PRIMARY KEY,
  state_name VARCHAR(100),
  state_abbr VARCHAR(2),
  state_region VARCHAR(50)
);
