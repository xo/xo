#!/bin/bash

PASS=changeit
HOST=sqlexpress

NAME=$1

if [ -z "$NAME" ]; then
  echo "need name"
  exit 1
fi

set -ex

mssql -u sa -p $PASS -s $HOST -q "exec sp_configure 'contained database authentication', 1;"
mssql -u sa -p $PASS -s $HOST -q "reconfigure;"
mssql -u sa -p $PASS -s $HOST -q "create database $NAME containment=partial;"
mssql -u sa -p $PASS -s $HOST -d $NAME -q "create user $NAME with password='$NAME';"
mssql -u sa -p $PASS -s $HOST -d $NAME -q "exec sp_addrolemember 'db_owner', '$NAME';"
mssql -u sa -p $PASS -s $HOST -d $NAME -q "create schema $NAME authorization $NAME;"
mssql -u sa -p $PASS -s $HOST -d $NAME -q "alter user $NAME with default_schema=$NAME;"
