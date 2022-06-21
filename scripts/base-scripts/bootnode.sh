#!/bin/sh

CONFIG=${1:-./scripts/base-files/bootnode.toml}

./alicenet --config "$CONFIG" bootnode
