#!/bin/sh
DATADIR=./local-geth/

geth --miner.threads 1 --miner.gasprice 1 --miner.gaslimit 10000000 --miner.etherbase 546f99f244b7b58b855330ae0e2bc1b30b41302f --nodiscover --mine --txpool.nolocals --maxpeers 0 --ws --ws.addr=0.0.0.0 --ws.port=8546 --ws.api="eth,net,web3" --http --http.corsdomain '*' --networkid 42 --http.addr=0.0.0.0 --http.port=8545 --http.vhosts='*' --datadir=$DATADIR --http.api="admin,eth,net,web3,personal,miner" --allow-insecure-unlock -unlock 0 --password ./assets/test/keys/password.txt
