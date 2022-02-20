#!/bin/sh
# Make sure everything runs
# Will eventually make it more in-depth

echo "$TMPDIR"
./*.shImg --help
[ $? -ne 0 ] && exit 1

echo "$TMPDIR"
./*.AppImage --help
[ $? -ne 0 ] && exit 1

exit 0
