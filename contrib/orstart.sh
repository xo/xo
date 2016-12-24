#!/bin/bash

#docker run --privileged -dP --name orcl oracle-12c

docker run -d \
  -p 8080:8080 \
  -p 1521:1521 \
  -v /media/src/opt/oracle:/u01/app/oracle \
  -e DBCA_TOTAL_MEMORY=1024 \
  --name orcl \
  sath89/oracle-12c
