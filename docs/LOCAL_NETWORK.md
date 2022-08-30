# Running a full local AliceNet network

For testing purposes some times is useful to start a full local AliceNet network. In order to do so, you will need to run an ethereum test node, deploy all smart contract infra-structure, run all the validators and bootnode locally. This documentation will guide you step by step on how to run your own local AliceNet network.

## Requirements

In order to run the test suite, you must have followed and completed the [Building from Source](./BUILD.md) instructions.

## Setup Test Environment

**[OPTIONAL] If already executed this step before, you may want to run the following command first, to clean any leftover files from a previous execution.**

```shell
./scripts/main.sh clean
```

In order to initialize a chain, we need to create all configuration files first. To do this, open a new terminal and them run the following command:

```shell
./scripts/main.sh init {# number of validators 4-32}
```

The command above will generate the configuration files, keystores, genesis file, and password files inside: `./scripts/generated`.

## Running a bootnode

In sequence, we need to run an AliceNet bootnode client. A bootnode is special AliceNet client that responsible for facilitating the node peering up in the network. Once a normal node starts, it reaches to a bootnode defined in the `config.toml` file to get information about other peers.

To run a bootnode, execute the following command:

```shell
./scripts/main.sh bootnode
```

## Running an ethereum test node

The next step is to run an ethereum test node so we can configure the AliceNet smart contract intra-structure.

To start the ethereum test node, just open a new terminal and run:

```shell
./scripts/main.sh hardhat
```

The above command will start a [hardhat node](https://hardhat.org/hardhat-network/docs/overview) locally, listening at `http://127.0.0.1:8545`

## Deploying the AliceNet Ethereum smart contracts locally

In order to deploy the AliceNet smart contract on your local ethereum node, open now another terminal and run the command bellow.

```shell
./scripts/main.sh deploy
```

The command will also register validators created by the `init` command and setup some AliceNet configurations in the smart contracts. After the command is executed, you should see something like this on your terminal:

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

In case you want to interact with ALCB and ALCA tokens using metamask and the AliceNet wallet, save the addresses of the `AliceNetFactory`, `ALCA` and `ALCB` contracts. In the example above, they are: `0x0b1F9c2b7bED6Db83295c7B5158E3806d67eC5bc`, `0x4426F1939966427C436c565a4532e41028c60301` and `0x3e4b55722e041c0FC55A90B6EBa50EA5d3f7c02A` respectively.

## Running the validators node

In sequence, we need to start the AliceNet validator nodes. For each validator, open a new terminal and run the following command:

```shell
./scripts/main.sh validator {# for the validator you want to start}
```

Now, you need to wait for all validators to peer up. In the validator's terminal, check for the following log `Peers=24/x/3/0`. This log gives you an idea on how many validator have been discovered in the network for the selected validator. You will see the `x` value increasing overtime. Eg if you initialized 5 validator the Peers value looks like:

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

## Starting ETHDKG

Once all the validators node have discovered their peers, we wil need to start the ETHDKG. On a free terminal run the following command:

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

The whole process should take around 15 minutes to complete. At the end of the process, you should see something similar to this being displayed in the validators console:

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


Once it has been completed, the validators should start validating (mining) blocks. At this point, you should see something like this:

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

Now you should be able to play around with the AliceNet environment (e.g do RPC calls against the node, do AliceNet transactions, ALCB deposits).

## Minting ALCA (AliceNet Staking Tokens)

## Minting ALCB (AliceNet Utility Tokens)

## Depositing ALCB into AliceNet

## Connecting Metamask to the local ethereum network

## Connecting AliceNet Wallet
## Doing AliceNet transactions

Deposits are required in order to submit DataStores. Run the following at least 4 times in order to deposit enough funds
to inject datastores.

```
./scripts/main.sh deposit
```

## API calls

The swagger-ui for the localRPC of a validator may be found at `http://localhost:8885/swagger`. The default port for
`validator4` is `8888`. Thus, you may speak to validator 4 (a non-mining node) at `http://localhost:8888`


