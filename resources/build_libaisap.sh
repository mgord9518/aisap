#!/bin/sh

set -e

echo "Building Go functions for libaisap"
cd ../cbindings
go mod tidy
CC="zig cc" go build -buildmode c-archive -o ../libaisap-x86_64.a

# Don't need this auto-generated header file
rm ../libaisap-x86_64.h

echo "Building Zig functions for libaisap"
cd ../zig
zig build -Doptimize=ReleaseSafe
#zig build-lib \
#	lib/c_api.zig -lc \
#	-I .. \
#	-I squashfuse-zig/squashfuse \
#	-fcompiler-rt \
#	-fPIE \
#	-target x86_64-linux

# Extract both, then combine them into a single lib
ar -x  ../libaisap-x86_64.a
ar -x  zig-out/lib/libaisap.a
ar -qfc ../libaisap-x86_64.a *.o

# Clean up
rm *.o
