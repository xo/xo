#!/bin/bash

SRC=$(realpath $(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd))

DB=pg://

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
  --go-import github.com/google/uuid
)
(set -ex;
  $XOBIN schema ${FLAGS[@]} -o pgcatalog -s pg_catalog         -S pgcatalog.xo.go $DB $@
  $XOBIN schema ${FLAGS[@]} -o ischema   -s information_schema                    $DB $@
  go build ./pgcatalog/
  go build ./ischema/
)
popd &> /dev/null
