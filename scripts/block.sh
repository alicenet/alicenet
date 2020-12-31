#!/bin/sh

txnCount=1

if [ "$1" != "" ]; then
	txnCount=$1
fi

for i in $(seq 1 $txnCount); do
	./madnet --config ./assets/config/owner.toml --logging ethereum=error,utils=info --ethereum.registryAddress 0xea2d8fa1b956a25479c7808c078b2c58a02b279c utils approvetokens 0x9AC1c9afBAec85278679fF75Ef109217f26b1417 1 >/dev/null
done
