# Running a full local AliceNet network

For testing purposes, some times is useful to start a full local AliceNet network. In order to do so, you will need to run an ethereum test node, deploy all smart contract infra-structure, run all the validators and bootnode locally. This documentation will guide you step by step on how to run your own local AliceNet network.

## Requirements

In order to run the AliceNet local network, you must have followed and completed the [Building from Source](./BUILD.md) instructions.

**MacOs Users**

The bash scripts that will be executed in the next steps will require some flags that don't come out with the MacOs version of `grep`, `ls`, `find` and `sed`. Therefore, we recommend that you install the GNU version of these tools before proceeding. You can do this by running the following commands in your terminal:

```shell
brew install grep
brew install gnu-sed
brew install coreutils
brew install findutils
```

In sequence, add the following lines to your terminal configuration file (e.g ~/.bashrc, ~/.zshrc, etc):

```bash
export PATH="/usr/local/opt/grep/libexec/gnubin:$PATH"
export PATH="/usr/local/opt/gnu-sed/libexec/gnubin:$PATH"
export PATH="$(brew --prefix)/opt/findutils/libexec/gnubin:$PATH"
export PATH="/usr/local/opt/coreutils/libexec/gnubin:${PATH}"
alias ls="/usr/local/opt/coreutils/libexec/gnubin/ls"
```

## Setup Test Environment

If already executed this step before, you may want to run the following command first to clean any leftover files from a previous execution.

```shell
./scripts/main.sh clean
```

The first step to run you local network is to create all configuration files. To do this, open a new terminal and then run:

```shell
./scripts/main.sh init {# number of validators 4-32}
```

The command above will generate the configuration files, keystores, genesis file, and password files inside: `./scripts/generated` folder.

## Running a bootnode

As a next step, you need to run an AliceNet bootnode client. A bootnode is special AliceNet client that is responsible for facilitating the connection between nodes in the network. Once a normal node starts, it reaches to a bootnode defined in the `config.toml` file to get information about other peers.

To run a bootnode, execute the following command in a new terminal:

```shell
./scripts/main.sh bootnode
```

## Running an ethereum test node

The next step is to run an ethereum test network in order to deploy the AliceNet smart contract intra-structure.

To start the ethereum test node, just open a new terminal and run:

```shell
./scripts/main.sh hardhat
```

The above command will start a [hardhat node instance](https://hardhat.org/hardhat-network/docs/overview) locally, listening at `http://127.0.0.1:8545`

## Deploying the AliceNet Ethereum smart contracts locally

In order to deploy the AliceNet smart contracts on your local ethereum network, just execute the following command in a new terminal:

```shell
./scripts/main.sh deploy
```

The command above will also register the validators created by the [init](#setup-test-environment) command and setup some AliceNet configurations in the smart contracts. After the command is executed, you should see something like this on your terminal:

```shell
Deployed: AliceNetFactory, at address: 0x0b1F9c2b7bED6Db83295c7B5158E3806d67eC5bc
Deployed Metamorphic for AToken (ALCA) at: 0x4426F1939966427C436c565a4532e41028c60301, deployment template at, 0x3bfdED457D9F230a91e6735771113478791C09C1, gas used: 3390937
Deployed ATokenBurner with proxy at 0x75E39fEEa676DBAe11Da1Ab5452510c04eC35FA2, gasCost: 440273
Deployed ATokenMinter with proxy at 0xb95e07ae46Fa2274cAd21b989b00FC0913749403, gasCost: 437437
Deployed Metamorphic for BToken (ALCB) at: 0x3e4b55722e041c0FC55A90B6EBa50EA5d3f7c02A, deployment template at, 0xdC87dEBA768b39eC371ff9C1D8f4D364697F15C6, gas used: 5387764
Deployed Distribution with proxy at 0x0200855f9FdB1bE7aB1A945D65d1cF35eDa0Bed2, gasCost: 634552
Deployed Dynamics with proxy at 0x7594465457d6b870039543345210A602E9f2d299, gasCost: 2279490
Deployed Foundation with proxy at 0x0b094A44994b28f21a7068BA104C192643B42439, gasCost: 771863
Deployed Governance with proxy at 0x9cd25e05b55BD08B22c4302D542a30BD5273aB97, gasCost: 322946
Deployed InvalidTxConsumptionAccusation with proxy at 0x2C37fAC6241D6A96106E2777820C7EEd4Bef3776, gasCost: 2964754
Deployed LiquidityProviderStaking with proxy at 0xE426557C98f5A75bcBA6A8a677e6848E81a4e833, gasCost: 4698999
Deployed MultipleProposalAccusation with proxy at 0x26AFe9e5536A6102c43fEd6Be329734e70437198, gasCost: 1489418
Deployed PublicStaking with proxy at 0x0F27E1c790d776CE13eF852623E2Bc6AcE1b9d4B, gasCost: 4698711
Deployed Snapshots with proxy at 0x0418C73320d1Bc86fad09f7125250Af3aD2c29eF, gasCost: 3651438
Deployed StakingPositionDescriptor with proxy at 0x70a8944A142A03fC62889B6206c801b171Ade659, gasCost: 1176485
Deployed ValidatorPool with proxy at 0x5D48E7935F895c1dC92c19F160Ef54e5C735EDA8, gasCost: 4847685
Deployed ValidatorStaking with proxy at 0xe37c14a34907f8Fcb99243AaeC4842FBb1ce1965, gasCost: 4845358
Deployed ETHDKGAccusations with proxy at 0x167Eb24B7b51f9322a4A2EfCC849310550BE86cD, gasCost: 4591576
Deployed ETHDKGPhases with proxy at 0xb701ce9AaF414Beb3Ed11422eEa6D1f08c5e995f, gasCost: 3569217
Deployed ETHDKG with proxy at 0x9a757acedE2bD60bd09b57163595078636c2bC36, gasCost: 3565690
total gas used: 55167643

Funding validators
account 0xA98A74404C7a0E8540A48bC35A056773D51af97c now has 100000000000000000000 ether
account 0x6DcD347968C8cb034F83171910f55f928da705d3 now has 100000000000000000000 ether
account 0x838aDD5bda57f4273B1aD43c2560Faa1a51E86a9 now has 100000000000000000000 ether
account 0x7586165AB696c645e747ba45D793546a6415A518 now has 100000000000000000000 ether
account 0x49dF85eFe81c958Ae210accC012EeD30147a2a98 now has 100000000000000000000 ether

Registering Validators
 [
  '0x49dF85eFe81c958Ae210accC012EeD30147a2a98',
  '0x6DcD347968C8cb034F83171910f55f928da705d3',
  '0x7586165AB696c645e747ba45D793546a6415A518',
  '0x838aDD5bda57f4273B1aD43c2560Faa1a51E86a9',
  '0xA98A74404C7a0E8540A48bC35A056773D51af97c'
]

Setting the setMinimumIntervalBetweenSnapshots to 10
```

In case you want to interact with ALCB and ALCA tokens using metamask and the AliceNet wallet, save the addresses of the `AliceNetFactory`, `ALCA` and `ALCB` contracts. In the example above, they are at: `0x0b1F9c2b7bED6Db83295c7B5158E3806d67eC5bc`, `0x4426F1939966427C436c565a4532e41028c60301` and `0x3e4b55722e041c0FC55A90B6EBa50EA5d3f7c02A` respectively.

## Running the validators node

The next step, is to start the AliceNet validator nodes. For each validator, open a new terminal and run the following command:

```shell
./scripts/main.sh validator {# for the validator you want to start}
```

Now, you need to wait for all validators to peer up. In the validator's terminal, check for the following log entry `Peers=24/x/3/0`. This log gives you an idea on how many validator have been discovered in the network for the selected validator. You will see the `x` value increasing overtime. Eg: if you initialized 5 validator the Peers value looks like:

```shell
...
...Peers=24/1/3/0...
...
...Peers=24/2/3/0...
...
...Peers=24/3/3/0..
...
...Peers=24/4/3/0...
```

## Starting ETHDKG

Once all the validator nodes have discovered their peers, the next step is to start the ETHDKG. On a free terminal run the following command:

```shell
./scripts/main.sh ethdkg
```

The above command will start the Distributed Key Generation Process (ETHDKG) on your local ethereum node, so the validators can create the Master Private Key and Master Public Key for validating new AliceNet blocks.

Once you have executed the command above, you should see something similar to this being displayed in the validators console:

```shell
... msg="processing registration" ...
... msg="ETHDKG RegistrationOpened" ...
... msg="Scheduling NewRegisterTask" ...
... msg="Scheduling NewDisputeRegistrationTask" ...
... msg="task is about to start" ...
... msg="Previous BaseFee:17793 GasUsed:0 GasLimit:30000000" ...
... msg="Creating TX with MaximumGasPrice: 1000035586 WEI" ...
... msg="Registering  publicKey (0x027...545) with ETHDKG" ...
... msg="Previous BaseFee:17793 GasUsed:0 GasLimit:30000000" ...
```

The whole process should take around 15 minutes to complete (you can speed up this process by changing some hardhat configurations, see the [Speeding up hardhat blocks](#speeding-up-hardhat-blocks) section for more details). At the end of the process, you should see something similar to this being displayed in the validators console:

```shell
... msg="ProcessValidatorSetCompleted()" ...
... msg="Building ValidatorSet..." ...
... msg="ValidatorMember" ...
... msg="ValidatorMember" ...
... msg="ValidatorMember" ...
... msg="ValidatorMember" ...
... msg="ValidatorMember" ...
... msg="Complete ValidatorSet..." ... Validators="0x49df85efe81c958ae210accc012eed30147a2a98 0x6dcd347968c8cb034f83171910f55f928da705d3,0xa98a74404c7a0e8540a48bc35a056773d51af97c,0x838add5bda57f4273b1ad43c2560faa1a51e86a9,0x7586165ab696c645e747ba45d793546a6415a518"
```

Once it has been completed, the validators will start validating (mining) blocks. At this point, you should see something like this:

```shell
... msg="HighestBlockFinalized: 326 -> 329, HighestBlockProcessed: 326 -> 329" Blk/Rnd=1/1 BlkHsh=42ec..1796 BlkTime=23m37.24s GRCnt=238 Peers=24/4/3/0 TxCt=0
... msg="HighestBlockFinalized: 329 -> 332, HighestBlockProcessed: 329 -> 332" Blk/Rnd=1/1 BlkHsh=42ec..1796 BlkTime=23m37.24s GRCnt=238 Peers=24/4/3/0 TxCt=0
... msg="HighestBlockFinalized: 329 -> 332, HighestBlockProcessed: 329 -> 332" Blk/Rnd=1/2 BlkHsh=42ec..1796 BlkTime=1m44.11s GRCnt=237 Peers=24/4/3/0 TxCt=0
... msg="HighestBlockFinalized: 332 -> 335, HighestBlockProcessed: 332 -> 335" Blk/Rnd=1/2 BlkHsh=42ec..1796 BlkTime=1m44.11s GRCnt=237 Peers=24/4/3/0 TxCt=0
... msg="HighestBlockFinalized: 332 -> 335, HighestBlockProcessed: 332 -> 335" Blk/Rnd=2/1 BlkHsh=287f..ec49 BlkTime=5s GRCnt=237 Peers=24/4/3/0 TxCt=0
... msg="HighestBlockFinalized: 335 -> 338, HighestBlockProcessed: 335 -> 338" Blk/Rnd=2/1 BlkHsh=287f..ec49 BlkTime=5s GRCnt=238 Peers=24/4/3/0 TxCt=0
... msg="HighestBlockFinalized: 335 -> 338, HighestBlockProcessed: 335 -> 338" Blk/Rnd=3/1 BlkHsh=91c1..36f5 BlkTime=5s GRCnt=237 Peers=24/4/3/0 TxCt=0
```

The `Blk/Rnd=3/1` field indicates which is the current block being validated.

Now you should finally be able to play around with the AliceNet environment (e.g do RPC calls against the node, do AliceNet transactions, ALCB deposits).

## Hardhat utility scripts

Inside the bridge folder, you will be able to find some useful hardhat scripts to run some actions against the AliceNet smart contracts. With these scripts, you will be able to change some AliceNet configs, mint some ALCB and ALCA, change the hardhat block time and many more. You can check all scripts available by executing the following command in a terminal inside the bridge folder (`cd ./bridge`):

```shell
npx hardhat--help
```

### Speeding up hardhat blocks

You can speed up the ETHDKG ceremony by changing hardhat mined blocks frequency by executing the following command inside the bridge folder (`cd ./bridge`):

```shell
npx hardhat --network dev setHardhatIntervalMining --interval AMOUNT_IN_MILLISECONDS
```

For a more realistic network, change the value back to the `13000` after ETHDKG has completed.

### Minting ALCA (AToken/AliceNet Staking Tokens)

Inside the bridge folder, you will be able to mint ALCA by running the following command:

```shell
npx hardhat --network dev mintATokenTo --factory-address ALICENET_FACTORY_ADDRESS --amount AMOUNT --to ETHEREUM_ADDRESS
```

For instance, assuming the addresses that we got from the [deployment section](#deploying-the-alicenet-ethereum-smart-contracts-locally):

```shell
npx hardhat --network dev mintATokenTo --factory-address 0x0b1F9c2b7bED6Db83295c7B5158E3806d67eC5bc --amount 1000000000000 --to 0x546F99F244b7B58B855330AE0E2BC1b30b41302F
```

You can check the AToken (ALCA) balance with:

```shell
npx hardhat --network dev getATokenBalance --factory-address ALICENET_FACTORY_ADDRESS --account ETHEREUM_ADDRESS
```

Using the example above:

```shell
npx hardhat --network dev getATokenBalance --factory-address 0x0b1F9c2b7bED6Db83295c7B5158E3806d67eC5bc --account 0x546F99F244b7B58B855330AE0E2BC1b30b41302F
```

You should see something like this on your terminal:

```shell
BigNumber { value: "1000000000000" }
```

### Minting ALCB (AliceNet Utility Tokens)

Inside the bridge folder, you will be able to mint ALCB by running the following command:

```shell
npx hardhat --network dev mintBTokenTo --factory-address ALICENET_FACTORY_ADDRESS --amount MIN_AMOUNT_MINTED_EXPECTED --num-wei AMOUNT_OF_ETHER_TO_BUY_BTOKEN --to ETHEREUM_ADDRESS
```

For instance, assuming the addresses that we got from the [deployment section](#deploying-the-alicenet-ethereum-smart-contracts-locally):

```shell
npx hardhat --network dev mintBTokenTo --factory-address 0x0b1F9c2b7bED6Db83295c7B5158E3806d67eC5bc --amount 1 --to 0x546F99F244b7B58B855330AE0E2BC1b30b41302F --num-wei 5555555555555
```

Now, you can check the BToken (ALCB) balance with:

```shell
npx hardhat --network dev getBTokenBalance --factory-address ALICENET_FACTORY_ADDRESS --account ETHEREUM_ADDRESS
```

Using the example above:

```shell
npx hardhat --network dev getBTokenBalance --factory-address 0x0b1F9c2b7bED6Db83295c7B5158E3806d67eC5bc --account 0x546F99F244b7B58B855330AE0E2BC1b30b41302F
```

You should see something like this on your terminal:

```shell
BigNumber { value: "558374317491666" }
```

### Depositing ALCB into AliceNet

In order to do AliceNet transactions, one have to have deposited BTokens (ALCB) into the AliceNet side chain to pay for the transaction gas. On your local testnet, you can deposit ALCB into the side chain with:

```shell
npx hardhat --network dev virtualMintDeposit --factory-address ALICENET_FACTORY_ADDRESS --account-type ACCOUNT_TYPE --deposit-owner-address ACCOUNT_ADDRESS --deposit-amount AMOUNT
```

The `ACCOUNT_TYPE` will be the type of account (elliptic curve used to generate the address) that you have inside the AliceNet chain. It will be `1` for `secp256k1` addresses and `2` for `BN128` addresses.

Assuming the addresses that we got from the [deployment section](#deploying-the-alicenet-ethereum-smart-contracts-locally), you can a deposit to the main account with:

```shell
npx hardhat --network dev virtualMintDeposit --factory-address 0x0b1F9c2b7bED6Db83295c7B5158E3806d67eC5bc --account-type 1 --deposit-owner-address "0x546F99F244b7B58B855330AE0E2BC1b30b41302F" --deposit-amount 1000000000000000
```

In your terminal, you should see something like this:

```shell
...
depositID: BigNumber { value: "2" },
accountType: 1,
depositor: '0x546F99F244b7B58B855330AE0E2BC1b30b41302F',
amount: BigNumber { value: "1000000000000000" }
```

Now, you can now check the ALCB balance of the `0x546F99F244b7B58B855330AE0E2BC1b30b41302F` inside the AliceNet network using the official AliceNet wallet.

## Connecting Metamask to the local ethereum network

You can connect your Metamask with the local ethereum network for better support on getting the tokens balances. In order to do this, open Metamask on your browser and add the hardhat network:

```toml
url = http://127.0.0.1:8545
chainID = 1337
```

Now, add the following private key (account) for the local ethereum admin account to your Metamask:

```toml
0x6aea45ee1273170fb525da34015e4f20ba39fe792f486ba74020bcacc9badfc1
```

> Important: This private is not secure. Ensure you do not use it on production blockchains, or else you risk losing funds.

Finally, import the ALCA and ALCB tokens to metamask. You can get the address in the output of your [deployment script](#deploying-the-alicenet-ethereum-smart-contracts-locally).

## Connecting AliceNet Wallet

Check out the [Wallet documentation](https://github.com/alicenet/wallet) for more information on how to connect your wallet to your local AliceNet network.

## API calls

The AliceNet nodes will be listening to RPC requests in the address defined by the field `localStateListeningAddress` in the configuration file. The AliceNet binary comes with a swagger implementation by default to make the experimentation with the API more user-friendly.

In order to access the swagger page, open one of the validator config file located at `./scripts/generated/config/` and grab the port defined at `localStateListeningAddress`. Now, open a new browser window and paste `http://localhost:<PORT_NUMBER>/swagger`.

For instance, the node running the validator 1 will be usually listening at: `http://localhost:8884/swagger`.
