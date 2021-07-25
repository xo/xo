-- SQL Schema template for the northwind schema.
-- Generated on Sun Jul 25 07:10:21 WIB 2021 by xo.

CREATE TABLE categories (
    category_id NUMBER NOT NULL,
    category_name NVARCHAR2 NOT NULL,
    description CLOB,
    picture BLOB,
    CONSTRAINT categories_pkey UNIQUE (category_id)
);


CREATE TABLE customer_customer_demo (
    customer_id NCHAR NOT NULL CONSTRAINT customer_customer_demo_customer_id_fkey REFERENCES customers (customer_id),
    customer_type_id NCHAR NOT NULL CONSTRAINT customer_customer_demo_customer_type_id_fkey REFERENCES customer_demographics (customer_type_id),
    CONSTRAINT customer_customer_demo_pkey UNIQUE (customer_id, customer_type_id)
);


CREATE TABLE customer_demographics (
    customer_type_id NCHAR NOT NULL,
    customer_desc CLOB,
    CONSTRAINT customer_demographics_pkey UNIQUE (customer_type_id)
);


CREATE TABLE customers (
    customer_id NCHAR NOT NULL,
    company_name NVARCHAR2 NOT NULL,
    contact_name NVARCHAR2,
    contact_title NVARCHAR2,
    address NVARCHAR2,
    city NVARCHAR2,
    region NVARCHAR2,
    postal_code NVARCHAR2,
    country NVARCHAR2,
    phone NVARCHAR2,
    fax NVARCHAR2,
    CONSTRAINT customers_pkey UNIQUE (customer_id)
);


CREATE TABLE employee_territories (
    employee_id NUMBER NOT NULL CONSTRAINT employee_territories_employee_id_fkey REFERENCES employees (employee_id),
    territory_id NVARCHAR2 NOT NULL CONSTRAINT employee_territories_territory_id_fkey REFERENCES territories (territory_id),
    CONSTRAINT employee_territories_pkey UNIQUE (employee_id, territory_id)
);


CREATE TABLE employees (
    employee_id NUMBER NOT NULL,
    last_name NVARCHAR2 NOT NULL,
    first_name NVARCHAR2 NOT NULL,
    title NVARCHAR2,
    title_of_courtesy NVARCHAR2,
    birth_date DATE,
    hire_date DATE,
    address NVARCHAR2,
    city NVARCHAR2,
    region NVARCHAR2,
    postal_code NVARCHAR2,
    country NVARCHAR2,
    home_phone NVARCHAR2,
    extension NVARCHAR2,
    photo BLOB,
    notes CLOB,
    reports_to NUMBER CONSTRAINT employees_reports_to_fkey REFERENCES employees (reports_to),
    photo_path NVARCHAR2,
    CONSTRAINT employees_pkey UNIQUE (employee_id)
);


CREATE TABLE order_details (
    order_id NUMBER NOT NULL CONSTRAINT order_details_order_id_fkey REFERENCES orders (order_id),
    product_id NUMBER NOT NULL CONSTRAINT order_details_product_id_fkey REFERENCES products (product_id),
    unit_price FLOAT NOT NULL,
    quantity NUMBER NOT NULL,
    discount FLOAT NOT NULL,
    CONSTRAINT order_details_pkey UNIQUE (order_id, product_id)
);


CREATE TABLE orders (
    order_id NUMBER NOT NULL,
    customer_id NCHAR CONSTRAINT orders_customer_id_fkey REFERENCES customers (customer_id),
    employee_id NUMBER CONSTRAINT orders_employee_id_fkey REFERENCES employees (employee_id),
    order_date DATE,
    required_date DATE,
    shipped_date DATE,
    ship_via NUMBER CONSTRAINT orders_ship_via_fkey REFERENCES shippers (ship_via),
    freight FLOAT,
    ship_name NVARCHAR2,
    ship_address NVARCHAR2,
    ship_city NVARCHAR2,
    ship_region NVARCHAR2,
    ship_postal_code NVARCHAR2,
    ship_country NVARCHAR2,
    CONSTRAINT orders_pkey UNIQUE (order_id)
);


CREATE TABLE products (
    product_id NUMBER NOT NULL,
    product_name NVARCHAR2 NOT NULL,
    supplier_id NUMBER CONSTRAINT products_suplier_id_fkey REFERENCES suppliers (supplier_id),
    category_id NUMBER CONSTRAINT products_category_id_fkey REFERENCES categories (category_id),
    quantity_per_unit NVARCHAR2,
    unit_price FLOAT,
    units_in_stock NUMBER,
    units_on_order NUMBER,
    reorder_level NUMBER,
    discontinued NUMBER NOT NULL,
    CONSTRAINT products_pkey UNIQUE (product_id)
);


CREATE TABLE region (
    region_id NUMBER NOT NULL,
    region_description NCHAR NOT NULL,
    CONSTRAINT regions_pkey UNIQUE (region_id)
);


CREATE TABLE shippers (
    shipper_id NUMBER NOT NULL,
    company_name NVARCHAR2 NOT NULL,
    phone NVARCHAR2,
    CONSTRAINT shippers_pkey UNIQUE (shipper_id)
);


CREATE TABLE suppliers (
    supplier_id NUMBER NOT NULL,
    company_name NVARCHAR2 NOT NULL,
    contact_name NVARCHAR2,
    contact_title NVARCHAR2,
    address NVARCHAR2,
    city NVARCHAR2,
    region NVARCHAR2,
    postal_code NVARCHAR2,
    country NVARCHAR2,
    phone NVARCHAR2,
    fax NVARCHAR2,
    homepage CLOB,
    CONSTRAINT suppliers_pkey UNIQUE (supplier_id)
);


CREATE TABLE territories (
    territory_id NVARCHAR2 NOT NULL,
    territory_description NCHAR NOT NULL,
    region_id NUMBER NOT NULL CONSTRAINT territories_region_id_fkey REFERENCES region (region_id),
    CONSTRAINT territories_pkey UNIQUE (territory_id)
);


CREATE TABLE us_states (
    state_id NUMBER NOT NULL,
    state_name NVARCHAR2,
    state_abbr NVARCHAR2,
    state_region NVARCHAR2,
    CONSTRAINT us_states_pkey UNIQUE (state_id)
);


