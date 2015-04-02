#!/bin/sh
rm -rf _debbuild
mkdir _debbuild
./create_src_pkg.sh
cp canopy-server_15.04.03.orig.tar.gz _debbuild
cd _debbuild
tar xf canopy-server_15.04.03.orig.tar.gz
rm canopy-server_15.04.03.orig.tar.gz
cp -r ../debian canopy-server-15.04.03
cp -r ../scripts canopy-server-15.04.03
cp -r ~/.canopy/golang canopy-server-15.04.03
cd canopy-server-15.04.03
dpkg-buildpackage -us -uc
cd ../..
