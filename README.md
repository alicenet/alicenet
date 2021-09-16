# MadNet
Mad Network Layer 1


## Environment Setup
First, clone this repository.
Next, clone the bridge repository to the same directory as MadNet;
that is, if this repository is located at the path `~/mn/MadNet`,
then the bridge repository should be located at the path `~/mn/bridge`.
Also, clone the MadNetWallet-JS repository;
directions for installing the Wallet software will be described below.


## Build MadNet
To build the project, run the following command in the root
of the repository:
```
make build
```


## Build Wallet

### Install NodeJS
Run the following commands to install `NodeJS`:
```
sudo apt -y install curl dirmngr apt-transport-https lsb-release ca-certificates
curl -sL https://deb.nodesource.com/setup_12.x | sudo -E bash -
sudo apt-get install -y nodejs
```
This is required to build the wallet.

### Install Wallet
Inside the MadNetWallet-JS repository,
execute the following to build the wallet:
```
npm install
```
In the same repository, we now need to create
`./tests/.env` with the following contents:
```
PRIVATE_KEY="6aea45ee1273170fb525da34015e4f20ba39fe792f486ba74020bcacc9badfc1"
CHAIN_ID="42"
RPC="http://127.0.0.1:8888/v1/"
```
This is required for the tests to run.

## Setup Test Environment

### Local execution

In order to initialize a chain with all validators
online and a valid group key, we use the `snapshot.zip` file
and then proceed from a valid snapshot of the chain.
The commands which follow enable all tests to be run
except for the ETHDKG test;
this test is described in its entirety [here](#ethdkg-test).

First, build the latest version of the repository.
We now setup the bootnode to allow for communication between
the nodes:
```
./scripts/bootnode.sh
```

First execute
```
./scripts/geth-local-snapshot-restore.sh
```
This will restore the testnet Ethereum database and
the validator databases and allows us to not have to run ETHDKG
to test functionality.
This script **must** run to completion before continuing the test setup.

Open a new terminal and run
```
./scripts/geth-local-resume.sh
```

Open five additional terminals and execute the validator scripts
```
./scripts/validator0.sh
./scripts/validator1.sh
./scripts/validator2.sh
./scripts/validator3.sh
./scripts/validator4.sh
```
The first 4 validators (validator0 -- validator3) mine transactions.
validator4 sends transactions to the other validators.

The validators load their configuration from `assets/config`, and use the following ports for P2P communication:

- `validator0`: 4242
- `validator1`: 4243
- `validator2`: 4244
- `validator3`: 4245
- `validator4`: 5343

Deposits are required in order to submit DataStores.
Run the following at least 4 times in order to deposit enough funds
to inject datastores.
```
./madnet --config=./assets/config/owner.toml utils deposit
```
Note that DataStores are injected in the Wallet-JS
[tests](#wallet-js-tests),
so submitting these deposits are required for the
tests to be successful.

At this point, the testnet should now be ready to run the standard tests.


### With Docker Compose

A local test cluster can be started with Docker Compose. Open a new terminal and run:
```
docker-compose up
```

This will build the docker images for `geth` and `madnet`, and start geth, the bootnode and validators.
The five validators are performing the same task as with the local execution and mount their configuration
from `assets/config`. They do bind the P2P poers on the bost machine.


## Test Sequences
In order to run the Wallet-JS and standard MadNet
tests, ensure the Test Environment is set up according to
[Setup Test Environment](#setup-test-environment).
The ETHDKG Test runs the entire distributed key generation
procedure and has its own protocol.

### Wallet-JS Tests
First, ensure the wallet has been compiled.
In the root of the Wallet repository, run the following:
```
npm test
```
Use `mocha` to run a specific subset of tests.
For example, execute the following to run the Account tests:
```
npx mocha --timeout 45000 tests/account.js
```

### Standard MadNet Tests
First, run the Wallet-JS tests; only run these MadNet tests
if all the wallet tests pass.

#### Inject DataStore Test
The first test we run is to inject datastores into MadNet.
Go to `./cmd/testutils/inject/` and run the following command:
```
go run . -d -m=<message> -i=<index>
```
As noted in the test setup, deposits to MadNet are required in
order for the datastore tests to succeed.
Successful completion of this test requires that the
datastores be consumed upon expiration;
datastore expiration and consumption is discussed
[below](#datastore-consumption-test)
and is not considered part of this test.
We describe how to interact with MadNet in order to view the contents
of a DataStore [here](#view-datastore).

#### Spam Transactions Test
Upon successful completion of DataStore injection, we proceed
to inject many transactions to Madnet:
```
go run . -b <baseIdx> -n <numChildren>
```
`numChildren` will specify the number of the transactions
that are submitted every second.

#### Random Kill and Restart
Randomly kill and restart the individual validators.
There should be no noticeable change in the behavior of
the other validators and MadNet consensus should not be affected.

#### Extended Kill and Restart
Randomly kill one validator for an extended period;
in particular, let at least 10 blocks pass before restarting
the killed validator.
This will cause the validator to be out of sync.
The validator should be able to rejoin and resync without
much delay.

#### Kill Half Nodes and Restart
Kill half the validators.
This will cause the remaining validators to stop because
they are unable to reach consensus due to a lack of validators.
After waiting for at least 20 seconds, restart the killed validators.
After resynchronizing, consensus should continue as before
and blocks should be mined.

#### Kill All Nodes and Restart
Shut down all the validators.
Wait at least 20 seconds and restart all of them.
Once the validators are resynchronized, blocks should continue
to be mined as though the shutdown did not occur.

#### DataStore Consumption Test
There are 1024 blocks per epoch; this is defined in
`./constants/shared.go` as `EpochLength = 1024`.
Blocks are mined approximately every 6 seconds,
so one epoch lasts over one hour.
The DataStore test stores the datastore for 5 epochs,
implying the data would be stored for over 8.5 hours.
To make this test more reasonable and
ensure that datastores are consumed, change
`EpochLength = 16`.

### ETHDKG Test
This test requires a different set of commands than
the previous tests and all commands have been included.
First ensure the project has been built.
From here, start the bootnode:
```
./scripts/bootnode.sh
```

Now, setup a local geth environment:
```
./scripts/geth-local.sh
```
This differs from the standard tests because the standard
tests store a valid snapshot of data after ETHDKG has been
completed in order to save time when running tests.
On the Ethereum mainnet, this test would take approximately
40 minutes; on the testnet, this test should take approximately
20 minutes.

Remove all directories containing previous validator information:
```
rm -r ~/validator*
```

We now deploy the ETHDKG smart contract, approve tokens,
transfer tokens to the validators, and register the validators:
```
./scripts/deploy.sh && ./scripts/approvetokens.sh && ./scripts/transfertokens.sh && ./scripts/register.sh
```
This script **must** run to completion before the test continues.

We now turn on the validators:
```
./scripts/validator0.sh
./scripts/validator1.sh
./scripts/validator2.sh
./scripts/validator3.sh
```
Note that we do not run validator4 at this time;
it is not required for this test.

We now wait for the validators to synchronize with the
Ethereum (or testnet) blockchain.
We will look for `InSync: false -> true` in the log statements.
Once the validators are synched, we finally run ETHDKG:
```
./scripts/ethdkg.sh
```
Once completed, the validators will start mining blocks.


## Interaction

### API Docs and Limited GUI
The swagger-ui for the localRPC of a validator may be found at
http://localhost:####/swagger/ ;
the default port for validator4 is 8888.
Thus, you may speak to validator 4 (a non-mining node) at
http://localhost:8888/swagger/ .

#### View DataStore
Tests involving datastores require being able to determine
whether the datastores are present.
One way to ensure this is using the Swagger API.

To understand how to run through this, we run an example.
We assume that the test environment has been successfully
setup and that the Swagger API is available.
Furthermore, validator4 must be running.

We submit
```
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
is the TxHash of the transaction of which overwrote the original
transaction we submitted.
As noted above, the original message is `hello` and the new message
is `hello-two`
In the Swagger API, if we click on `get-mined-transaction`
and set the TxHash value to
`d3e6cc649be314aca13c9172c2b5bb9faa9a389440e277c323a0d9cbcfcd0ed5`.
Here is a portion of the transaction:
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

Although we could use the value of `Rawdata` to obtain its value,
we will use another Swagger API call.
This is useful because upon consumption, `Rawdata` will not be
present and this will ensure that the DataStore was consumed
as expected.
Instead, we will use `get-data`, which requires the `CurveSpec`,
`Account`, and `Index`.
We have `CurveSpec = 1` because the test uses the `Secp256k1`
elliptic curve.
The value for `Account` comes from removing the first 4 characters
of the `Owner` hexidecimal string.
`Index` can be copied directly from above.
Thus, we enter
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
error will be returned;
this error will happen when the DataStore is consumed.

### Programmatic Interaction
The localRPC directory contains a client library that abstracts the localRPC system for easier
development.

### Graphical Wallet
Coming soon
