#!/bin/bash

set -e

NETWORK=${1:-"dev"}
ADDRESSES=$(ls ./scripts/generated/keystores/keys | grep -v '^0x546f99f244b' | xargs)
CURRENT_WD=$PWD
BRIDGE_DIR=./bridge

cd $BRIDGE_DIR

if [[ -s ../scripts/generated/deployconfig.txt ]]; then
    addr=$(go run ../cmd/testutils/extractor/main.go -p ../scripts/generated/deployconfig.txt)
else
    echo "deployconfig file doesn't exist in scripts/generated/deployconfig.txt path. Exiting..."
    exit 1
fi

export FACTORY_ADDRESS=$addr
if [[ -z "${FACTORY_ADDRESS}" ]]; then
    echo "It was not possible to find Factory Address in the environment variable FACTORY_ADDRESS! Exiting script!"
    exit 1
fi

npx hardhat --network "$NETWORK" --show-stack-traces unregister-validators --factory-address "$FACTORY_ADDRESS" $ADDRESSES

cd "$CURRENT_WD"
