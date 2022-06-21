#!/bin/sh

for addr in $(ls ./scripts/generated/keystores/keys | xargs); do
    
    ./alicenet --config ./scripts/base-files/owner.toml --ethereum.defaultAccount $addr utils register &
    
done

wait
