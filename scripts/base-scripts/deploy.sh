#!/bin/bash

set -e

CURRENT_WD=$PWD
BRIDGE_DIR=./bridge
NETWORK=${1:-"dev"}

cd $BRIDGE_DIR

# if on hardhat network this switches automine on to deploy faster
npx hardhat --network "$NETWORK" set-local-environment-interval-mining --enable-auto-mine

# deploying legacy token (dummy erc20) to test migration of tokens
echo "Deploying legacy token"
npx hardhat --network "$NETWORK" deploy-legacy-token-and-update-deployment-args

# we pipe the output of the command to a file so we can grab the factory address later
npx hardhat --network "$NETWORK" --show-stack-traces deploy-contracts --skip-checks 2>&1 | tee ../scripts/generated/deployconfig.txt

# -f  will check for the file existence but -s will check for file existence along with file size greater than 0 (zero).
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

# inserting the factory address in the config.toml for each validator and the owner
for filePath in $(ls ../scripts/generated/config | xargs); do
    sed -e "s/factoryAddress = .*/factoryAddress = \"$FACTORY_ADDRESS\"/" "../scripts/generated/config/$filePath" >"../scripts/generated/config/$filePath".bk &&
        mv "../scripts/generated/config/$filePath".bk "../scripts/generated/config/$filePath"
done
cp ../scripts/base-files/owner.toml ../scripts/generated/owner.toml
sed -e "s/factoryAddress = .*/factoryAddress = \"$FACTORY_ADDRESS\"/" "../scripts/generated/owner.toml" >"../scripts/generated/owner.toml".bk &&
    mv "../scripts/generated/owner.toml".bk "../scripts/generated/owner.toml"

# funds validator accounts
npx hardhat fund-validators --network $NETWORK
# creating the bonus staking position to be redistributed at the end of the lockup period
npx hardhat --network $NETWORK create-bonus-pool-position --factory-address ${FACTORY_ADDRESS}
npx hardhat --network $NETWORK set-min-ethereum-blocks-per-snapshot --factory-address $FACTORY_ADDRESS --block-num 10

cd $CURRENT_WD

if [[ -n "${SKIP_REGISTRATION}" ]]; then
    cd $BRIDGE_DIR
    npx hardhat set-local-environment-interval-mining --network $NETWORK --interval 2500
    cd $CURRENT_WD
    echo "SKIPPING VALIDATOR REGISTRATION"
    exit 0
fi

./scripts/main.sh register

cd $BRIDGE_DIR
npx hardhat set-local-environment-interval-mining --network $NETWORK --interval 2500
cd $CURRENT_WD
