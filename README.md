# AliceNet 2.0

![GitHub Repo stars](https://img.shields.io/github/stars/alicenet/alicenet?style=social)
<!-- [![Discord](https://img.shields.io/badge/Discord-7289DA?logo=discord&logoColor=white)](https://discord.gg/8esBWyr8h3) -->

AliceNet 2.0 is a cutting-edge blockchain built using the **Polygon Chain Development Kit (CDK)**. Operating as a **Validium**, AliceNet 2.0 leverages **Polygon zkEVM** and a **Data Availability Committee (DAC)** to deliver a secure, scalable, and low-cost blockchain solution.

<!-- To learn more, check out our official [website](https://www.alice.net/) and join our official [Discord community](https://discord.gg/8esBWyr8h3). -->

## Key Features

- **Polygon CDK Validium**: A unique setup that uses validity proofs for state transitions without storing transaction data on Ethereum, significantly reducing gas fees.
- **Zero-Knowledge Proofs**: Ensures transaction integrity while maintaining privacy and trustlessness.
- **Data Availability Committee (DAC)**: Manages off-chain transaction data and ensures data availability through authenticated hashes.
- **Dual Token System**: AliceNet 2.0 uses two tokens:
  - **ALCB**: The native ERC-20 token used as the gas token on Layer 2. It can be purchased directly with Ether.
  - **ALCA**: The staking token, which enables users to stake and earn rewards, incentivizing network participation and stability.

For more details about the Polygon CDK, visit the [Polygon CDK Overview](https://docs.polygon.technology/cdk/overview/).

---

## Security

AliceNet 2.0 employs a robust governance structure and advanced security measures to ensure the network's reliability and integrity. Key roles and mechanisms include:

### Multisig Smart Contract

The multisig wallet acts as the primary administrative authority, requiring multiple owner approvals to execute transactions. This ensures decentralized decision-making and prevents unilateral actions. The multisig oversees updates to Polygon CDK contracts, manages parameters on Layer 1, and maintains control of legacy systems like the AliceNet factory. The accounts with multisig authority are carefully selected to represent the network's interests and ensure responsible governance.

### Timelock Contract

The timelock contract enforces a mandatory delay for updates to Polygon CDK contracts, allowing time for review and validation before execution. This mechanism ensures transparency to the users and prevents unauthorized or rushed changes, further safeguarding the network.

### Emergency Consul

The Emergency Consul role, handled by the multisig wallet, enables swift action to halt the system in emergencies. This protects the network from unforeseen threats while maintaining decentralized oversight.

---

## Smart Contracts

All AliceNet 2.0 smart contracts are deployed on the Ethereum mainnet and the Sepolia testnet. Check the [AliceNet Contract Addresses](docs/contracts.md) for more details and the `./src/smart-contracts` folder for the contract code.

---

## Adding AliceNet to MetaMask

### Testnet Configuration

- **Network Name**: AliceNet Testnet
- **RPC URL**: `https://api.staging.alice.net/`
- **Chain ID**: `10042`
- **Currency Symbol**: `ALCB`
- **Block Explorer**: [https://block-explorer.staging.alice.net/](https://block-explorer.staging.alice.net/)

### Mainnet Configuration

- **Network Name**: AliceNet Mainnet
- **RPC URL**: `https://api.prod.alice.net/`
- **Chain ID**: `1042`
- **Currency Symbol**: `ALCB`
- **Block Explorer**: [https://block-explorer.prod.alice.net/](https://block-explorer.prod.alice.net/)

---
