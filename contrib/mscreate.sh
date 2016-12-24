#!/bin/bash

USER=sa
PASS=changeit
HOST=localhost

NAME=$1

if [ -z "$NAME" ]; then
  echo "need name"
  exit 1
fi

set -x

mssql -u $USER -p $PASS -s $HOST -q "exec sp_configure 'contained database authentication', 1;"
mssql -u $USER -p $PASS -s $HOST -q "reconfigure;"
mssql -u $USER -p $PASS -s $HOST -q "drop login $NAME;"
mssql -u $USER -p $PASS -s $HOST -q "drop database $NAME;"
mssql -u $USER -p $PASS -s $HOST -q "create database $NAME containment=partial;"
mssql -u $USER -p $PASS -s $HOST -d $NAME -q "create login $NAME with password='$NAME', check_policy=off, default_database=$NAME;"
mssql -u $USER -p $PASS -s $HOST -d $NAME -q "create user $NAME for login $NAME with default_schema=$NAME;"
mssql -u $USER -p $PASS -s $HOST -d $NAME -q "create schema $NAME authorization $NAME;"
mssql -u $USER -p $PASS -s $HOST -d $NAME -q "exec sp_addrolemember 'db_owner', '$NAME';"
