#!/bin/sh
DATADIR=./local-geth/

# rm -rf $DATADIR

# geth --datadir $DATADIR init ./scripts/genesis.json

# cp assets/test/keys/* $DATADIR/keystore/

geth --miner.threads 1 --miner.gasprice 1 --miner.gaslimit 10000000 --miner.etherbase 546f99f244b7b58b855330ae0e2bc1b30b41302f --nodiscover --mine --txpool.nolocals --maxpeers 0 --ws --wsaddr=0.0.0.0 --wsport=8546 --wsapi="eth,net,web3" --rpc --rpcaddr=0.0.0.0 --rpcport=8545 --datadir=$DATADIR --nousb --rpcapi="admin,eth,net,web3,personal,miner" --networkid 42
