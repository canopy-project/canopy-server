#!/bin/bash
VERSION=`cat src/canopy/VERSION`
mkdir canopy-server-$VERSION
cp -r src canopy-server-$VERSION
tar -czvf canopy-server_${VERSION}.src.tar.gz canopy-server-$VERSION
rm -r canopy-server-$VERSION

