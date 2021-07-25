#!/bin/bash

SRC=$(realpath $(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd))

set -e

PROJECT=proj
APP=booktest
APP_CLASS="${APP^}"
TYPE=

OPTIND=1
while getopts "d:" opt; do
case "$opt" in
  d) TYPE=$OPTARG ;;
esac
done

if [ -z "$TYPE" ]; then
  echo "usage: $0 -d <DATABASE>"
  exit 1
fi

declare -A CONFIG
case $TYPE in
  mysql)
    CONFIG+=(
      [engine]=django.db.backends.mysql
      [host]=127.0.0.1
      [user]=django
      [pass]=django
      [name]=django
    )
    ;;
  oracle)
    CONFIG+=(
      [engine]=django.db.backends.oracle
      [user]=django
      [pass]=django
      [name]=127.0.0.1:1521/db1
    )
    ;;
  postgres)
    CONFIG+=(
      [engine]=django.db.backends.postgresql
      [host]=127.0.0.1
      [user]=django
      [pass]=django
      [name]=django
    )
    ;;
  sqlite3)
    CONFIG+=(
      [engine]=django.db.backends.sqlite3
      [name]=$SRC/django.db
    )
    ;;
  sqlserver)
    CONFIG+=(
      [engine]=mssql
      [host]=127.0.0.1
      [user]=django
      [pass]=django
      [name]=django
      [options]="'OPTIONS': {'driver': 'ODBC Driver 17 for SQL Server'},"
    )
    ;;
esac

mkdir -p $SRC/$TYPE
pushd $SRC/$TYPE &> /dev/null
# remove
(set -x;
  rm -rf $PROJECT
)
# create django project
(set -x;
  pipenv run django-admin startproject $PROJECT
)
popd &> /dev/null

pushd $SRC/$TYPE/$PROJECT &> /dev/null
# create app
(set -x;
  pipenv run ./manage.py startapp $APP
)

# config database
DATABASES=$(cat << END
DATABASES = {
    'default': {
      'ENGINE': '${CONFIG[engine]}',
      'HOST': '${CONFIG[host]}',
      'PORT': '${CONFIG[port]}',
      'USER': '${CONFIG[user]}',
      'PASSWORD': '${CONFIG[pass]}',
      'NAME': '${CONFIG[name]}',
      ${CONFIG[options]}
    },
}
END
)
echo "$DATABASES"
awk -i inplace -v x="$DATABASES" '/DATABASES\s*=\s*\{/{f=1} !f{print} /^}$/{print x; f=0}' $PROJECT/settings.py

# config models
perl -pi -e "s/^(INSTALLED_APPS.*)/\1\n    '$APP.apps.${APP_CLASS}Config',/" $PROJECT/settings.py
cp $SRC/models.py $APP/

# migrate
(set -x;
  pipenv run ./manage.py makemigrations $APP
  pipenv run ./manage.py migrate
)
popd &> /dev/null
