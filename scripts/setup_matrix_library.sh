#!/bin/sh

cd $GOPATH/pkg/mod/github.com/tfk1410/go-rpi-rgb-led-matrix@v0.0.0-20210404121211-ed43f29cbccb
chmod 700 .
mkdir vendor
cd vendor
git clone https://github.com/hzeller/rpi-rgb-led-matrix.git
cd rpi-rgb-led-matrix
#sed -i '' 's/.*RGB_SLOWDOWN_GPIO.*/DEFINE+=-DRGB_SLOWDOWN_GPIO=2/' lib/Makefile
make -j