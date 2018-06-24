#!/bin/bash

# get/update all dependencies

go get -tags oracle -u \
  github.com/denisenkom/go-mssqldb \
  github.com/go-sql-driver/mysql \
  gopkg.in/goracle.v2 \
  github.com/lib/pq \
  github.com/mattn/go-sqlite3
