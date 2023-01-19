#!/bin/sh
DATADIR=./local-geth2/

rm -rf $DATADIR

make build

geth --datadir $DATADIR init ./scripts/generated/genesis2.json

cp ./scripts/base-files/0x546f99f244b7b58b855330ae0e2bc1b30b41302f $DATADIR/keystore/

./scripts/base-scripts/geth2-local-resume.sh
