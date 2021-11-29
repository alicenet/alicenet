#!/bin/sh
timestamp=$(date "+%Y%m%d-%H:%M")
./madnet --config ./assets/config/validator0.toml validator 2>&1 | tee "validator0.$timestamp.log"
