#!/bin/sh
# make build && ./madnet --config ./scripts/base-files/owner.toml --deploy.migrations=true deploy

CURRENT_WD=$PWD
BRIDGE_DIR=../bridge

cd $BRIDGE_DIR

# npx hardhat run scripts/getDeployList.ts --network dev

# # todo: replace MadToken owner address
# npx hardhat run scripts/getDeploymentArgs.ts --network dev

npx hardhat run scripts/deployscripts.ts --network dev

cd $CURRENT_WD

