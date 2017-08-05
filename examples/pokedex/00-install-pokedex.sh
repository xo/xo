#!/bin/bash

SRC=$(realpath $(cd -P "$( dirname "${BASH_SOURCE[0]}" )" && pwd ))

set -ex

pushd $SRC &> /dev/null

# get latest pokedex
if [ ! -d pokedex ]; then
  git clone https://github.com/veekun/pokedex.git pokedex
fi

# clean repo and pull latest
pushd pokedex &> /dev/null
git clean -f -x -d
git pull
popd &> /dev/null

# install
sudo python2.7 -m pip install psycopg2
sudo python2.7 -m pip install pymssql
sudo python2.7 -m pip install MySQL-python
sudo easy_install cx_Oracle
sudo python2.7 -m pip install -e ./pokedex/

popd &> /dev/null
