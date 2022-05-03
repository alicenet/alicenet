#!/bin/sh

set -x

CURRENT_WD=$PWD
BRIDGE_DIR=./bridge

cd $BRIDGE_DIR


npx hardhat --network dev --show-stack-traces registerValidators --factory-address "$@"


cd $CURRENT_WD


