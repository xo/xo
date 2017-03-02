#!/bin/bash

EXTRA=$1

SRC=$(realpath $(cd -P "$( dirname "${BASH_SOURCE[0]}" )" && pwd ))

XOBIN=$(which xo)
if [ -e $SRC/../../xo ]; then
  XOBIN=$SRC/../../xo
fi

set -e

pushd $SRC &> /dev/null

for i in postgres mysql sqlite3 oracle mssql; do
  # skip if no config
  if [ ! -f $i/config ]; then
    continue
  fi

  # erase generated files
  rm -f $i/*.xo.go

  # load config
  . $i/config

  $XOBIN $EXTRA -o $i $DB

  go build ./$i/
done

popd &> /dev/null
