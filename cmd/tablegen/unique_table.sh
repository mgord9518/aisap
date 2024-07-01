#!/bin/sh

./tablegen | gawk '!a[$0]++'
