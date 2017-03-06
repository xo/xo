#!/bin/bash

set -e

EXTRA=$1

SRC=$(realpath $(cd -P "$( dirname "${BASH_SOURCE[0]}" )" && pwd ))

XOBIN=$(which xo)
if [ -e $SRC/../../xo ]; then
  XOBIN=$SRC/../../xo
fi

XO_ORACLE=$($XOBIN --has-oracle-support)
USQL_ORACLE=$(usql --has-oracle-support)

pushd $SRC &> /dev/null

for i in */config; do
  i=$(dirname $i)

  # skip
  if [ -f $i/skip ]; then
    continue
  fi

  # skip oracle if no oracle support
  if [[ $i == "oracle" && ( "$XO_ORACLE" != "1" || "$USQL_ORACLE" != "1" ) ]]; then
    continue
  fi

  source $i/config

  MODELS=$i/models
  OUT=booktest-$i

  mkdir -p $MODELS
  rm -f $OUT $MODELS/*.xo.go

  echo -e "------------------------------------------------------\n$i='$DB'"

  if [ -f $i/pre ]; then
    echo -e "\nsourcing $i/pre"
    source $i/pre
  fi

  echo -e "\nusql $DB -f $i/schema.sql"
  usql $EXTRA "$DB" -f $i/schema.sql

  echo -e "\nxo $DB -o $MODELS"
  $XOBIN $EXTRA "$DB" -o $MODELS

  echo -e "\nxo $DB -o $MODELS < $i/custom-query.xo.sql"
  $XOBIN $EXTRA \
    -o $MODELS \
    -N -M -B -T AuthorBookResult \
    --query-type-comment='AuthorBookResult is the result of a search.' \
    "$DB" < $i/custom-query.xo.sql

  echo -e "\ngo build -o $OUT ./$i/"
  go build -o $OUT ./$i/

  echo -e "\n./$OUT"
  ./$OUT $EXTRA -url "$DB"

  echo -e "\nselect * from books;"
  usql $EXTRA "$DB" -c 'select * from books;'

  echo ""
done

popd &> /dev/null
