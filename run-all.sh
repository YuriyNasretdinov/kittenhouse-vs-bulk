#!/bin/bash

set -e
set -x

cleanup() {
    killall clickhouse kittenhouse inserter clickhouse-bulk || true
    sleep 0.1
    rm -rf /tmp/kittenhouse ~/dumps
}
trap cleanup SIGINT
sleep 0.1

export PATH=$PATH:$HOME/go/bin

# SLOW='-delay-per-byte=1us'
# SLOW='-delay-per-byte=500ns'
SLOW='-delay-per-byte=100ns'
# SLOW=

# PERSISTENT='-persistent=true'
PERSISTENT=''

# KITTENHOUSE='-kittenhouse=true'
KITTENHOUSE=

rm -f ~/*.log
mkdir -p /tmp/kittenhouse ~/dumps

cd $HOME/go/src/github.com/YuriyNasretdinov/kittenhouse-vs-bulk/clickhouse
go install -v

clickhouse $SLOW >~/clickhouse.log 2>&1 &

sleep 0.2

if [ -z "$KITTENHOUSE" ]; then
    cd $HOME
    clickhouse-bulk >~/ch-bulk.log 2>&1 &
else
    kittenhouse -u= -g= -port=8124 >~/kittenhouse.log 2>&1 &
fi

sleep 0.2

cd $HOME/go/src/github.com/YuriyNasretdinov/kittenhouse-vs-bulk/inserter
go install -v
inserter $PERSISTENT $KITTENHOUSE

tail -f ~/*.log

