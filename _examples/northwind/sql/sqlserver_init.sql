EXEC sp_configure
  'contained database authentication', 1;

RECONFIGURE;

DROP LOGIN northwind;

DROP DATABASE northwind;

CREATE DATABASE northwind
  CONTAINMENT=PARTIAL;

\connect ms://localhost/northwind

CREATE LOGIN northwind
  WITH
    PASSWORD='northwind',
    CHECK_POLICY=OFF,
    DEFAULT_DATABASE=northwind;

CREATE USER northwind
  FOR LOGIN northwind
  WITH DEFAULT_SCHEMA=northwind;

CREATE SCHEMA northwind AUTHORIZATION northwind;

EXEC sp_addrolemember
  'db_owner', 'northwind';
