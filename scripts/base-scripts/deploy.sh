#!/bin/sh

CURRENT_WD=$PWD
BRIDGE_DIR=../bridge

cd $BRIDGE_DIR

npx hardhat run scripts/deployscripts.ts --network dev

cd $CURRENT_WD

