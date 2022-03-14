#!/bin/sh
DATADIR=./local-geth/

rm -rf $DATADIR

make build

geth --datadir $DATADIR init ./scripts/generated/genesis.json

cp assets/test/keys/* $DATADIR/keystore/

./scripts/base-scripts/geth-local-resume.sh &
GETH_PID="$!"

# trap "trap - SIGTERM && kill -- $GETH_PID" SIGTERM SIGINT SIGKILL EXIT
trap "trap - SIGTERM && kill -- $GETH_PID" SIGTERM SIGINT SIGKILL EXIT

wait
# wait "$GETH_PID"
