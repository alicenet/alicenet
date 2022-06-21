#!/bin/sh

for addr in $(ls ./scripts/generated/keystores/keys | xargs); do
	./alicenet --config ./scripts/base-files/owner.toml utils approvetokens $addr 2000000 &
done

wait
