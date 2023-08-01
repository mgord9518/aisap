#!/bin/sh

set -e

cd "$(dirname $0)"

echo "Building Go functions for libaisap"
cd ../cbindings
go mod tidy
CC="zig cc" go build -buildmode c-archive -o ../libaisap-x86_64.a

# Don't need this auto-generated header file
rm ../libaisap-x86_64.h

echo "Building Zig functions for libaisap"
cd ../zig
zig build -Doptimize=ReleaseSafe

# Extract both, then combine them into a single lib
zig ar -x  ../libaisap-x86_64.a
zig ar -x  zig-out/lib/libaisap.a
zig ar -qc ../libaisap-x86_64.a *.o

# Clean up
rm *.o
