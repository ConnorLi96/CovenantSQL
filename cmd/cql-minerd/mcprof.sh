#!/bin/sh -x

../../cleanupDB.sh
make -C ../../

(go test -bench=^BenchmarkMinerTwo$ -benchtime=40s -run ^$ &) && \
 (sleep 25 && DSN=`cat .dsn` go test '-bench=^BenchmarkClientOnly$' -benchtime=20s -run '^$')

