#!/bin/bash

set -e

SRC=$(realpath $(cd -P "$( dirname "${BASH_SOURCE[0]}" )" && pwd ))

pushd $SRC &> /dev/null

for i in */config; do
  pushd pokedex &> /dev/null
  git reset --hard
  git clean -f -x -d
  popd &> /dev/null

  i=$(dirname $i)

  # skip
  if [[ -f $i/skip || -f $i/setup-skip ]]; then
    echo "skipping $i"
    continue
  fi

  source $i/config

  echo -e "------------------------------------------------------\n$i='$DB'"

  if [ -f $i/pre ]; then
    echo -e "\nsourcing $i/pre"
    source $i/pre
  fi

  if [ -z "$ENGINE" ]; then
    ENGINE=$DB
  fi

  set -x

  python2.7 -m pokedex setup -e "$ENGINE"

  set +x

  touch $i/setup-skip
done
