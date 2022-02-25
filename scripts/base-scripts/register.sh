#!/bin/sh

ADDRESSES=$(ls ./scripts/generated/keystores/keys | grep -v '^0x546f99f244b' | xargs)
CURRENT_WD=$PWD
BRIDGE_DIR=../bridge

cd $BRIDGE_DIR


npx hardhat registerValidators --network dev --show-stack-traces $ADDRESSES


cd $CURRENT_WD


