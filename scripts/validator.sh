#!/bin/sh

if [ "$#" -ne 1 ] || ! [ -f "$1" ]; then
    echo "Usage: $0 CONFIG" >&2
    exit 1
fi

CONFIG=${1:-./assets/config/validator0.toml}

./madnet --config "${CONFIG}" validator
