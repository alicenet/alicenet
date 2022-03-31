#!/bin/sh

set -ex

NETWORK=${1:-"dev"}
ADDRESSES=$(ls ./scripts/generated/keystores/keys | grep -v '^0x546f99f244b' | xargs)
CURRENT_WD=$PWD
BRIDGE_DIR=./bridge

cd $BRIDGE_DIR


npx hardhat --network "$NETWORK" --show-stack-traces registerValidators --migrate-from-legacy --factory-address "$FACTORY_ADDRESS" $ADDRESSES


cd "$CURRENT_WD"


