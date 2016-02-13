#!/bin/bash

set -ex

rm -rf p t
mkdir -p p t

go generate
go build

./xo pgsql://xodb:xodb@localhost/xodb -o p -s pg_catalog -C pgcatalog --single-file
./xo pgsql://xodb:xodb@localhost/xodb -o t -s information_schema -C pgcatalog

go build ./p/
go build ./t/
