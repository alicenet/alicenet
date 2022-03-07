# Madnet/Bridge

This repository contains all solidity smart contracts used by the MadNet. 

## Requirements
* [dapp tools](https://github.com/dapphub/dapptools#installation)
* HardHat (following)

### HardHat

To install hardhat and all requirements to compile and test the smart contracts in this repository,
run the following command at the root of this repository

```bash
$ npm i -g hardhat-shorthand
$ hardhat-completion install
```


## Setup

### Install dependencies 

Install all the necessary dependencies
```bash
$ npm ci
$ npm run init 
```
It is mandatory to run this latest command `npm run init`, every time you run the `clean` command.  

### Running unit tests

You can choose between two different commands to run unit tests. The `test-parallel` will be faster
but you won't see the gas reporter as logged once the tests are running successfully 
`Note: Gas reporting has been skipped because plugin hardhat-gas-reporter does not support the --parallel flag.)`

```bash
$ npm run test 
$ npm run test-parallel 
```

### Deploying

Once you reach this point you will now be able to compile and deploy all the contracts. In case you have
new contracts **MAKE SURE** to add them in the configuration file `hardha.config.ts` under `"abiExporter" -> "only"` array.

```bash
$ npm run build
```


### Resources

**Sintax Guidelines** - We follow the standard coding convention for writing solodity code. You can find
the documentation [here](https://docs.soliditylang.org/en/v0.8.9/style-guide.html).

**Formatting** - there is a command, `npm run format` to prettify the codebase based on 
the configuration file `prettier.json`

**Visual Studio Code Plugin** - Group of pluing to make your life easier if you are using 
Visual studio
* Solidity Visual Auditor - [plugin link](https://marketplace.visualstudio.com/items?itemName=tintinweb.solidity-visual-auditor)
* Solidity - [plugin link](https://marketplace.visualstudio.com/items?itemName=JuanBlanco.solidity)
* Remix - [plugin link](https://marketplace.visualstudio.com/items?itemName=RemixProject.ethereum-remix)

**Diamond Smart Contract pattern** - The main architecture pattern used by the MadNet smart contracts 
available [here](https://eips.ethereum.org/EIPS/eip-2535).
