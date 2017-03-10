#!/bin/bash

set -e

TAG="$(git tag -l --points-at HEAD)"
if [[ -n "$TAG" ]] ; then

	MAJOR=`echo $TAG |  awk 'BEGIN {FS = "." } ; { printf $1;}'`
	MINOR=`echo $TAG |  awk 'BEGIN {FS = "." } ; { printf $2;}'`
	BUILD=`echo $TAG |  awk 'BEGIN {FS = "." } ; { printf $3;}'`

	`sed -i -e "s/Major:.*/Major: $MAJOR,/" \
		-e "s/Minor:.*/Minor: $MINOR,/" \
		-e "s/Build:.*/Build: $BUILD,/" command/copy_plugin.go`
else
    TAG=$1
fi

if [[ -n "$TAG" ]] ; then
    docker build . -t cf-cli-copy-release --squash --build-arg VER=$TAG

    if [[ -n $3 ]] && [[ -n $4 ]]; then
        docker login -u $3 -p $4

        docker tag cf-cli-copy-release $2/cf-cli-copy:latest
        docker tag cf-cli-copy-release $2/cf-cli-copy:$TAG
        docker push $2/cf-cli-copy

        # clean up
        docker rmi $2/cf-cli-copy:latest
        docker rmi $2/cf-cli-copy:$TAG
        docker rmi cf-cli-copy-release
    fi
	exit 0
fi

set +e
