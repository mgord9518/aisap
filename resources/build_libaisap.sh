#!/bin/sh

set -e

cd "$(dirname $0)"

cd ../zig
zig build #-Doptimize=ReleaseSafe

mv zig-out/lib/libaisap.a ../libaisap-x86_64.a
