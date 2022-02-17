#!/bin/sh
# Make sure everything runs
# Will eventually make it more in-depth

./*.shImg --help
[ $0 -ne 0 ] && exit 1

./*.AppImage --help
[ $0 -ne 0 ] && exit 1
