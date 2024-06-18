#!/bin/sh
# In order to build the test C program,
# `../resources/build_libaisap.sh` must first be run. Then, run this script,
# which will generate the `test` executable, which prints debug information
# about an AppImage

# TODO: also build `test.c` in `build.zig`

zig cc -static \
    -o test test.c \
    ../zig-out/lib/libaisap.a \
    ../zig-out/lib/libzstd.a \
    ../zig-out/lib/libdeflate.a \
    ../zig-out/lib/liblz4.a \
    ../zig-out/lib/libfuse.a \
    -I../../include
