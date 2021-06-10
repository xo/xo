CREATE TABLE categories (
  category_id SMALLINT NOT NULL CONSTRAINT categories_pkey PRIMARY KEY,
  category_name NVARCHAR2(15) NOT NULL,
  description CLOB,
  picture BLOB
);

CREATE TABLE customer_demographics (
  customer_type_id NCHAR(255) NOT NULL CONSTRAINT customer_demographics_pkey PRIMARY KEY,
  customer_desc CLOB
);

CREATE TABLE customers (
  customer_id NCHAR(255) NOT NULL CONSTRAINT customers_pkey PRIMARY KEY,
  company_name NVARCHAR2(40) NOT NULL,
  contact_name NVARCHAR2(30),
  contact_title NVARCHAR2(30),
  address NVARCHAR2(60),
  city NVARCHAR2(15),
  region NVARCHAR2(15),
  postal_code NVARCHAR2(10),
  country NVARCHAR2(15),
  phone NVARCHAR2(24),
  fax NVARCHAR2(24)
);

CREATE TABLE customer_customer_demo (
  customer_id NCHAR(255) NOT NULL CONSTRAINT customer_customer_demo_customer_id_fkey REFERENCES customers(customer_id),
  customer_type_id NCHAR(255) NOT NULL CONSTRAINT customer_customer_demo_customer_type_id_fkey REFERENCES customer_demographics(customer_type_id),
  CONSTRAINT customer_customer_demo_pkey PRIMARY KEY (customer_id, customer_type_id)
);

CREATE TABLE employees (
  employee_id SMALLINT NOT NULL CONSTRAINT employees_pkey PRIMARY KEY,
  last_name NVARCHAR2(20) NOT NULL,
  first_name NVARCHAR2(10) NOT NULL,
  title NVARCHAR2(30),
  title_of_courtesy NVARCHAR2(25),
  birth_date DATE,
  hire_date DATE,
  address NVARCHAR2(60),
  city NVARCHAR2(15),
  region NVARCHAR2(15),
  postal_code NVARCHAR2(10),
  country NVARCHAR2(15),
  home_phone NVARCHAR2(24),
  extension NVARCHAR2(4),
  photo BLOB,
  notes CLOB,
  reports_to SMALLINT CONSTRAINT employees_reports_to_fkey REFERENCES employees(employee_id),
  photo_path NVARCHAR2(255)
);

CREATE TABLE suppliers (
  supplier_id SMALLINT NOT NULL CONSTRAINT suppliers_pkey PRIMARY KEY,
  company_name NVARCHAR2(40) NOT NULL,
  contact_name NVARCHAR2(30),
  contact_title NVARCHAR2(30),
  address NVARCHAR2(60),
  city NVARCHAR2(15),
  region NVARCHAR2(15),
  postal_code NVARCHAR2(10),
  country NVARCHAR2(15),
  phone NVARCHAR2(24),
  fax NVARCHAR2(24),
  homepage CLOB
);

CREATE TABLE products (
  product_id SMALLINT NOT NULL CONSTRAINT products_pkey PRIMARY KEY,
  product_name NVARCHAR2(40) NOT NULL,
  supplier_id SMALLINT CONSTRAINT products_suplier_id_fkey REFERENCES suppliers(supplier_id),
  category_id SMALLINT CONSTRAINT products_category_id_fkey REFERENCES categories(category_id),
  quantity_per_unit NVARCHAR2(20),
  unit_price REAL,
  units_in_stock SMALLINT,
  units_on_order SMALLINT,
  reorder_level SMALLINT,
  discontinued integer NOT NULL
);

CREATE TABLE region (
  region_id SMALLINT NOT NULL CONSTRAINT regions_pkey PRIMARY KEY,
  region_description NCHAR(255) NOT NULL
);

CREATE TABLE shippers (
  shipper_id SMALLINT NOT NULL CONSTRAINT shippers_pkey PRIMARY KEY,
  company_name NVARCHAR2(40) NOT NULL,
  phone NVARCHAR2(24)
);

CREATE TABLE orders (
  order_id SMALLINT NOT NULL CONSTRAINT orders_pkey PRIMARY KEY,
  customer_id NCHAR(255) CONSTRAINT orders_customer_id_fkey REFERENCES customers(customer_id),
  employee_id SMALLINT CONSTRAINT orders_employee_id_fkey REFERENCES employees(employee_id),
  order_date DATE,
  required_date DATE,
  shipped_date DATE,
  ship_via SMALLINT CONSTRAINT orders_ship_via_fkey REFERENCES shippers(shipper_id),
  freight REAL,
  ship_name NVARCHAR2(40),
  ship_address NVARCHAR2(60),
  ship_city NVARCHAR2(15),
  ship_region NVARCHAR2(15),
  ship_postal_code NVARCHAR2(10),
  ship_country NVARCHAR2(15)
);

CREATE TABLE territories (
  territory_id NVARCHAR2(20) NOT NULL CONSTRAINT territories_pkey PRIMARY KEY,
  territory_description NCHAR(255) NOT NULL,
  region_id SMALLINT NOT NULL CONSTRAINT territories_region_id_fkey REFERENCES region(region_id)
);

CREATE TABLE employee_territories (
  employee_id SMALLINT NOT NULL CONSTRAINT employee_territories_employee_id_fkey REFERENCES employees(employee_id),
  territory_id NVARCHAR2(20) NOT NULL CONSTRAINT employee_territories_territory_id_fkey REFERENCES territories(territory_id),
  CONSTRAINT employee_territories_pkey PRIMARY KEY (employee_id, territory_id)
);

CREATE TABLE order_details (
  order_id SMALLINT NOT NULL CONSTRAINT order_details_order_id_fkey REFERENCES orders(order_id),
  product_id SMALLINT NOT NULL CONSTRAINT order_details_product_id_fkey REFERENCES products(product_id),
  unit_price REAL NOT NULL,
  quantity SMALLINT NOT NULL,
  discount REAL NOT NULL,
  CONSTRAINT order_details_pkey PRIMARY KEY (order_id, product_id)
);

CREATE TABLE us_states (
  state_id SMALLINT NOT NULL CONSTRAINT us_states_pkey PRIMARY KEY,
  state_name NVARCHAR2(100),
  state_abbr NVARCHAR2(2),
  state_region NVARCHAR2(50)
);
