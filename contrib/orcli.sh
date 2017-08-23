#!/bin/bash

# connects to oracle docker instance

HOST=$(docker port orcl 1521)

USER=$1
PASS=$2

SVC=xe

if [ -z "$USER" ]; then
  USER=system
fi

if [[ "$PASS" == "" && ( $USER == "system" || $USER == "sys" ) ]]; then
  PASS=oracle
elif [[ "$PASS" == "" ]]; then
  PASS=$USER
fi

set -x

#sqlplus $USER/$PASS@$HOST/$SVC
usql oracle://$USER:$PASS@$HOST/$SVC
