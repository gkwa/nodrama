#!/usr/bin/env bash

set -e

tmp=$(mktemp -d nodrama.XXXXX)

if [ -z "${tmp+x}" ] || [ -z "$tmp" ]; then
    echo "Error: \$tmp is not set or is an empty string."
    exit 1
fi

{
    rg --files . \
        | grep -v $tmp/filelist.txt \
        | grep -vE 'nodrama$' \
        | grep -v README.org \
        | grep -v make_txtar.sh \
        | grep -v go.sum \
        | grep -v go.mod \
        | grep -v Makefile \
        | grep -v cmd/main.go \
        | grep -v logger.go \
        # | grep -v nodrama.go \

} | tee $tmp/filelist.txt
tar -cf $tmp/nodrama.tar -T $tmp/filelist.txt
mkdir -p $tmp/nodrama
tar xf $tmp/nodrama.tar -C $tmp/nodrama
rg --files $tmp/nodrama
txtar-c $tmp/nodrama | pbcopy

rm -rf $tmp
