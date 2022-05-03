#!/bin/sh
CURRENT_WD=$PWD
BRIDGE_DIR=./bridge

cd $BRIDGE_DIR

addr="$(grep -Pzo "\[$NETWORK\]\ndefaultFactoryAddress = \".*\"\n" ../scripts/generated/factoryState | grep -a "defaultFactoryAddress = .*" | awk '{print $NF}')"
FACTORY_ADDRESS=$addr
echo $FACTORY_ADDRESS
npx hardhat spamEthereum --network dev --factory-address "$FACTORY_ADDRESS" --show-stack-traces


cd $CURRENT_WD