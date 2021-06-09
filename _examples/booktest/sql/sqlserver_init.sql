EXEC sp_configure
  'contained database authentication', 1;

RECONFIGURE;

DROP LOGIN booktest;

DROP DATABASE booktest;

CREATE DATABASE booktest
  CONTAINMENT=PARTIAL;

\connect ms://localhost/booktest

CREATE LOGIN booktest
  WITH
    PASSWORD='booktest',
    CHECK_POLICY=OFF,
    DEFAULT_DATABASE=booktest;

CREATE USER booktest
  FOR LOGIN booktest
  WITH DEFAULT_SCHEMA=booktest;

CREATE SCHEMA booktest AUTHORIZATION booktest;

EXEC sp_addrolemember
  'db_owner', 'booktest';
