#!/bin/sh
make build && ./madnet --config ./scripts/base-files/owner.toml --deploy.migrations=true deploy
