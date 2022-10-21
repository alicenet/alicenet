#!/bin/sh


CURRENT_WD=$PWD
BRIDGE_DIR=./bridge

cd $BRIDGE_DIR

#make the genesis node
 
aliceNetBlocknum=$(npx hardhat create-local-seed-node) &&
echo $aliceNetBlocknum
# testnetBlocknum=$(npx hardhat get-latest-blockheight)

#npx hardhat node --fork https://eth-mainnet.alchemyapi.io/v2/uhLzz4c430rgAJG6UKGLzaRrnLCkkl2o --fork-block-number 14542800 --show-stack-traces &&

#turn on impersonating
# npx hardhat enable-local-environment-impersonate --account 0xb9670e38d560c5662f0832cacaac3282ecffddb1 --network dev &&
#mine 9000 blocks
# npx hardhat mine-num-blocks --num-blocks 9000 --network dev &&
#pause validator at
#npx hardhat pause-consensus-at-height  --height $aliceNetBlocknum --signer 0xb9670e38d560c5662f0832cacaac3282ecffddb1 --network dev &&
#evict validators 
#npx hardhat unregister-all-validators --factory-address 0xA85Fcfba7234AD28148ebDEe054165AeF6974a65 --signer 0xb9670e38d560c5662f0832cacaac3282ecffddb1 --network dev &&
#npx hardhat start-local-seed-node --network dev


wait
cd $CURRENT_WD
