#!/bin/bash

# connects to the docker instance
# see: https://github.com/wscherphof/oracle-12c

HOST=$(docker port orcl 1521)

USER=$1
PASS=$2

if [ -z "$USER" ]; then
  USER=system
fi

if [[ "$PASS" == "" && $USER == "system" ]]; then
  PASS=manager
elif [[ "$PASS" == "" ]]; then
  PASS=$USER
fi

sqlplus $USER/$PASS@$HOST/orcl
