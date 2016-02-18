#!/bin/bash

go-bindata \
  -pkg templates \
  -prefix templates/ \
  -o templates/tpls.go \
  -ignore .go$ \
  -ignore .swp$ \
  -nometadata \
  -nomemcopy templates/*.tpl
