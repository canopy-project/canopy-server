#!/bin/sh
rm -r _debbuild
mkdir _debbuild
./create_src_pkg.sh
cp canopy-server_15.04.orig.tar.gz _debbuild
cd _debbuild
tar xf canopy-server_15.04.orig.tar.gz
cp -r ../debian canopy-server-15.04
cd canopy-server-15.04
debuild -us -uc
cd ../..
