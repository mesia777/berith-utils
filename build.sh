#!/usr/bin/env bash

SCRIPT_PATH=$(cd "$(dirname $0)" && pwd)

rm -rf ${SCRIPT_PATH}/bin/berithutils
go build -o ${SCRIPT_PATH}/bin/berithutils ${SCRIPT_PATH}/cmd/berithutils/
