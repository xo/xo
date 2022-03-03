#!/bin/bash

SRC=$(realpath $(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd))

DATABASES="mysql oracle postgres sqlite3 sqlserver"
NAME=
PASS=
DATAFILE=

declare -A INIT
INIT+=(
  [mysql]=my://localhost/
  [oracle]=or://localhost:1521/db1
  [postgres]=pg://localhost/
  [sqlserver]=ms://localhost/
)

OPTIND=1
while getopts "d:f:n:" opt; do
case "$opt" in
  d) DATABASES=$OPTARG ;;
  f) DATAFILE=$OPTARG ;;
  n) NAME=$OPTARG ;;
  p) PASS=$OPTARG ;;
esac
done

if [ -z "$NAME" ]; then
  echo "usage: $0 [-d DATABASES] [-f DATAFILE] [-p PASS] -n <NAME>"
  exit 1
fi

if [ -z "$PASS" ]; then
  PASS="$NAME"
fi

pushd $SRC &> /dev/null
for TYPE in $DATABASES; do
  if [ "$TYPE" = "sqlite3" ]; then
    if [ -e "$NAME.db" ]; then
      (set -x;
        rm "$NAME.db"
      )
    fi
    continue
  fi
  DB=${INIT[$TYPE]}
  if [[ -z "$DATAFILE" && "$TYPE" = "oracle" ]]; then
    DATAFILE=$(printf '/opt/oracle/oradata/ORASID/%s.dbf' "$NAME")
  elif [[ "$TYPE" != "oracle" ]]; then
    DATAFILE=""
  fi
  (set -x;
    usql $DB \
      --set=DB="$DB" \
      --set=NAME="$NAME" \
      --set=PASS="$PASS" \
      --set=DATAFILE="$DATAFILE" \
      -c "\\echo 'DB:       ':'DB'" \
      -c "\\echo 'NAME:     ':'NAME'" \
      -c "\\echo 'PASS:     ':'PASS'" \
      -c "\\echo 'DATAFILE: ':'DATAFILE'" \
      -f $SRC/init/$TYPE.sql
  )
done
popd &> /dev/null
