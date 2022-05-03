
#!/bin/sh
CURRENT_WD=$PWD
BRIDGE_DIR=./bridge
NETWORK=${1:-"dev"}
cd $BRIDGE_DIR

npx hardhat node

cd $CURRENT_WD