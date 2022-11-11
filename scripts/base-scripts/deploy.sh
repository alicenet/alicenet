#!/bin/bash

set -e

CURRENT_WD=$PWD
BRIDGE_DIR=./bridge
NETWORK=${1:-"dev"}

cd $BRIDGE_DIR

# if on hardhat network this switches automine on to deploy faster
npx hardhat set-local-environment-interval-mining --network $NETWORK --interval 500

# Copy the deployList to the generated folder so we have deploymentList and deploymentArgsTemplate in the same folder
cp ../scripts/base-files/deploymentList ../scripts/generated/deploymentList
cp ../scripts/base-files/deploymentArgsTemplate ../scripts/generated/deploymentArgsTemplate

npx hardhat --network "$NETWORK" --show-stack-traces deploy-contracts --input-folder ../scripts/generated --wait-confirmation 1
addr="$(grep -Pzo "\[$NETWORK\]\ndefaultFactoryAddress = \".*\"\n" ../scripts/generated/factoryState | grep -a "defaultFactoryAddress = .*" | awk '{print $NF}')"

export FACTORY_ADDRESS=$addr
if [[ -z "${FACTORY_ADDRESS}" ]]; then
    echo "It was not possible to find Factory Address in the environment variable FACTORY_ADDRESS! Exiting script!"
    exit 1
fi

for filePath in $(ls ../scripts/generated/config | xargs); do
    sed -e "s/factoryAddress = .*/factoryAddress = $FACTORY_ADDRESS/" "../scripts/generated/config/$filePath" >"../scripts/generated/config/$filePath".bk &&
        mv "../scripts/generated/config/$filePath".bk "../scripts/generated/config/$filePath"
done

cp ../scripts/base-files/owner.toml ../scripts/generated/owner.toml
sed -e "s/factoryAddress = .*/factoryAddress = $FACTORY_ADDRESS/" "../scripts/generated/owner.toml" >"../scripts/generated/owner.toml".bk &&
    mv "../scripts/generated/owner.toml".bk "../scripts/generated/owner.toml"
# funds validator accounts
npx hardhat fund-validators --network $NETWORK
cd $CURRENT_WD

if [[ ! -z "${SKIP_REGISTRATION}" ]]; then
    echo "SKIPPING VALIDATOR REGISTRATION"
    exit 0
fi

FACTORY_ADDRESS="$(echo "$addr" | sed -e 's/^"//' -e 's/"$//')"

if [[ -z "${FACTORY_ADDRESS}" ]]; then
    echo "It was not possible to find Factory Address in the environment variable FACTORY_ADDRESS! Not starting the registration!"
    exit 1
fi

cd $BRIDGE_DIR
echo
# deploy ALCB
npx hardhat --network $NETWORK deploy-alcb --factory-address ${FACTORY_ADDRESS}
cd $CURRENT_WD

./scripts/main.sh register

cd $BRIDGE_DIR
npx hardhat --network $NETWORK set-min-ethereum-blocks-per-snapshot --factory-address $FACTORY_ADDRESS --block-num 10
npx hardhat set-local-environment-interval-mining --network $NETWORK --interval 2500
cd $CURRENT_WD

if [[ -n "${AUTO_START_VALIDATORS}" ]]; then
    if command -v gnome-terminal &>/dev/null; then
        i=1
        for filePath in $(ls ./scripts/generated/config | xargs); do
            gnome-terminal --tab --title="Validator $i" -- bash -c "./scripts/main.sh validator $i"
            i=$((i + 1))
        done
        exit 0
    fi
    echo -e "failed to auto start validators terminals, manually open a terminal for each validator and execute"
fi
