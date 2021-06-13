#!/bin/bash

SRC=$(realpath $(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd))

declare -A DSNS
DSNS+=(
  [mysql]=my://django:django@localhost/django
  [oracle]=or://django:django@localhost/db1
  [postgres]=pg://django:django@localhost/django
  [sqlite3]=sq:django.db
  [sqlserver]=ms://django:django@localhost/django
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
      $SRC/../createdb.sh -d $TYPE -n django
      usql -f sql/${TYPE}_schema.sql $DB
    )
  fi
  (set -ex;
    $XOBIN schema $DB -o $TYPE
    go build ./$TYPE
    go build
    ./django -v -dsn $DB
  )
done

popd &> /dev/null
