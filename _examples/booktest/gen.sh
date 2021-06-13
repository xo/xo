#!/bin/bash

SRC=$(realpath $(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd))

declare -A DSNS
DSNS+=(
  [mysql]=my://booktest:booktest@localhost/booktest
  [oracle]=or://booktest:booktest@localhost/db1
  [postgres]=pg://booktest:booktest@localhost/booktest
  [sqlite3]=sq:booktest.db
  [sqlserver]=ms://booktest:booktest@localhost/booktest
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

for TYPE in $DATABASES; do
  DB=${DSNS[$TYPE]}
  mkdir -p $TYPE
  rm -f $TYPE/*.xo.go
  echo "------------------------------------------------------"
  echo "$TYPE: $DB"
  echo ""
  if [ "$APPLY" = "1" ]; then
    (set -ex;
      $SRC/../createdb.sh -d $TYPE -n booktest
      usql -f sql/${TYPE}_schema.sql $DB
    )
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
    ./booktest -v -dsn $DB
    usql -c 'select * from books;' $DB
  )
done

popd &> /dev/null
