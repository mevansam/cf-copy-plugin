#!/bin/bash

TAG=$1
echo $TAG | grep "^[0-9]*\.[0-9]*\.[0-9]*$" 2>&1 > /dev/null
if [[ $? == 1 ]]; then
    echo -e "\nUSAGE: ./createRelease.sh [TAG] [BRANCH]"
    echo -e "\n    TAG     Should be of the format #.#.#"
    echo -e "    BRANCH  Commits on this branch will be merged to master and tagged for the release\n"
    exit 1
fi

set -e

MERGE_BRANCH=$(git branch | awk '/^\*/{ print $2 }')
if [[ $MERGE_BRANCH == "master" ]]; then
    echo "Cannot create a release from the master branch. Release is created by merging changes from a feature branch."
    exit 1
fi
trap 'git checkout $MERGE_BRANCH' INT TERM EXIT

echo "Creating release $TAG by merging changes from branch $MERGE_BRANCH"

git checkout master
git pull
git merge $MERGE_BRANCH
git tag -a $TAG -m "CF copy plugin release $TAG - created by $(git config user.name)"
git push --follow-tags

git checkout $MERGE_BRANCH

set +e
