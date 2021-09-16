#!/bin/sh

CONFIG=${1:-./assets/config/bootnode.toml}

./madnet --config "$CONFIG" bootnode