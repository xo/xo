#!/bin/bash

SRC=$(realpath $(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd))

DB=pg://localhost

XOBIN=$(which xo)
if [ -e $SRC/../../xo ]; then
  XOBIN=$SRC/../../xo
fi
XOBIN=$(realpath $XOBIN)

pushd $SRC &> /dev/null
mkdir -p models
rm -f models/*.xo.go
(set -x;
  usql $DB -c 'drop database northwind;'
)
(set -ex;
  usql $DB -c 'create database northwind;'
  usql $DB/northwind -f northwind_ddl.sql
  usql $DB/northwind -f northwind_data.sql
  $XOBIN schema -o models $DB/northwind $@
  go build ./models/
)
popd &> /dev/null
