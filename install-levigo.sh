#! /usr/bin/env sh

wget https://github.com/google/leveldb/archive/master.zip
unzip master.zip
cd leveldb-master
make
CGO_CFLAGS="-I./include" CGO_LDFLAGS="-L./out-shared -lsnappy" go get github.com/jmhodges/levigo



