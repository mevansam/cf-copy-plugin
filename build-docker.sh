#!/bin/bash

set -e
set -x

TAG="$(git tag -l --points-at HEAD)"
if [[ -z "$TAG" ]]; then
    TAG=$4
fi

MAJOR=`echo $TAG |  awk 'BEGIN {FS = "." } ; { printf $1;}'`
MINOR=`echo $TAG |  awk 'BEGIN {FS = "." } ; { printf $2;}'`
BUILD=`echo $TAG |  awk 'BEGIN {FS = "." } ; { printf $3;}'`

sed -i.bak1 "s/Major:.*/Major: $MAJOR,/" command/copy_plugin.go
sed -i.bak2 "s/Minor:.*/Minor: $MINOR,/" command/copy_plugin.go
sed -i.bak3 "s/Build:.*/Build: $BUILD,/" command/copy_plugin.go

echo "Building docker image with plugin release $TAG."

if [[ -n "$TAG" ]] ; then
    GOOS=linux GOARCH=amd64 go build
    docker build . -t cf-cli-copy-release --squash

    cp command/copy_plugin.go.bak1 command/copy_plugin.go
    rm command/copy_plugin.go.bak*
    rm -f cf-copy-plugin

    if [[ -n $2 ]] && [[ -n $3 ]]; then
        docker login -u $2 -p $3

        docker tag cf-cli-copy-release $1/cf-cli-copy:latest
        docker tag cf-cli-copy-release $1/cf-cli-copy:$TAG
        docker push $1/cf-cli-copy

        # clean up
        docker rmi $1/cf-cli-copy:latest
        docker rmi $1/cf-cli-copy:$TAG
        docker rmi cf-cli-copy-release
    fi
fi

set +x
set +e
