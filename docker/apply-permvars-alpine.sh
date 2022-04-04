#!/bin/sh

# This file replicates all the permissions that the host OS saved with `set-permvars.sh` to an alpine container
# Paste the following lines into the Dockerfile to use it:
#
#   ARG BUILDER_UID
#   ARG BUILDER_GIDS
#   ADD docker/apply-permvars-alpine.sh /
#   RUN /apply-permvars-alpine.sh
#   USER $BUILDER_UID
#

# skip errors, because some groups, and perhaps even the UID may already exist
# this isn't a problem, only the UID or GID matters, not the actual name of the user or group
set +e

# create user if doesn't exist already
adduser -D -g $BUILDER_UID container_builder
BUILDER_NAME=$(getent passwd $BUILDER_UID | cut -d: -f1)

# add all group ids that the user belongs to if they don't exist already
echo $BUILDER_GIDS | xargs -n 1 | xargs -I % addgroup -g % group%

# assign user to all group ids
echo $BUILDER_GIDS | xargs -n 1 | xargs -I % getent group % | cut -d: -f1 | xargs -I % adduser $BUILDER_NAME %

echo $BUILDER_NAME