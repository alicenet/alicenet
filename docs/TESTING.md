# Testing

This documentation will guide you through how to run the AliceNet test suite. AliceNet is basically split into two different set of unit tests: solidity unit tests and golang unit tests.

## Requirements

In order to run the test suite, you must have followed and completed the [Building from Source](./BUILD.md) instructions.

## Solidity

To run all the smart contracts tests, you have to be inside the bridge folder. Open a new terminal in the root, and step in the bridge folder with:

```shell
cd bridge
```

Now, you can choose between two different commands to run unit tests.

The two commands are respectively

```bash
$ npm run test
```

and

```bash
$ npm run test-parallel
```

We recommend using the `test-parallel` since it will be faster but as trade-off you won't see the gas cost reporter when the tests are finished.

## Golang

The golang tests suite is split into two different sets of unit tests: normal and integration tests.

To run all normal golang unit tests, open a new terminal in the repository root, then run the following command:

```shell
go test -race -v ./...
```

To run the golang integration tests, open a new terminal in the repository root, then run the following command:

```shell
go test -race -tags=integration -timeout=30m -v github.com/alicenet/alicenet/layer1/...
```

During the integration tests, if you want to have even more verbose output, you can set the env variable `ENABLE_SCRIPT_LOG` to `true`.

```shell
ENABLE_SCRIPT_LOG=true go test -race -tags=integration -timeout=30m -v github.com/alicenet/alicenet/layer1/...
```
