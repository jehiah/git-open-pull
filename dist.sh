#!/bin/bash

# build binary distributions for linux/amd64 and darwin/amd64
set -e 

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
echo "working dir $DIR"

rm -rf $DIR/vendor
echo "... refreshing vendor directory"
./vendor.sh

echo "... running tests"
gb test || exit 1

arch=$(go env GOARCH)
version=$(cat $DIR/src/cmd/git-open-pull/version.go | grep "const Version" | awk '{print $NF}' | sed 's/"//g')
goversion=$(go version | awk '{print $3}')

mkdir -p dist
for os in linux darwin; do
    echo "... building v$version for $os/$arch"
    BUILD=$(mktemp -d -t git-open-pull)
    TARGET="git-open-pull-$version.$os-$arch.$goversion"
    GOOS=$os GOARCH=$arch CGO_ENABLED=0 gb build
    mkdir -p $BUILD/$TARGET
    cp bin/git-open-pull-$os-$arch $BUILD/$TARGET/git-open-pull
    pushd $BUILD >/dev/null
    tar czvf $TARGET.tar.gz $TARGET
    if [ -e $DIR/dist/$TARGET.tar.gz ]; then
        echo "... WARNING overwriting dist/$TARGET.tar.gz"
    fi
    mv $TARGET.tar.gz $DIR/dist
    echo "... built dist/$TARGET.tar.gz"
    popd >/dev/null
done
