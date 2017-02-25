#!/bin/bash

if [[ "$1" == "help" ]] || ( [[ "$1" == "release" ]] && [[ $# -ne 3 ]] ); then
    echo -e "\nUSAGE: ./build-all.sh [release] [USER] [TOKEN]"
    echo -e "\n    release  Specifiying 'release' will publish the release to Github"
	echo -e "    USER     Github user required to publish the release"
    echo -e "    TOKEN    Github API Token for invoking the publish APIs"
	echo -e "\nIf release is specified then USER and TOKEN must be provided.\n"
    exit 1
fi

which github-release 2>&1 > /dev/null
if [[ $? == 1 ]]; then
    echo -e "Unable to find the 'github-release' CLI. You need to download it"
    echo -e "from https://github.com/aktau/github-release/releases/tag/v0.6.2"
    echo -e "save the CLI binary to the system path."
    exit 1
fi

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

	# Create archives of release
	rm -f *.tar.gz
	rm -f *.zip

	tar cvzf cf-copy-plugin_linux64.tar.gz -C bin/linux64 cf-copy-plugin
	tar cvzf cf-copy-plugin_osx.tar.gz -C bin/osx cf-copy-plugin

	pushd bin/win64
	zip ../../cf-copy-plugin_win64.zip cf-copy-plugin.exe
	popd

	# Publish release
	USER=$2
	REPO=$(basename $(pwd))
	export GITHUB_TOKEN=$3

	echo "Creating Github release draft..."
	github-release release \
		--user $USER \
		--repo $REPO \
		--tag $TAG \
		--name "cf-copy-plugin" \
		--draft

	echo "Uploading cf-copy-plugin_linux64.tar.gz to release..."
	github-release upload \
		--user $USER \
		--repo $REPO \
		--tag $TAG \
		--name "cf-copy-plugin_linux64.tar.gz" \
		--file cf-copy-plugin_linux64.tar.gz

	echo "Uploading cf-copy-plugin_osx.tar.gz to release..."
	github-release upload \
		--user $USER \
		--repo $REPO \
		--tag $TAG \
		--name "cf-copy-plugin_osx.tar.gz" \
		--file cf-copy-plugin_osx.tar.gz

	echo "Uploading cf-copy-plugin_win64.zip to release..."
	github-release upload \
		--user $USER \
		--repo $REPO \
		--tag $TAG \
		--name "cf-copy-plugin_win64.zip" \
		--file cf-copy-plugin_win64.zip
		
	echo "Modifying Github release as final..."
	github-release edit \
		--user $USER \
		--repo $REPO \
		--tag $TAG \
		--name "cf-copy-plugin" \
		--description "Release version $TAG of cf-copy-plugin. It is recommend that you install this plugin via the community website https://plugins.cloudfoundry.org/."
fi

set +e
set +x