#!/bin/bash

DB=mssql://sa:changeit@localhost/

NAME=$1
if [ -z "$NAME" ]; then
  echo "need name"
  exit 1
fi

set -x

usql $DB -c "exec sp_configure 'contained database authentication', 1;"
usql $DB -c "reconfigure;"
usql $DB -c "drop login $NAME;"
usql $DB -c "drop database $NAME;"
usql $DB -c "create database $NAME containment=partial;"

usql $DB/$NAME -c "create login $NAME with password='$NAME', check_policy=off, default_database=$NAME;"
usql $DB/$NAME -c "create user $NAME for login $NAME with default_schema=$NAME;"
usql $DB/$NAME -c "create schema $NAME authorization $NAME;"
usql $DB/$NAME -c "exec sp_addrolemember 'db_owner', '$NAME';"
