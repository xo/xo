EXEC sp_configure
  'contained database authentication', 1;

RECONFIGURE;

DROP LOGIN booktest;

DROP DATABASE booktest;

CREATE DATABASE booktest
  containment=partial;

\connect ms://localhost/booktest

CREATE LOGIN booktest
  WITH
    password='booktest',
    check_policy=off,
    default_database=booktest;

CREATE USER booktest
  FOR login booktest
  WITH default_schema=booktest;

CREATE SCHEMA booktest authorization booktest;

EXEC sp_addrolemember
  'db_owner', 'booktest';
