#!/bin/bash

set -e

NETWORK=${1:-"dev"}
# ADDRESSES=$(ls ./scripts/generated/keystores/keys | grep -v '^0x546f99f244b' | xargs)
ADDRESSES="0xA770BA8C45A194590EcB4984Bea28E4168eF832D"
CURRENT_WD=$PWD
BRIDGE_DIR=./bridge

cd $BRIDGE_DIR

addr="$(grep -Pzo "\[$NETWORK\]\ndefaultFactoryAddress = \".*\"\n" ../scripts/generated/factoryState | grep -a "defaultFactoryAddress = .*" | awk '{print $NF}')"

export FACTORY_ADDRESS="$(echo "$addr" | sed -e 's/^"//' -e 's/"$//')"
if [[ -z "${FACTORY_ADDRESS}" ]]; then
    echo "It was not possible to find Factory Address in the environment variable FACTORY_ADDRESS! Exiting script!"
    exit 1
fi

if [[ -z "${FACTORY_ADDRESS}" ]]; then
    echo "It was not possible to find Factory Address in the environment variable FACTORY_ADDRESS! Exiting script!"
    exit 1
fi

npx hardhat --network "$NETWORK" --show-stack-traces unregisterValidators --factory-address "$FACTORY_ADDRESS" $ADDRESSES


cd "$CURRENT_WD"

