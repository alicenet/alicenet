#!/bin/sh

CURRENT_WD=$PWD
BRIDGE_DIR=../bridge

cd $BRIDGE_DIR

npx hardhat run scripts/deployscripts.ts --network dev_deploy --show-stack-traces

cd $CURRENT_WD

