# AliceNet 2.0

<!-- ![GitHub Repo stars](https://img.shields.io/github/stars/alicenet/alicenet?style=social) -->
<!-- [![Discord](https://img.shields.io/badge/Discord-7289DA?logo=discord&logoColor=white)](https://discord.gg/8esBWyr8h3) -->

AliceNet 2.0 is a cutting-edge blockchain built using the
[**Agglayer Chain Development Kit (CDK)**](https://docs.agglayer.dev/cdk/).
Operating as a [**Validium**](https://docs.agglayer.dev/cdk/cdk-erigon/architecture/#cdk-erigon-validium) node,
AliceNet 2.0 leverages
[**Agglayer zkEVM**](https://github.com/agglayer/agglayer-contracts)
and a
[**Data Availability Committee (DAC)**](https://github.com/0xPolygon/cdk-data-availability)
to deliver a secure, scalable, and low-cost blockchain solution.

We note that the Agglayer zkEVM is a continuation of the Polygon zkEVM.

To learn more, check out our official [website](https://app.alice.net/)
and join our official [Discord community](https://discord.gg/RQ3yQ5hwYz).

## Key Features

- **Agglayer CDK Validium**: A unique setup that uses validity proofs
  based on Agglayer zkEVM for state transitions without storing transaction data
  on Ethereum, significantly reducing gas fees.
- **Zero-Knowledge Proofs**: Ensures transaction integrity while maintaining trustlessness.
- **Data Availability Committee (DAC)**: Manages off-chain transaction data and ensures data availability through authenticated hashes.
- **Dual Token System**: AliceNet 2.0 uses two tokens:
  - **ALCA**: The governance token, which allows users to lock and earn rewards, incentivizes network participation and stability.
  - **ALCB**: The native ERC-20 token used as the gas token on Layer 2;
    it can be purchased directly on Ethereum with *ETH*.
    By design, ALCB may only be transferred from Layer 1 to the Layer 2;
    it is not possible to transfer Layer 2 ALCB back to Layer 1.
    ALCB is a utility token with its utility use only available on AliceNet.

### Links to In-Depth Discussions

- For more details about the Agglayer CDK, visit the [Agglayer CDK Overview](https://docs.agglayer.dev/cdk/get-started/overview/).
- The stability of AliceNet 2.0 relies on the two-token system
  of ALCA and ALCB;
  we have separate discussions on
  ALCA ([Governance Token](./docs/alca_discussion.md)) and
  ALCB ([Utility Token](./docs/alcb_discussion.md)).
- Some projects on AliceNet 2.0 use Decentralized Identifiers,
  Verifiable Credentials, and Verifiable Presentations.
  These are based on the [W3C DID Specification](https://www.w3.org/TR/did-core/)
  and [W3C VC Specification](https://w3c.github.io/vc-data-model/).
  An extended discussion on DIDs, VCs, and VPs on AliceNet 2.0 may be found [here](./docs/w3c_discussion.md).
- The security of Verifiable Credentials and Verifiable Presentations
  are based on BBS signatures and NIZK BBS proofs
  (which allow for selective disclosure).
  More information on BBS signatures and NIZK BBS proofs
  may be found [here](./docs/bbs_discussion.md).

---

## Security

AliceNet 2.0 employs a robust governance structure and advanced security measures to ensure the network's reliability and integrity. Key roles and mechanisms include:

### Multisig Smart Contract

The multisig wallet acts as the primary administrative authority, requiring multiple owner approvals to execute transactions.
This ensures decentralized decision-making and prevents unilateral actions.
The multisig oversees updates to Agglayer CDK contracts, manages parameters on Layer 1, and maintains control of legacy systems like the AliceNet factory.
The accounts with multisig authority are carefully selected to represent the network's interests and ensure responsible governance.

### Timelock Contract

The timelock contract enforces a mandatory delay for updates to Agglayer CDK contracts, allowing time for review and validation before execution.
This mechanism ensures transparency to the users and prevents unauthorized or rushed changes, further safeguarding the network.

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
