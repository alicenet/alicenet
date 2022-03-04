#!/bin/sh

set -x

CURRENT_WD=$PWD
BRIDGE_DIR=./bridge
# NETWORK=${1:-"dev"}

cd $BRIDGE_DIR

npx hardhat node --show-stack-traces

cd $CURRENT_WD

