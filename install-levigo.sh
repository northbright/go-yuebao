#! /usr/bin/env sh

wget https://github.com/google/leveldb/archive/master.zip
unzip master.zip
cd leveldb-master
make
echo $CGO_CFLAGS
echo $CGO_LDFLAGS



