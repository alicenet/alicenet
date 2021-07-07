#!/bin/sh
make build && ./madnet --config ./assets/config/owner.toml --deploy.migrations=true deploy
