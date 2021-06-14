CREATE TABLE categories (
  category_id INTEGER NOT NULL PRIMARY KEY,
  category_name VARCHAR(15) NOT NULL,
  description TEXT,
  picture BYTEA
);

CREATE TABLE customer_demographics (
  customer_type_id BPCHAR NOT NULL PRIMARY KEY,
  customer_desc TEXT
);

CREATE TABLE customers (
  customer_id BPCHAR NOT NULL PRIMARY KEY,
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
  customer_id BPCHAR NOT NULL REFERENCES customers,
  customer_type_id BPCHAR NOT NULL REFERENCES customer_demographics,
  PRIMARY KEY (customer_id, customer_type_id)
);

CREATE TABLE employees (
  employee_id INTEGER NOT NULL PRIMARY KEY,
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
  photo BYTEA,
  notes TEXT,
  reports_to INTEGER REFERENCES employees,
  photo_path VARCHAR(255)
);

CREATE TABLE suppliers (
  supplier_id INTEGER NOT NULL PRIMARY KEY,
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
  product_id INTEGER NOT NULL PRIMARY KEY,
  product_name VARCHAR(40) NOT NULL,
  supplier_id INTEGER REFERENCES suppliers,
  category_id INTEGER REFERENCES categories,
  quantity_per_unit VARCHAR(20),
  unit_price REAL,
  units_in_stock INTEGER,
  units_on_order INTEGER,
  reorder_level INTEGER,
  discontinued integer NOT NULL
);

CREATE TABLE region (
  region_id INTEGER NOT NULL PRIMARY KEY,
  region_description BPCHAR NOT NULL
);

CREATE TABLE shippers (
  shipper_id INTEGER NOT NULL PRIMARY KEY,
  company_name VARCHAR(40) NOT NULL,
  phone VARCHAR(24)
);

CREATE TABLE orders (
  order_id INTEGER NOT NULL PRIMARY KEY,
  customer_id BPCHAR REFERENCES customers,
  employee_id INTEGER REFERENCES employees,
  order_date DATE,
  required_date DATE,
  shipped_date DATE,
  ship_via INTEGER REFERENCES shippers,
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
  territory_description BPCHAR NOT NULL,
  region_id INTEGER NOT NULL REFERENCES region
);

CREATE TABLE employee_territories (
  employee_id INTEGER NOT NULL REFERENCES employees,
  territory_id VARCHAR(20) NOT NULL REFERENCES territories,
  PRIMARY KEY (employee_id, territory_id)
);

CREATE TABLE order_details (
  order_id INTEGER NOT NULL REFERENCES orders,
  product_id INTEGER NOT NULL REFERENCES products,
  unit_price REAL NOT NULL,
  quantity INTEGER NOT NULL,
  discount REAL NOT NULL,
  PRIMARY KEY (order_id, product_id)
);

CREATE TABLE us_states (
  state_id INTEGER NOT NULL PRIMARY KEY,
  state_name VARCHAR(100),
  state_abbr VARCHAR(2),
  state_region VARCHAR(50)
);
