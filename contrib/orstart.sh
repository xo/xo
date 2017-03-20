#!/bin/bash

DNAME=orcl

set -x

# update to latest
docker pull sath89/oracle-12c

docker stop $DNAME
docker rm $DNAME

# run instance
#  -p 8080:8080 \
docker run \
  -d \
  -p 1521:1521 \
  -v /media/src/opt/oracle:/u01/app/oracle \
  -e DBCA_TOTAL_MEMORY=1024 \
  --name $DNAME \
  sath89/oracle-12c

# create booktest database
if [ "$1" == "--create=yes" ]; then
  echo "sleeping"

  sleep 15

  SRC=$(realpath $(cd -P "$( dirname "${BASH_SOURCE[0]}" )" && pwd ))
  $SRC/orcreate.sh booktest booktest
fi
