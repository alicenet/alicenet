# Running an AliceNet node

Before running an AliceNet node, you must create and configure a `config.toml` file, so the node can be executed in the chosen network.


## Creating a configuration file


```toml

```

## AliceNet Networks

Right now, there are 2 AliceNet networks being run. AliceNet main (anchored on ethereum mainnet) and AliceNet testnet (anchored on goerli ethereum testnet).

| AliceNet Network | Ethereum Network | ChainID | Bootnode address                                                                               | Ethereum Smart Contract Factory Address    | Starting Block |
| ---------------- | ---------------- | ------- | ---------------------------------------------------------------------------------------------- | ------------------------------------------ | -------------- |
| main             | mainnet          | 21      | 00000015\|036a4dbb1f572de1f7ae090514c5487b943177c7919504d9e047a0d5bf265eefa1@34.132.49.25:4242 | 0x0000000000000000000000000000000000000000 | 00000000       |
| testnet          | goerli           | 42      | 00000015\|036a4dbb1f572de1f7ae090514c5487b943177c7919504d9e047a0d5bf265eefa1@34.132.49.25:4242 | 0x0000000000000000000000000000000000000000 | 00000000       |


## Running the node

Finally, to run your node, open a new terminal and execute:

```shell
<PATH>/alicenet --config <PATH>/config.toml validator
```

E.g If you copied and pasted the config

```shell
./alicenet --config ./config.toml validator
```
