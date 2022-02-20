#!/bin/sh
# Make sure everything runs
# Will eventually make it more in-depth

./*.shImg --help
[ $? -ne 0 ] && exit 1

mkdir tmp

./*.AppImage --help
TMPDIR='./tmp' [ $? -ne 0 ] && exit 1

exit 0
