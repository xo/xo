#!/bin/bash

SRC=$(realpath $(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd))

declare -A DSNS
DSNS+=(
  [mysql]=my://a_bit_of_everything:a_bit_of_everything@localhost/a_bit_of_everything
  [oracle]=or://a_bit_of_everything:a_bit_of_everything@localhost/db1
  [postgres]=pg://a_bit_of_everything:a_bit_of_everything@localhost/a_bit_of_everything
  [sqlite3]=sq:a_bit_of_everything.db
  [sqlserver]=ms://a_bit_of_everything:a_bit_of_everything@localhost/a_bit_of_everything
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
      $SRC/../createdb.sh -d $TYPE -n a_bit_of_everything
      usql -f sql/${TYPE}_schema.sql $DB
    )
  fi
  (set -ex;
    $XOBIN schema $DB -o $TYPE
    go build ./$TYPE
    go build
    ./a_bit_of_everything -v -dsn $DB
  )
done

popd &> /dev/null
