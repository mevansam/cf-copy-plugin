#!/bin/bash

set -x
set -e

TAG="$(git tag -l --points-at HEAD)"
if [[ "$1" == "release" ]] && [[ -n "$TAG" ]] ; then

	git checkout master
	
	git tag -d $TAG
	git push origin :refs/tags/$TAG	

	MAJOR=`echo $TAG |  awk 'BEGIN {FS = "." } ; { printf $1;}'`
	MINOR=`echo $TAG |  awk 'BEGIN {FS = "." } ; { printf $2;}'`
	BUILD=`echo $TAG |  awk 'BEGIN {FS = "." } ; { printf $3;}'`

	`sed -i -e "s/Major:.*/Major: $MAJOR,/" \
		-e "s/Minor:.*/Minor: $MINOR,/" \
		-e "s/Build:.*/Build: $BUILD,/" command/copy_plugin.go`

	git rm --cached bin/repo-index.yml
	git rm --cached bin/linux64/cf-copy-plugin
	git rm --cached bin/osx/cf-copy-plugin
	git rm --cached bin/win64/cf-copy-plugin.exe
fi

go get -u github.com/kardianos/govendor
govendor sync
govendor test +local

GOOS=linux GOARCH=amd64 go build
mkdir -p bin/linux64
mv cf-copy-plugin bin/linux64/cf-copy-plugin
LINUX64_SHA1=$(cat bin/linux64/cf-copy-plugin | openssl sha1 | awk '{ print $2 }')

GOOS=darwin GOARCH=amd64 go build
mkdir -p bin/osx
mv cf-copy-plugin bin/osx/cf-copy-plugin
OSX_SHA1=$(cat bin/osx/cf-copy-plugin | openssl sha1 | awk '{ print $2 }')

GOOS=windows GOARCH=amd64 go build
mkdir -p bin/win64
mv cf-copy-plugin.exe bin/win64/cf-copy-plugin.exe
WIN64_SHA1=$(cat bin/win64/cf-copy-plugin.exe | openssl sha1 | awk '{ print $2 }')

cat repo-index.yml |
sed "s/osx-sha1/$OSX_SHA1/" |
sed "s/win64-sha1/$WIN64_SHA1/" |
sed "s/linux64-sha1/$LINUX64_SHA1/" |
sed "s/_TAG_/$TAG/" |
sed "s/_TIMESTAMP_/$(date --utc +%FT%TZ)/" |
cat > bin/repo-index.yml

if [[ "$1" == "release" ]] && [[ -n "$TAG" ]] ; then

	git add bin/repo-index.yml
	git add bin/linux64/cf-copy-plugin
	git add bin/osx/cf-copy-plugin
	git add bin/win64/cf-copy-plugin.exe

	MSG="CF copy plugin binary releases for $TAG"

	git commit -am "$MSG"
	git tag -a "$TAG" -m "$MSG"
	git push --follow-tags
fi

set +e
set +x