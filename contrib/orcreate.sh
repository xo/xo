#!/bin/bash

USER=system
PASS=oracle
SVC=xe
HOST=$(docker port orcl 1521)

NAME=$1

if [ -z "$NAME" ]; then
  echo "usage: $0 <NAME>"
  exit 1
fi

#sqlplus $USER/$PASS@$HOST/$SVC << ENDSQL

usql oracle://$USER:$PASS@$HOST/$SVC << ENDSQL
create tablespace $NAME nologging datafile '${NAME}.dat' size 100m autoextend on;

create user $NAME identified by $NAME default tablespace $NAME;

grant create session, create table, create view, create sequence, create procedure, create trigger, unlimited tablespace, select any dictionary to $NAME;

alter system set open_cursors=400 scope=both;
ENDSQL
