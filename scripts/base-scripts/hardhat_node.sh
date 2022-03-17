#!/bin/sh

set -x

CURRENT_WD=$PWD
BRIDGE_DIR=./bridge
# NETWORK=${1:-"dev"}

cd $BRIDGE_DIR

npx hardhat node --show-stack-traces &
GETH_PID="$!"

# trap "trap - SIGTERM && kill -- $GETH_PID" SIGTERM SIGINT SIGKILL EXIT
trap "trap - SIGTERM && kill -- $GETH_PID" SIGTERM SIGINT SIGKILL EXIT

wait

cd $CURRENT_WD

