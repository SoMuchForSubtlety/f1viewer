#!/bin/bash
now=$(date --iso-8601=seconds)
go build -v -ldflags="-X main.version=${GITHUB_REF/refs\/tags\/v/} -X main.commit=${GITHUB_SHA} -X main.date=$now"
if [ $? -eq 0 ]; then
    echo build OK, creating archive
    tar -zcvf f1viewer_"${GITHUB_REF/refs\/tags\/v/}"_macOS_64-bit.tar.gz LICENSE README.md f1viewer
else
    echo FAILED TO BUILD
    exit 1
fi
