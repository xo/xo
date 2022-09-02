#!/bin/bash

SRC=$(realpath $(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd))

DB=postgres://postgres:P4ssw0rd@localhost/

BUILD=0
while getopts "b" opt; do
case "$opt" in
  b) BUILD=1 && shift ;;
esac
done

if [ "$BUILD" = "1" ]; then
  pushd $SRC/../../ &> /dev/null
  (set -x;
    go build
  )
  popd &> /dev/null
fi

XOBIN=$(which xo)
if [ -e $SRC/../../xo ]; then
  XOBIN=$SRC/../../xo
fi
XOBIN=$(realpath $XOBIN)

pushd $SRC &> /dev/null
mkdir -p pgcatalog ischema
rm -f pgcatalog/*.xo.go ischema/*.xo.go
FLAGS=(
  --postgres-oids
  --go-custom=pgtypes
  --go-import github.com/xo/xo/_examples/pgcatalog/pgtypes
  --go-uuid github.com/google/uuid
)
(set -ex;
  $XOBIN schema ${FLAGS[@]} -o pgcatalog         -s pg_catalog         -S pgcatalog.xo.go   $DB $@
  $XOBIN schema ${FLAGS[@]} -o pgcatalog -t yaml -s pg_catalog                              $DB $@
  $XOBIN schema ${FLAGS[@]} -o pgcatalog -t json -s pg_catalog                              $DB $@
  $XOBIN schema ${FLAGS[@]} -o ischema           -s information_schema                      $DB $@
  $XOBIN schema ${FLAGS[@]} -o ischema   -t yaml -s information_schema                      $DB $@
  $XOBIN schema ${FLAGS[@]} -o ischema   -t json -s information_schema                      $DB $@
  go build ./pgcatalog/
  go build ./ischema/
)
popd &> /dev/null
