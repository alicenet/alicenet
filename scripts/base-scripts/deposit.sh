#!/bin/bash

set -e

CURRENT_WD=$PWD
BRIDGE_DIR=./bridge
NETWORK=${1:-"dev"}
cd $BRIDGE_DIR
addr="$(grep -Pzo "\[$NETWORK\]\ndefaultFactoryAddress = \".*\"\n" ../scripts/generated/factoryState | grep -a "defaultFactoryAddress = .*" | awk '{print $NF}')"
FACTORY_ADDRESS=${2:-"$(echo "$addr" | sed -e 's/^"//' -e 's/"$//')"}

npx hardhat --network "$NETWORK" --show-stack-traces virtual-mint-deposit \
--account-type 1 \
--deposit-owner-address "0x546F99F244b7B58B855330AE0E2BC1b30b41302F" \
--deposit-amount 1000 \
--factory-address $FACTORY_ADDRESS

cd $CURRENT_WD
