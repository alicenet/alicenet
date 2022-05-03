#!/bin/bash

CURRENT_WD=$PWD
BRIDGE_DIR=./bridge

make build

cd $BRIDGE_DIR

npx hardhat node

cd $CURRENT_WD