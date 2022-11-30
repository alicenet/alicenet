#!/bin/bash

set -e

CURRENT_WD=$PWD
BRIDGE_DIR=./bridge
NETWORK=${1:-"dev"}
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

npx hardhat --network "$NETWORK" --show-stack-traces virtual-mint-deposit \
    --account-type 1 \
    --deposit-owner-address "0x546F99F244b7B58B855330AE0E2BC1b30b41302F" \
    --deposit-amount 1000 \
    --factory-address $FACTORY_ADDRESS

cd $CURRENT_WD
