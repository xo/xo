#!/bin/bash

EXTRA=$1

SRC=$(realpath $(cd -P "$( dirname "${BASH_SOURCE[0]}" )" && pwd ))

XOBIN=$(which xo)
if [ -e $SRC/../../xo ]; then
  XOBIN=$SRC/../../xo
fi

set -x

mkdir -p postgres mysql oracle sqlite3
rm -f postgres/*.go mysql/*.go oracle/*.go sqlite3/*.go

xo -o postgres postgres://django:django@localhost/django
xo -o mysql mysql://django:django@localhost/django
xo -o oracle oracle://django:django@$(docker port orcl 1521)/orcl
xo -o sqlite3 file:$SRC/django.sqlite3

pushd $SRC &> /dev/null

go build ./postgres/
go build ./mysql/
go build ./oracle/
go build ./sqlite3/

popd &> /dev/null
