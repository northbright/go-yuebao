#! /usr/bin/env sh

wget https://github.com/google/leveldb/archive/master.zip
unzip master.zip
cd leveldb-master
make
CGO_CFLAGS="-I./leveldb-master/include" CGO_LDFLAGS="-L./leveldb-master/out-shared -lsnappy" go get github.com/jmhodges/levigo



