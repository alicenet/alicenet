#!/bin/sh

set -ex

CURRENT_WD=$PWD
BRIDGE_DIR=./bridge
NETWORK=${1:-"dev"}

cd $BRIDGE_DIR

# npx hardhat run ./scripts/getDeployList.ts
npx hardhat --network "$NETWORK" --show-stack-traces deployContracts --deploy-factory
addr="$(grep -Pzo "(?s)\[$NETWORK\]\ndefaultFactoryAddress = \"(.*?)\"\n" ../scripts/generated/factoryState | grep -a "defaultFactoryAddress = .*" | awk '{print $NF}')"
export FACTORY_ADDRESS=$addr
for filePath in $(ls ../scripts/generated/config | xargs); do
    sed -i "s/registryAddress = .*/registryAddress = $FACTORY_ADDRESS/" "../scripts/generated/config/$filePath"
done
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
