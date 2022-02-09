#!/bin/sh
DATADIR=./local-geth/

rm -rf $DATADIR

geth --datadir $DATADIR init ./scripts/generated/genesis.json

cp assets/test/keys/* $DATADIR/keystore/
  
./scripts/base-scripts/geth-local-resume.sh