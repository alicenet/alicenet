#!/bin/sh

for addr in $(ls ./scripts/generated/keystores/keys | xargs); do

./madnet --config ./scripts/base-files/owner.toml --ethereum.defaultAccount $addr utils transfertokens 0x546F99F244b7B58B855330AE0E2BC1b30b41302F 2000000 &

done

wait
