#!/bin/sh
VERSION=`cat src/canopy/VERSION`
rm -rf _debbuild
mkdir _debbuild
./create_src_pkg.sh
cp canopy-server_${VERSION}.src.tar.gz _debbuild/canopy-server_${VERSION}.orig.tar.gz
cd _debbuild
tar xf canopy-server_${VERSION}.orig.tar.gz
rm canopy-server_${VERSION}.orig.tar.gz
cp -r ../debian canopy-server-${VERSION}

# Generate debian/files file
echo "canopy-server_${VERSION}-1_amd64.deb misc optional" > canopy-server-${VERSION}/debian/files

cp -r ../scripts canopy-server-${VERSION}
cp -r ~/.canopy/golang canopy-server-${VERSION}
cd canopy-server-${VERSION}
dpkg-buildpackage -us -uc
cd ../..
cp _debbuild/canopy-server_${VERSION}-1_amd64.deb .
rm -rf _debbuild
