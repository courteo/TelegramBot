#/bin/bash

set -exuo pipefail

root=$PWD
arr=$(find . | grep '99_hw$' | grep -v '4/99_hw$' | grep -v '10/99_hw$' | grep -v '99_hw/code' |  grep -v '9/99_hw')
for i in $arr; do golangci-lint -c .golangci.yml run $i/...;done

if [ -d "$root/9/99_hw/server" ] 
then
    cd $root/9/99_hw/server
    golangci-lint -c $root/.golangci.yml run ./...
fi

if [ -d "$root/4/99_hw/taskbotr" ] 
then
    cd $root/4/99_hw/taskbot
    golangci-lint -c $root/.golangci.yml run --modules-download-mode=vendor ./...
fi
