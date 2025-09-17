# AliceNet Smart Contracts

This repository hosts the AliceNet 1.0 smart contracts, which are vital for managing the AliceNet
ecosystem. These contracts facilitate a wide array of functionalities critical for the network’s
operation.

>> This repository also contains a modified copy of the legacy smart contracts used by AliceNet 0.0.11, as some of them are still in use on the network. If you want to access the original AliceNet 0.0.11 smart contracts, please refer to the [AliceNet 0.0.11 repository](https://github.com/alicenet/alicenet/tree/legacy/0.0.11/bridge).

>> In order to fit the AliceNet protocol needs, some of the polygon CDK smart contracts were modified. The modified contracts can be found in the [AliceNet zkevm contracts repository](https://github.com/alicenet/zkevm-contracts).

## Installation

To install the AliceNet smart contracts, follow these steps:

1. **Clone the repository:** Start by cloning this repository to your local environment.
2. **Install dependencies:** Navigate to the cloned repository’s folder and run the following command to install the necessary dependencies:

    ```bash
    npm install --force
    ```

3. **Compile the contracts:** Execute the following command to compile the smart contracts:

    ```bash
    npx hardhat compile
    ```

## Running Tests

To run the tests for the AliceNet smart contracts, execute the following command:

```bash
npm test
```
