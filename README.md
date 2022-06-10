# AliceNet

## Requirements
#### Always required
* [Docker v20 with docker-compose](https://docs.docker.com/get-docker)
#### Required for working on the Golang client (until it is Dockerized)
* [Go 1.17](https://go.dev/dl/)
* [Geth 1.10.8](https://geth.ethereum.org/docs/install-and-build/installing-geth)
#### Required for working on Solidity contracts
* [Node 16](https://nodejs.org/en/download/)

## Build AliceNet
First, this repository needs to be cloned, and be the current working dir.

<br />

Then, to generate all files necessary for the project to build:
```
make generate
```
This user running this command needs to have the `docker` command set up. The generated files will belong to that user. 
The generate command above may need to be rerun after making certain changes. Check the [File generation](#file-generation)
 section for more info on this.

<br />

Then, to compile an executable:
```
make build
```

## File generation
The `make generate` command runs two subcommands:
* `make generate-bridge`
* `make generate-go`

Both commands run in a Docker container, so that nothing needs to be installed on your machine directly.

<br />

### Command: make generate-bridge
This command:
* (re)compiles all the solidity contracts
* (re)compiles the Go bindings for the solidity contracts
* (re)generates the ABI definitions for the solidity contracts

Rerun this every time you made changes to the solidity contracts and want these to be used by the madnet binary

Under the hood, this commend runs the `bridge` module's `compile` script to compile the contracts, and then the `bridge` module's `generate` script to generate the bindings and ABI definitions. These can also be run on the system directly, as long as the dependencies defined in `docker/generate-bridge` are installed.

<br />

### Command: make generate-go
This command:
* (re)compiles all the protobuf type definitions into Go sourcecode
* (re)compiles all the capnproto type definitions into Go sourcecode
* (re)compiles all the grpc endpoint definitions into Go sourcecode
* (re)generates convenient wrapper functions for the grpc endpoints using `cmd/mngen`
* (re)generates a new swagger json file based on the grpc endpoints
* (re)generates a Go source file containing the swagger json file in binary format, so it can be baked into the final executable

Rerun this command every time you made changes to the public API surface of AliceNet.

<br />

## Setup Test Environment

In order to initialize a chain with all validators online we need to generate the configuration files, keystores, geth
genesis file, and password files run this command (if already executed, run `./scripts/main.sh clean` first)

```
./scripts/main.sh init {# of validators 4-32}
```

This will generate all necessary files for the local chain inside: `./scripts/generated`

Open a terminal to start the bootnode with

```
./scripts/main.sh bootnode
```

Open another terminal to start geth with

```
./scripts/main.sh geth
```

Open now another terminal to deploy the contracts. This command will also register validators in the ValidatorPool. If you are having problems in this step, 
check [POSSIBLE TROUBLE SHOOTING](#TROUBLESHOOTING) section.

```
./scripts/main.sh deploy
```

Once this has finished, turn on each of the validators, each in its own terminal shell

```
./scripts/main.sh validator {# for the validator you want to start}
```

In the validators shell, there is this log looking like `Peers=24/x/3/0`, which gives you an idea on how many validator
have been discovered in the network for the selected validator. You will see the `x`
value increasing overtime . Eg if you initialized 5 validator the Peers value looks like

```shell
...
Peers=24/1/3/0
...
Peers=24/2/3/0
...
Peers=24/3/3/0
...
Peers=24/4/3/0
```

Once all the validators discovered the others and have peered together, then start ethDKG:

```
./scripts/main.sh ethdkg
```

This will print out blocks at which ethDKG events will happen.
> For a quicker local setup, you might want to change the `scripts/base-files/baseConfig` template and
> set `finalityDelay` from `6` -> `1` to speed up the ethdkg process.

Once it has been completed and AliceNet starts mining blocks, the system is ready.

Deposits are required in order to submit DataStores. Run the following at least 4 times in order to deposit enough funds
to inject datastores.

```
./scripts/main.sh deposit
```

Note that DataStores are injected in the [Wallet-JS tests](https://github.com/MadBase/MadNetWallet-v2), so submitting 
these deposits are required for the tests to be successful. 
At this point, the testnet should now be ready to run the standard tests.

To list other commands from the script simply run 
```
./scripts/main.sh 
```

# TROUBLESHOOTING

1. `go mod tidy`
   It is actually in place a fix for the following error

``` bash
./scripts/main.sh init 5
go mod tidy; \
	go build -o madnet ./cmd/main.go;
go: finding module for package github.com/MadBase/MadNet/localrpc/swagger-bindata
github.com/MadBase/MadNet/localrpc imports
	github.com/MadBase/MadNet/localrpc/swagger-bindata: no matching versions for query "latest"
localrpc/local.go:12:2: no required module provides package github.com/MadBase/MadNet/localrpc/swagger-bindata; to add it:
	go get github.com/MadBase/MadNet/localrpc/swagger-bindata
make: *** [build] Error 1
```

2. `./scripts/main.sh deploy`

```
+ npx hardhat --network dev --show-stack-traces updateDeploymentArgsWithFactory
  Loading deploymentArgs from: ./deployments/dev/deploymentArgsTemplate.json
  Future factory Address: 0x77D7c620E3d913AA78a71acffA006fc1Ae178b66
  Saving file at: ./deployments/dev/deploymentArgsTemplate.json
+ npx hardhat run scripts/deployscripts.ts --no-compile --network dev --show-stack-traces
  Deployed: MadnetFactory, at address: 0x77D7c620E3d913AA78a71acffA006fc1Ae178b66
  Error: cannot estimate gas; transaction may fail or may require manual gas limit
```

> Assuming you have `bridge/deployments/dev/deployList.json` defined, otherwise read [bridge documentation](bridge/README.md) file and then apply these changes
>```shell
>  cd bridge
>  npx hardhat run ./scripts/getDeployList.ts 
>  cd ..
>```
> now open the file `./scripts/getDeployList.ts` and move `"contracts/ETHDKG.sol:ETHDKG"` as the latest element in the `deployments` array.

# Different Node versions
In case you are using a different version of Node and you need to keep it you can choose to use one of many Node.js
version managements system such as [n - link](https://www.npmjs.com/package/n). Quick snippet to get the right version

```shell
$ npm prefix --global
$ sudo npm install -g n
$ echo $PATH
$ command -v n
$ n --version
$ n i lts
```

# WHAT'S NEXT

Now that AliceNet is successfully running on your machine, connect it
with [Madnet Wallet](https://github.com/MadBase/MadNetWallet-v2).

# TEST

### Verbose log
To show all the scripts logs in the console during tests you must set the env variable `ENABLE_SCRIPT_LOG` to `true`. 
For instance to run blockchain tests you will execute
```shell
ENABLE_SCRIPT_LOG=true go test -v github.com/MadBase/MadNet/blockchain
```

### Random Kill and Restart

Randomly kill and restart the individual validators. There should be no noticeable change in the behavior of the other
validators and AliceNet consensus should not be affected.

### Extended Kill and Restart

Randomly kill one validator for an extended period; in particular, let at least 10 blocks pass before restarting the
killed validator. This will cause the validator to be out of sync. The validator should be able to rejoin and resync
without much delay.

### Kill Half Nodes and Restart

Kill half the validators. This will cause the remaining validators to stop because they are unable to reach consensus
due to a lack of validators. After waiting for at least 20 seconds, restart the killed validators. After
resynchronizing, consensus should continue as before and blocks should be mined.

### Kill All Nodes and Restart

Shut down all the validators. Wait at least 20 seconds and restart all of them. Once the validators are resynchronized,
blocks should continue to be mined as though the shutdown did not occur.

### DataStore Consumption Test

There are 1024 blocks per epoch; this is defined in
`./constants/shared.go` as `EpochLength = 1024`. Blocks are mined approximately every 6 seconds, so one epoch lasts over
one hour. The DataStore test stores the datastore for 5 epochs, implying the data would be stored for over 8.5 hours. To
make this test more reasonable and ensure that datastore are consumed, change `EpochLength = 16`.

### View DataStore

Tests involving datastores require being able to determine whether the datastores are present. One way to ensure this is
using the Swagger API.

To understand how to run through this, we run an example. We assume that the test environment has been successfully
setup and that the Swagger API is available. Furthermore, validator4 must be running.

```
cd ./cmd/testutils/inject/
go run . -d -m=hello -i=foo
```

This test submits a DataStore of message `hello` at index `foo`
and then overwrites the DataStore, setting the value to `hello-two`
Here is a truncation of the output:

```
Running in DataStore Mode
...
...
...
secp
Consumed Value:12000    ValueOut:2695
DS:  index:4943a3029516011ac24ba62f5e4183eb6e7dbbe8a3c6644fbc3a515188eef7f8    deposit:2695    EpochOfExpire:7    msg:hello-two
Consumed Value:12000    ValueOut:2695
secp
Consumed Value:12000    ValueOut:12000
...
...
...
Getting Tx  err: rpc error: code = Unknown desc = unknown transaction: d3e6cc649be314aca13c9172c2b5bb9faa9a389440e277c323a0d9cbcfcd0ed5
Sending Tx
Sending Tx  err: rpc error: code = Unknown desc = the object is invalid:utxoID already in trie
GetMinedTransaction Tx  err: <nil>
```

Here, `d3e6cc649be314aca13c9172c2b5bb9faa9a389440e277c323a0d9cbcfcd0ed5`
(64 hex characters on the fourth-from-bottom line)
is the TxHash of the transaction of which overwrote the original transaction we submitted. As noted above, the original
message is `hello` and the new message is `hello-two`
In the Swagger API, if we click on `get-mined-transaction`
and set the TxHash value to
`d3e6cc649be314aca13c9172c2b5bb9faa9a389440e277c323a0d9cbcfcd0ed5`. Here is a portion of the transaction:

```
"DataStore": {
  "DSLinker": {
    "DSPreImage": {
      "ChainID": 42,
      "Index": "4943a3029516011ac24ba62f5e4183eb6e7dbbe8a3c6644fbc3a515188eef7f8",
      "IssuedAt": 1,
      "Deposit": 2695,
      "RawData": "68656c6c6f2d74776f",
      "Owner": "0301546f99f244b7b58b855330ae0e2bc1b30b41302f"
    },
    "TxHash": "d3e6cc649be314aca13c9172c2b5bb9faa9a389440e277c323a0d9cbcfcd0ed5"
  },
  "Signature": "03018e362b2f4fade93d4b06f1cd32b43a8905ccf6dce9f2326a0411a96b652211026671c240cc5d0a23335eabe10a5accd0199dc0dec050f0ec463af8130235557d00"
}
```

Although we could use the value of `Rawdata` to obtain its value, we will use another Swagger API call. This is useful
because upon consumption, `Rawdata` will not be present and this will ensure that the DataStore was consumed as
expected. Instead, we will use `get-data`, which requires the `CurveSpec`,
`Account`, and `Index`. We have `CurveSpec = 1` because the test uses the `Secp256k1`
elliptic curve. The value for `Account` comes from removing the first 4 characters of the `Owner` hexidecimal string.
`Index` can be copied directly from above. Thus, we enter

```
{
  "CurveSpec": 1,
  "Account": "546f99f244b7b58b855330ae0e2bc1b30b41302f",
  "Index": "4943a3029516011ac24ba62f5e4183eb6e7dbbe8a3c6644fbc3a515188eef7f8"
}
```

Upon execution, we have

```
{
  "Rawdata": "68656c6c6f2d74776f"
}
```

This matches `Rawdata` from the transaction and decodes to
`hello-two`, as expected.

If a DataStore is not present, then a "Key not found"
error will be returned; this error will happen when the DataStore is consumed.

# Interaction

### API Docs and Limited GUI

The swagger-ui for the localRPC of a validator may be found at `http://localhost:8885/swagger`. The default port for
`validator4` is `8888`. Thus, you may speak to validator 4 (a non-mining node) at `http://localhost:8888`

### Programmatic Interaction

The localRPC directory contains a client library that abstracts the localRPC system for easier development.
