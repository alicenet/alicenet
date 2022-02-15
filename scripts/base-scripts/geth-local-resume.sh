#!/bin/sh
DATADIR=./local-geth/
  
geth \
  --port "35555" \
  --networkid 1337 \
  --datadir=$DATADIR \
  --fakepow \
  --vmdebug \
  \
  --mine \
  --miner.threads 1 \
  --miner.gasprice 1 \
  --miner.gaslimit 10000000 \
  --miner.etherbase 546f99f244b7b58b855330ae0e2bc1b30b41302f \
  \
  --nodiscover \
  --txpool.nolocals \
  --maxpeers 0 \
  \
  --ws \
  --ws.addr=0.0.0.0 \
  --ws.port=8546 \
  --ws.api="admin,eth,net,web3,personal,miner,txpool,debug" \
  \
  --http \
  --http.addr=0.0.0.0 \
  --http.port=8545 \
  --http.api="admin,eth,net,web3,personal,miner,txpool,debug" \
  --http.vhosts='*' \
  --http.corsdomain='*' \
  \
  --allow-insecure-unlock \
  --unlock 0,1,2,3,4,5 \
  --password ./assets/test/keys/password.txt
