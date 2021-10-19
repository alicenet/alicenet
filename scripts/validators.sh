#!/bin/sh

rm -rf ~/validator?/state
rm -rf ~/validator?/mon

if [ -z "$ENDPOINT" ]; then
    ./madnet --config ./assets/config/validator0.toml validator > validator0.log 2>&1 &
    ./madnet --config ./assets/config/validator1.toml validator > validator1.log 2>&1 &
    ./madnet --config ./assets/config/validator2.toml validator > validator2.log 2>&1 &
    ./madnet --config ./assets/config/validator3.toml validator > validator3.log 2>&1 &
    ./madnet --config ./assets/config/validator4.toml validator > validator4.log 2>&1 &
else
    ./madnet --config ./assets/config/validator0.toml --ethereum.endpoint $ENDPOINT validator > validator0.log 2>&1 &
    ./madnet --config ./assets/config/validator1.toml --ethereum.endpoint $ENDPOINT validator > validator1.log 2>&1 &
    ./madnet --config ./assets/config/validator2.toml --ethereum.endpoint $ENDPOINT validator > validator2.log 2>&1 &
    ./madnet --config ./assets/config/validator3.toml --ethereum.endpoint $ENDPOINT validator > validator3.log 2>&1 &
    ./madnet --config ./assets/config/validator4.toml --ethereum.endpoint $ENDPOINT validator > validator4.log 2>&1 &
fi

wait
