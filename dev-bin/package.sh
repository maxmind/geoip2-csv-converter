#!/bin/bash

set -e

VERSION=$1

if [ ! "$VERSION" ]; then
    echo "Your must specify a version"
    exit 1
fi

echo -e "\nBuilding for Windows"
GOOS=windows GOARCH=amd64 go build
zip "geoip2-csv-converter-${VERSION}-windows-64.zip" geoip2-csv-converter.exe \
     README.md LICENSE

rm -f geoip2-csv-converter.exe

ARCH=amd64
for OS in linux darwin
do
    echo -e "\nBuilding for $OS"
    DIR="geoip2-csv-converter-${VERSION}"
    mkdir $DIR
    GOOS=$OS GOARCH=$ARCH go build
    mv geoip2-csv-converter $DIR
    cp README.md LICENSE $DIR
    tar cfvz "${DIR}-${OS}-${ARCH}.tar.gz" $DIR
    rm -r $DIR
done

git tag -a $VERSION
git push --tags
git push -u
