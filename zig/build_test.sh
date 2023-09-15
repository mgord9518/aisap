#!/bin/sh
# In order to build the test C program,
# `../resources/build_libaisap.sh` must first be run. Then, run this script,
# which will generate the `test` executable, which prints debug information
# about an AppImage

$CC -static \
    -o test test.c \
    ../libaisap-x86_64.a \
    -I../include

#    -lfuse3 \
