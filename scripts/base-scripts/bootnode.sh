#!/bin/sh

CONFIG=${1:-./scripts/base-files/bootnode.toml}

./madnet --config "$CONFIG" bootnode