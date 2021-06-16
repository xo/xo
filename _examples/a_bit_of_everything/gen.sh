#!/bin/bash

SRC=$(realpath $(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd))

TEST=$(basename $SRC)

declare -A DSNS
DSNS+=(
  [mysql]=my://$TEST:$TEST@localhost/$TEST
  [oracle]=or://$TEST:$TEST@localhost/db1
  [postgres]=pg://$TEST:$TEST@localhost/$TEST
  [sqlite3]=sq:$TEST.db
  [sqlserver]=ms://$TEST:$TEST@localhost/$TEST
)

APPLY=0
DATABASES="mysql oracle postgres sqlite3 sqlserver"
ARGS=()

OPTIND=1
while getopts "ad:v" opt; do
case "$opt" in
  a) APPLY=1 ;;
  d) DATABASES=$OPTARG ;;
  v) ARGS+=(-v)
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
  if [ "$APPLY" = "1" ]; then
    if [[ "$TYPE" = "sqlite3" && -f $TEST.db ]]; then
      (set -ex;
        rm $TEST.db
      )
    fi
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
    $XOBIN schema $DB -o $TYPE ${ARGS[@]}
    go build ./$TYPE
    go build
    ./$TEST -dsn $DB ${ARGS[@]}
  )
done

popd &> /dev/null
