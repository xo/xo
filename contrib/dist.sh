#!/bin/bash

goxc -v -pv=0.9.0 -build-tags=oracle -main-dirs-exclude=out,examples -tasks-=go-test -arch="amd64" -os="linux" -d=./out/
