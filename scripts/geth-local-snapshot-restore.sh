#!/bin/bash
DATADIR=./local-geth/

set -e


rm -rf ~/validator0
rm -rf ~/validator1
rm -rf ~/validator2
rm -rf ~/validator3
rm -rf $DATADIR
unzip snapshot.zip
mv ./validator0 ~/validator0
mv ./validator1 ~/validator1
mv ./validator2 ~/validator2
mv ./validator3 ~/validator3
rm -rf ~/validator4
