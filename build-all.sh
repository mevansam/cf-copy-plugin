#!/bin/bash

set -x

TAG="$(git tag -l --points-at HEAD)"
if [[ "$1" == "release" ]] && [[ -n "$TAG" ]] ; then

	MAJOR=`echo $TAG |  awk 'BEGIN {FS = "." } ; { printf $1;}'`
	MINOR=`echo $TAG |  awk 'BEGIN {FS = "." } ; { printf $2;}'`
	BUILD=`echo $TAG |  awk 'BEGIN {FS = "." } ; { printf $3;}'`

	`sed -i -e "s/Major:.*/Major: $MAJOR,/" \
		-e "s/Minor:.*/Minor: $MINOR,/" \
		-e "s/Build:.*/Build: $BUILD,/" command/copy_plugin.go`
fi

GOOS=linux GOARCH=amd64 go build
LINUX64_SHA1=`cat cf-copy-plugin | openssl sha1`
mkdir -p bin/linux64
mv cf-copy-plugin bin/linux64/copy_plugin

GOOS=darwin GOARCH=amd64 go build
OSX_SHA1=`cat cf-copy-plugin | openssl sha1`
mkdir -p bin/osx
mv cf-copy-plugin bin/osx/copy_plugin

GOOS=windows GOARCH=amd64 go build
WIN64_SHA1=`cat cf-copy-plugin.exe | openssl sha1`
mkdir -p bin/win64
mv cf-copy-plugin.exe bin/win64/copy_plugin.exe

cat repo-index.yml |
sed "s/osx-sha1/$OSX_SHA1/" |
sed "s/win64-sha1/$WIN64_SHA1/" |
sed "s/linux64-sha1/$LINUX64_SHA1/" |
sed "s/_TAG_/$TAG/" |
sed "s/_TIMESTAMP_/$(date --utc +%FT%TZ)/" |
cat

if [[ "$1" == "release" ]] && [[ -n "$TAG" ]] ; then
	git tag -d $TAG
	git commit -am "Build version $TAG"
	git tag -a $TAG
	git push --follow-tags
fi

set +x