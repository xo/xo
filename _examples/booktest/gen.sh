#!/bin/bash

SRC=$(realpath $(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd))

declare -A DSNS
TEST=$(basename $SRC)
DSNS+=(
  [mysql]=my://$TEST:$TEST@localhost/$TEST
  [oracle]=or://$TEST:$TEST@localhost/db1
  [postgres]=pg://$TEST:$TEST@localhost/$TEST
  [sqlite3]=sq:$TEST.db
  [sqlserver]=ms://$TEST:$TEST@localhost/$TEST
)

APPLY=0
DATABASES="mysql oracle postgres sqlite3 sqlserver"

OPTIND=1
while getopts "ad:" opt; do
case "$opt" in
  a) APPLY=1 ;;
  d) DATABASES=$OPTARG ;;
esac
done

XOBIN=$(which xo)
if [ -e $SRC/../../xo ]; then
  XOBIN=$SRC/../../xo
fi
XOBIN=$(realpath $XOBIN)

pushd $SRC &> /dev/null

rm *.db

for TYPE in $DATABASES; do
  DB=${DSNS[$TYPE]}
  mkdir -p $TYPE
  rm -f $TYPE/*.xo.go
  echo "------------------------------------------------------"
  echo "$TYPE: $DB"
  echo ""
  if [ "$APPLY" = "1" ]; then
    (set -ex;
      $SRC/../createdb.sh -d $TYPE -n $TEST
      usql -f sql/${TYPE}_schema.sql $DB
    )
    if [ -f sql/${TYPE}_data.sql ]; then
      (set -ex;
        usql -f sql/${TYPE}_data.sql $DB
      )
    fi
  fi
  (set -ex;
    $XOBIN schema $DB -o $TYPE
    $XOBIN query $DB < sql/${TYPE}_query.sql \
      -o $TYPE \
      -M \
      -B \
      -2 \
      -T AuthorBookResult \
      --type-comment='AuthorBookResult is the result of a search.'
    go build ./$TYPE
    go build
    ./$TEST -v -dsn $DB
    usql -c 'select * from books;' $DB
  )
done

popd &> /dev/null
