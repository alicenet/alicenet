#!/bin/bash

set -ex

CURRENT_WD=$PWD
BRIDGE_DIR=./bridge
NETWORK=${1:-"dev"}

cd $BRIDGE_DIR

# Deploy a dummy erc20 token called legacy, so we can turn them into ATokens to proceed with the
# other tasks. This task also updates the deploymentArgsTemplate with the legacyToken address and
# saves it in the ./scripts/generated folder
npx hardhat --network "$NETWORK" deployLegacyTokenAndUpdateDeploymentArgs
# Copy the deployList to the generated folder so we have deploymentList and deploymentArgsTemplate in the same folder
cp ../scripts/base-files/deploymentList ../scripts/generated/deploymentList

npx hardhat --network "$NETWORK" --show-stack-traces deployContracts --input-folder ../scripts/generated
addr="$(grep -zo "\[$NETWORK\]\ndefaultFactoryAddress = \".*\"\n" ../scripts/generated/factoryState | grep -a "defaultFactoryAddress = .*" | awk '{print $NF}')"
export FACTORY_ADDRESS=$addr
for filePath in $(ls ../scripts/generated/config | xargs); do
    sed -e "s/registryAddress = .*/registryAddress = $FACTORY_ADDRESS/" "../scripts/generated/config/$filePath" > "../scripts/generated/config/$filePath".bk &&\
    mv "../scripts/generated/config/$filePath".bk "../scripts/generated/config/$filePath"
done

cp ../scripts/base-files/owner.toml ../scripts/generated/owner.toml
sed -e "s/registryAddress = .*/registryAddress = $FACTORY_ADDRESS/" "../scripts/generated/owner.toml" > "../scripts/generated/owner.toml".bk &&\
mv "../scripts/generated/owner.toml".bk "../scripts/generated/owner.toml"

cd $CURRENT_WD

if [[ ! -z "${SKIP_REGISTRATION}" ]]; then
    echo "SKIPPING VALIDATOR REGISTRATION"
    exit 0
fi

FACTORY_ADDRESS="$(echo "$addr" | sed -e 's/^"//' -e 's/"$//')"
./scripts/main.sh register
if command -v gnome-terminal &>/dev/null; then
    i=1
    for filePath in $(ls ./scripts/generated/config | xargs); do
        gnome-terminal --tab --title="Validator $i" -- bash -c "./scripts/main.sh validator $i"
        i=$((i + 1))
    done
    exit 0
fi
echo -e "failed to auto start validators terminals, manually open a terminal for each validator and execute"
