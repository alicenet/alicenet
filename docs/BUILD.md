# Building from Source

## Prerequisites

- [Docker v20 with docker-compose](https://docs.docker.com/get-docker)
- [Go 1.19](https://go.dev/dl/)
- [Geth 1.10.22](https://geth.ethereum.org/docs/install-and-build/installing-geth)
- [Node 16](https://nodejs.org/en/download/)

Before proceeding, make sure that you have your `GOPATH` in your `PATH`. You can do this by adding the following line to your terminal configuration file (e.g $HOME/.bashrc, $HOME/.zshrc):

```shell
export PATH="$PATH:$(go env GOPATH)/bin"
```

Note: Changes made to a configuration file may not apply until you restart your terminal.

## Clone the Repository

Once you have installed all pre-requisites, clone the AliceNet repository.

```shell
git clone --recursive https://github.com/alicenet/alicenet.git
cd alicenet
```

If you want to help to develop AliceNet, you should fork AliceNet. See our [Contribution Guide Lines](../CONTRIBUTING.md) for more details.

## Installing dependencies

Once you have cloned AliceNet and have installed all requisites, open a new terminal window in the alicenet folder and run:

```shell
make setup
```

The above command will install all AliceNet dependencies necessary for building the AliceNet binary.

**Note**:

To simplify local (or remote via [CodeSpaces](https://github.com/features/codespaces)) development, a [devcontainer](https://code.visualstudio.com/docs/remote/containers) is provided. The remainder of the readme should work without additional configuration from within the devcontainer.

## [OPTIONAL] File generation

If you are developing a new functionality for AliceNet and you are implementing changes to:

- The solidity contracts and want these to be used by the AliceNet binary.
- The public API surface of AliceNet.

You will need to run an additional command before building the binary. In the AliceNet root run the following command:

- `make generate`

This command:

- (re)compiles all the protobuf type definitions into Go sourcecode
- (re)compiles all the capnproto type definitions into Go sourcecode
- (re)compiles all the grpc endpoint definitions into Go sourcecode
- (re)generates convenient wrapper functions for the grpc endpoints using `cmd/mngen`
- (re)generates a new swagger json file based on the grpc endpoints
- (re)generates a Go source file containing the swagger json file in binary format, so it can be baked into the final executable
- (re)compiles all the solidity contracts
- (re)compiles the Go bindings for the solidity contracts
- (re)generates the ABI definitions for the solidity contracts

## Build AliceNet

Finally, the following command will build the AliceNet binary for you:

```shell
make build
```

## What's next

Once you have an AliceNet binary, see the documentation [How to configure your node](CONFIGURE.md) to run your own node against one of AliceNet's networks.

Or, check the [Testing Documentation](./TESTING.md) to see how to run the AliceNet tests.
