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

  MODELS=$i.xo.go
  OUT=pokedex-$i

  rm -f $OUT $MODELS

  echo -e "------------------------------------------------------\n$i='$DB'"

  echo -e "\nxo $DB --single-file -o $MODELS -p main -tags $i -x"
  $XOBIN $EXTRA $DB --single-file -o $MODELS -p main -tags $i -x

  echo -e "\ngo build -o $OUT -tags $i"
  go build -o $OUT -tags $i

  echo -e "\n./$OUT -url $DB"
  ./$OUT $EXTRA -url $DB

  echo ""
done

popd &> /dev/null
