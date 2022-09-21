# AliceNet/Bridge

This repository contains all solidity smart contracts used by the AliceNet.

## Requirements

- HardHat (following)

To install hardhat and all requirements to compile and test the smart contracts in this repository, run the following
command at the root of this repository. It might require sudo permission

```bash
$ npm i -g hardhat-shorthand
$ hardhat-completion install
```

## Setup

### Install dependencies

Install all the necessary dependencies, and compile.

```bash
$ npm ci
$ npm run compile
$ npm run generate
```

### Running unit tests

You can choose between two different commands to run unit tests. The `test-parallel` will be faster but you won't see
the gas reporter when the tests are finished
`Note: Gas reporting has been skipped because plugin hardhat-gas-reporter does not support the --parallel flag.)`
The two commands are respectivetely

```bash
$ npm run test
```

and

```bash
$ npm run test-parallel
```

## HARDHAT

There is a suite of hardhat scripts and tasks you can run from the bridge folder. To list them simply
run `npx hardhat --help`.

- `generate-immutable-auth-contract` Generates authorization for contracts deployed at deterministic address via factory

```bash

 $ npx hardhat generate-immutable-auth-contract
 $ npx hardhat generate-immutable-auth-contract --input ./path/to/folder/containing/deploymentList
 $ npx hardhat generate-immutable-auth-contract --output ./
 $ npx hardhat generate-immutable-auth-contract --input ./path/to/folder/containing/deploymentList --output ./
```

If you are linking a custom Deployment list file it must follow the following format

```toml
 deploymentList = [
    "path/to/contract/ContractName.sol:ContractName",
    ...
 ]
```

### Golang code binding generation

Once you reach this point you will now be able to compile and generate code bindings for all the contracts. In case you
have new contracts **MAKE SURE** to add them in the configuration file `hardhat.config.ts`
under `"abiExporter" -> "only"` array

```bash
$ npm run build
```

# TROUBLESHOOTING

In the case there is any problem compiling the contracts, run the following command to clean up your environment and try
again.

```bash
$ npm run clean
```

## Resources

**Formatting** - there is a command, `npm run format` to prettify the codebase based on the configuration
file `prettier.json`

**Visual Studio Code Plugin** - Group of pluing to make your life easier if you are using Visual studio

- Solidity Visual Auditor
  - [plugin link](https://marketplace.visualstudio.com/items?itemName=tintinweb.solidity-visual-auditor)
- Solidity - [plugin link](https://marketplace.visualstudio.com/items?itemName=JuanBlanco.solidity)
- Remix - [plugin link](https://marketplace.visualstudio.com/items?itemName=RemixProject.ethereum-remix)

## Solidity Style guide

We follow the standard coding convention for writing solidity code. You can find the
documentation [here](https://docs.soliditylang.org/en/v0.8.9/style-guide.html).
