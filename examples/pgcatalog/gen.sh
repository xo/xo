#!/bin/bash

EXTRA=$1

SRC=$(realpath $(cd -P "$( dirname "${BASH_SOURCE[0]}" )" && pwd ))

XOBIN=$(which xo)
if [ -e $SRC/../../xo ]; then
  XOBIN=$SRC/../../xo
fi

set -ex

pushd $SRC &> /dev/null

mkdir -p pgcatalog ischema
rm -f pgcatalog/*.xo.go ischema/*.xo.go

$XOBIN $EXTRA pgsql://xodb:xodb@localhost/xodb -C pgtypes -o pgcatalog -s pg_catalog --single-file --enable-postgres-oids
$XOBIN $EXTRA pgsql://xodb:xodb@localhost/xodb -C pgtypes -o ischema -s information_schema --enable-postgres-oids

go build ./pgcatalog/
go build ./ischema/

popd &> /dev/null
