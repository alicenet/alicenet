#!/bin/sh

set -x

CURRENT_WD=$PWD
BRIDGE_DIR=./bridge
NETWORK=${1:-"dev"}

cd $BRIDGE_DIR

# npx hardhat run ./scripts/getDeployList.ts
npx hardhat --network $NETWORK --show-stack-traces updateDeploymentArgsWithFactory
npx hardhat run scripts/deployscripts.ts --no-compile --network $NETWORK --show-stack-traces

cd $CURRENT_WD

