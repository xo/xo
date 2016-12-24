#!/bin/bash

EXTRA=$1

SRC=$(realpath $(cd -P "$( dirname "${BASH_SOURCE[0]}" )" && pwd ))

XOBIN=$(which xo)
if [ -e $SRC/../../xo ]; then
  XOBIN=$SRC/../../xo
fi

set -ex

pushd $SRC &> /dev/null

mkdir -p postgres mysql sqlite3 oracle
rm -f postgres/*.xo.go mysql/*.xo.go sqlite3/*.xo.go oracle/*.xo.go

$XOBIN $EXTRA -o postgres postgres://django:django@localhost/django
$XOBIN $EXTRA -o mysql mysql://django:django@localhost/django
$XOBIN $EXTRA -o sqlite3 file:$SRC/django.sqlite3
#$XOBIN $EXTRA -o oracle oracle://django:django@$(docker port orcl 1521)/xe.oracle.docker

go build ./postgres/
go build ./mysql/
go build ./sqlite3/
go build ./oracle/

popd &> /dev/null
