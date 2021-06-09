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
  [mysql]=my://localhost/northwind
  [oracle]=or://northwind:northwind@localhost/db1
  [postgres]=pg://localhost/northwind
  [sqlite3]=sq:northwind.db
  [sqlserver]=ms://northwind:northwind@localhost/northwind
)

APPLY=0
DATABASES="mysql oracle postgres sqlite3 sqlserver"

OPTIND=1
while getopts "ad:" opt; do
case "$opt" in
  a) APPLY=1 ;;
  d) DATABASES=$OPTARG
esac
done

XOBIN=$(which xo)
if [ -e $SRC/../../xo ]; then
  XOBIN=$SRC/../../xo
fi
XOBIN=$(realpath $XOBIN)

pushd $SRC &> /dev/null

rm -f *.db

for NAME in $DATABASES; do
  INITDB=${INIT[$NAME]}
  TESTDB=${TEST[$NAME]}
  mkdir -p $NAME
  rm -f $NAME/*.xo.go
  echo "------------------------------------------------------"
  echo "$NAME"
  echo "  init: $INITDB"
  echo "  test: $TESTDB"
  echo ""
  if [ "$APPLY" = "1" ]; then
    if [ -f sql/${NAME}_init.sql ]; then
      (set -ex;
        usql -f sql/${NAME}_init.sql $INITDB
      )
    fi
    (set -ex;
      usql -f sql/${NAME}_schema.sql $TESTDB
    )
    if [ -f sql/${NAME}_data.sql ]; then
      (set -ex;
        usql -f sql/${NAME}_data.sql $TESTDB
      )
    fi
  fi
  (set -ex;
    $XOBIN schema $TESTDB -o $NAME
    go build ./$NAME
    go build
    ./northwind -v -dsn $TESTDB
  )
done

popd &> /dev/null
