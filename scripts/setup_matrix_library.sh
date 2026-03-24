#!/bin/sh

# Pinned commit from hzeller/rpi-rgb-led-matrix — update this when intentionally upgrading.
MATRIX_LIB_COMMIT="d259502789520e4ee415efff638605bd510f6509"

repodir=$(pwd)

cd $GOPATH/pkg/mod/github.com/tfk1410/go-rpi-rgb-led-matrix@v0.0.0-20210404121211-ed43f29cbccb
chmod 700 .
mkdir -p vendor
cd vendor
git clone https://github.com/hzeller/rpi-rgb-led-matrix.git
cd rpi-rgb-led-matrix
git checkout $MATRIX_LIB_COMMIT
sed -i.bak 's/.*RGB_SLOWDOWN_GPIO.*/DEFINE+=-DRGB_SLOWDOWN_GPIO=2/' lib/Makefile && rm -f lib/Makefile.bak
make -j

cd $repodir
go get -v github.com/tfk1410/go-rpi-rgb-led-matrix
