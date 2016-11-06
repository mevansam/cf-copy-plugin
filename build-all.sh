#!/bin/bash

set -x
set -e

TAG="$(git tag -l --points-at HEAD)"
if [[ "$1" == "release" ]] && [[ -n "$TAG" ]] ; then

	MAJOR=`echo $TAG |  awk 'BEGIN {FS = "." } ; { printf $1;}'`
	MINOR=`echo $TAG |  awk 'BEGIN {FS = "." } ; { printf $2;}'`
	BUILD=`echo $TAG |  awk 'BEGIN {FS = "." } ; { printf $3;}'`

	`sed -i -e "s/Major:.*/Major: $MAJOR,/" \
		-e "s/Minor:.*/Minor: $MINOR,/" \
		-e "s/Build:.*/Build: $BUILD,/" command/copy_plugin.go`
fi

go get -u github.com/kardianos/govendor
govendor sync
govendor test +local

GOOS=linux GOARCH=amd64 go build
LINUX64_SHA1=`cat cf-copy-plugin | openssl sha1`
mkdir -p bin/linux64
mv cf-copy-plugin bin/linux64/cf-copy-plugin

GOOS=darwin GOARCH=amd64 go build
OSX_SHA1=`cat cf-copy-plugin | openssl sha1`
mkdir -p bin/osx
mv cf-copy-plugin bin/osx/cf-copy-plugin

GOOS=windows GOARCH=amd64 go build
WIN64_SHA1=`cat cf-copy-plugin.exe | openssl sha1`
mkdir -p bin/win64
mv cf-copy-plugin.exe bin/win64/cf-copy-plugin.exe

cat repo-index.yml |
sed "s/osx-sha1/$OSX_SHA1/" |
sed "s/win64-sha1/$WIN64_SHA1/" |
sed "s/linux64-sha1/$LINUX64_SHA1/" |
sed "s/_TAG_/$TAG/" |
sed "s/_TIMESTAMP_/$(date --utc +%FT%TZ)/" |
cat > bin/repo-index.yml

if [[ "$1" == "release" ]] && [[ -n "$TAG" ]] ; then

	MSG="CF copy plugin binary releases for $TAG"

	git checkout master
	git commit -am "$MSG"
	git push
fi

set +e
set +x