#!/bin/bash

SRC=$(realpath $(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd))

declare -A INIT
INIT+=(
  [mysql]=my://localhost
  [oracle]=or://localhost/db1
  [postgres]=pg://localhost
  [sqlserver]=ms://localhost
)
declare -A TEST
TEST+=(
  [mysql]=my://localhost/booktest
  [oracle]=or://booktest:booktest@localhost/db1
  [postgres]=pg://localhost/booktest
  [sqlite3]=sq:booktest.db
  [sqlserver]=ms://localhost/booktest
)

XOBIN=$(which xo)
if [ -e $SRC/../../xo ]; then
  XOBIN=$SRC/../../xo
fi
XOBIN=$(realpath $XOBIN)

pushd $SRC &> /dev/null

rm -f *.db

for NAME in mysql oracle postgres sqlite3 sqlserver; do
  INITDB=${INIT[$NAME]}
  TESTDB=${TEST[$NAME]}
  mkdir -p $NAME
  rm -f $NAME/*.xo.go
  echo "------------------------------------------------------"
  echo "$NAME"
  echo "  init: $INITDB"
  echo "  test: $TESTDB"
  echo ""
  if [ -f sql/${NAME}_init.sql ]; then
    (set -ex;
      usql -f sql/${NAME}_init.sql $INITDB
    )
  fi
  (set -ex;
    usql -f sql/${NAME}_schema.sql $TESTDB
    $XOBIN schema $TESTDB -o $NAME
    $XOBIN query $TESTDB < sql/${NAME}_query.sql \
      -o $NAME \
      -M \
      -B \
      -2 \
      -T AuthorBookResult \
      --type-comment='AuthorBookResult is the result of a search.'
    go build ./$NAME
    #go build
    #./booktest -dsn $TESTDB
    #usql -c 'select * from books;' $TESTDB
  )
done

popd &> /dev/null
