# How to configure and run your node

This document will guide you through on how to configure and run your AliceNet node against one of the AliceNet Networks available.

## Creating the alicenet folder

Before running an AliceNet node, we must setup some files and dependencies that will be used by the node. Firstly, let's create a folder where we are going to store all the AliceNet dependencies. During this documentation, we will be creating a folder on the $HOME directory, but you can change the folder to a desired location, just make sure to input the folder on the necessary configuration files.

If you want to run your node against AliceNet mainnet, open a new terminal and run the following command:

```shell
mkdir -p ~/alicenet/mainnet/keystores/keys
```

If you want to run your node against AliceNet testnet, use the following command:

```shell
mkdir -p ~/alicenet/testnet/keystores/keys
```

Copy the alicenet binary and paste to `~/alicenet` folder. You may need to unzip the binary first, in case you downloaded the binary from the [release page](https://github.com/alicenet/alicenet/releases).

Now, let's step in inside the newly created directory:
```shell
cd ~/alicenet
```

## Creating a configuration file

To create and configure a `config.toml` file, so the node can be executed in the chosen network.

```toml
[loglevel]
# Controls the logging level of the alicenet sub-services. To get the complete
# list of supported services check ./constants/shared.go
admin = "info"
blockchain = "info"
consensus = "info"

[chain]
# AliceNet ChainID that corresponds to the AliceNet Network that you are trying
# to connect.
id = "<CHAIN_ID>"
# Path to the location where to save the state database. The state database is
# responsible for storing the AliceNet blockchain data (blocks, validator's
# data). Store this database in a safe location. If this database is deleted,
# the node will re-sync with its peers from scratch. DON'T DELETE THIS DATABASE
# IF YOU ARE RUNNING A VALIDATOR NODE!!! If this database is deleted and you
# are running a validator node, the validator's data will be permanently
# deleted and the node will not be able to proceed with its work as a
# validator, even after a re-sync. Therefore, you may be susceptible to a
# slashing event.
stateDB = "<MY_PATH>/"
# Path to the location where to save the transaction database. The transaction
# database is responsible for storing the AliceNet blockchain data
# (transactions). Store this database in a safe location. If this database is
# deleted, the node will re-sync all transactions with its peers.
transactionDB = "<MY_PATH>/"
# Path to the location where to save the monitor database. The monitor database
# is responsible for storing the (events, tasks, receipts) coming from layer 1
# blockchains that AliceNet is anchored with. Store this database in a safe
# location. If this database is deleted, the node will replay all events by
# querying the layer1 chains using the information provided below.
monitorDB = "<MY_PATH>/"
# Flags to save any of the databases above only on memory. USE ONLY RECOMMENDED
# TO SET TRUE FOR TESTING PURPOSES.
monitorDBInMemory = false
stateDBInMemory = false
transactionDBInMemory = false

[transport]
# IF UPNP should be used to discover opened ports to connect with the peers.
upnp = false
# 16 Byte private key that is used to encrypt and decrypt information shared
# with peers. Generate this with a safe random generator.
privateKey = "<16_BYTES_TRASNPORT_PRIVATE_KEY>"
# Address to a bootnode running on the desired AliceNet network that you are
# trying to connect with. A bootnode is a software client responsible for
# sharing information about alicenet peers. Your node will connect to a
# bootnode to retrieve an initial list of peers to try to connect with.
bootNodeAddresses = "<BOOTNODE_ADDRESS>"
# Maximum number of peers that we can retrieve from the bootnode?
originLimit = 50
# Address and port where your node will be listening for rpc requests.
localStateListeningAddress = "0.0.0.0:8883"
# Address and port where you node will be listening for requests coming from
# other peers. The address should be publicly reachable.
p2pListeningAddress = "0.0.0.0:4342"
# Maximum number of peers that you wish to be connected with. Upper bound to
# limit bandwidth shared with the peers.
peerLimitMax = 24
# Minimum number of peers that you wish to be connect with, before trying to
# attempt to download blockchain data and participate in consensus.
peerLimitMin = 3

[ethereum]
# Ethereum endpoint url to the ethereum chain where the AliceNet network
# infra-structure that you are trying to connect lives. Ethereum mainnet for
# AliceNet mainnet and Goerli for Alicenet Testnet. Infura and Alchemy services
# can be used, but if you are running your own validator node, we recommend to
# use a more secure node.
endpoint = "<ETHEREUM_ENDPOINT_URL>"
# Minimum number of peers connected to your ethereum node that you wish to
# reach before trying to process ethereum blocks to retrieve the AliceNet
# events.
endpointMinimumPeers = 0
# Ethereum address that will be used to sign transactions and connect to the
# AliceNet services on ethereum.
defaultAccount = "<0xETHEREUM_ADDRESS>"
# Path to the encrypted private key used on the address above.
keystore = "<PATH>/keystores/keys"
# Path to the file containing the password to unlock the account private key.
passCodes = "<PATH>/keystores/passcodes.txt"
# Ethereum address of the AliceNet factory of smart contracts. The factory is
# responsible for registering and deploying every contract used by the AliceNet
# infra-structure.
factoryAddress = "<0xFACTORY_ETHEREUM_ADDRESS>"
# The ethereum block where the AliceNet contract factory was deployed. This
# block is used as starting block to retrieve all events (e.g snapshots,
# deposits) necessary to initialize and validate your AliceNet node.
startingBlock = <StartingBlock>
# Batch size of blocks that will be downloaded and processed from your endpoint
# address. Every ethereum block starting from the `startingBlock` until the
# latest ethereum block will be downloaded and all events (e.g snapshots,
# deposits) coming from AliceNet smart contracts will be processed in a
# chronological order. If this value is too large, your endpoint may end up
# being overloaded with API requests.
processingBlockBatchSize = 1_000
# The maximum gas price that you are willing to pay (in GWEI) for a transaction
# done by your node. If you are validator, putting this value too low can
# result in your node failing to fulfill the validators duty, hence, being
# passive for a slashing.
txMaxGasFeeAllowedInGwei = 500
# Flag to decide if the ethereum transactions information will be shown on the
# logs.
txMetricsDisplay = false

[utils]
# Flag to decide if the status will be shown on the logs. Maybe be a little
# noisy.
status = true

# OPTIONAL: Only necessary if you plan to run a validator node.
[validator]
# Type of elliptic curve used to generate the AliceNet address. 1: secp256k1 (same
# as ethereum), 2: BN128
rewardCurveSpec = 1
# Address of the AliceNet account used to do transactions in the AliceNet
# network.
rewardAccount = "0x<ALICENET_ADDRESS>"
symmetricKey = "<SOME_SUPER_FANCY_SECRET_THAT_WILL_BE_HASHED>"
```

## AliceNet Networks

Right now, there are 2 AliceNet networks being run. AliceNet main (anchored on ethereum mainnet) and AliceNet testnet (anchored on goerli ethereum testnet).

TODO: add the table
| AliceNet Network | Ethereum Network | ChainID | Bootnode address                                                                               | Ethereum Smart Contract Factory Address    | Starting Block |
| ---------------- | ---------------- | ------- | ---------------------------------------------------------------------------------------------- | ------------------------------------------ | -------------- |
| main             | mainnet          | 21      | 00000015\|029570051a8573e865af31a066eb100e7744bcbd05d814e899a763500163675be9@127.0.0.1:4242 | 0x0000000000000000000000000000000000000000 | 00000000       |
| testnet          | goerli           | 42      | 0000002A\|029570051a8573e865af31a066eb100e7744bcbd05d814e899a763500163675be9@127.0.0.1:4242 | 0x0000000000000000000000000000000000000000 | 00000000       |


## Running the node

Finally, to run your node, open a new terminal and execute:

```shell
<PATH>/alicenet --config <PATH>/config.toml validator
```

E.g If you copied and pasted the config

```shell
./alicenet --config ./config.toml validator
```
