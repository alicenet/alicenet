#!/bin/bash

set -x

CURRENT_WD=$PWD
BRIDGE_DIR=./bridge

cd $BRIDGE_DIR || exit

npx hardhat node --show-stack-traces &
trap 'pkill -9 -f hardhat' SIGTERM
wait

cd "$CURRENT_WD" || exit

