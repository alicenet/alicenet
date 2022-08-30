#!/bin/bash

set -e

NETWORK=${1:-"dev"}
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

npx hardhat --network "$NETWORK" --show-stack-traces scheduleMaintenance --factory-address "$FACTORY_ADDRESS"


cd "$CURRENT_WD"

