#!/bin/sh
DATADIR=./local-geth2/

geth \
  --port "45555" \
  --authrpc.port 8552 \
  --networkid 1338 \
  --datadir=$DATADIR \
  --fakepow \
  --vmdebug \
  \
  --mine \
  --miner.threads 1 \
  --miner.gasprice 1 \
  --txpool.pricelimit 1 \
  --miner.gaslimit 10000000 \
  --miner.etherbase 546f99f244b7b58b855330ae0e2bc1b30b41302f \
  \
  --nodiscover \
  --maxpeers 0 \
  \
  --ws \
  --ws.addr=0.0.0.0 \
  --ws.port=8646 \
  --ws.api="admin,eth,net,web3,personal,miner,txpool,debug" \
  \
  --http \
  --http.addr=0.0.0.0 \
  --http.port=8645 \
  --http.api="admin,eth,net,web3,personal,miner,txpool,debug" \
  --http.vhosts='*' \
  --http.corsdomain='*' \
  \
  --allow-insecure-unlock \
  --unlock 0 \
  --password ./scripts/base-files/passwordFile
