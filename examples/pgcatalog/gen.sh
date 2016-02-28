#!/bin/bash

set -ex

mkdir -p pgcatalog ischema
rm -f pgcatalog/*.xo.go ischema/*.xo.go

xo pgsql://xodb:xodb@localhost/xodb -C pgtypes -o pgcatalog -s pg_catalog --single-file --enable-postgres-oids
xo pgsql://xodb:xodb@localhost/xodb -C pgtypes -o ischema -s information_schema

go build ./pgcatalog/
go build ./ischema/
