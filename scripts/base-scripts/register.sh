#!/bin/bash

set -e

NETWORK=${1:-"dev"}
ADDRESSES=$(ls ./scripts/generated/keystores/keys | grep -v '^0x546f99f244b' | xargs)
CURRENT_WD=$PWD
BRIDGE_DIR=./bridge

cd $BRIDGE_DIR

if [[ -z "${FACTORY_ADDRESS}" ]]; then
    echo "It was not possible to find Factory Address in the environment variable FACTORY_ADDRESS! Exiting script!"
    exit 1
fi

npx hardhat --network "$NETWORK" --show-stack-traces registerValidators --factory-address "$FACTORY_ADDRESS" $ADDRESSES


cd "$CURRENT_WD"


