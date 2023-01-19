#!/bin/bash

CURRENT_WD=$PWD
BRIDGE_DIR=./bridge

make build

cd $BRIDGE_DIR

npx hardhat node --config hardhat2.config.ts --port 8645

cd $CURRENT_WD
