#!/bin/sh

# This file saves all the permission information of the current user into bash variables
# These variables can then be passed onto a container, where they can be used to recreate the user ids and groups
# This allows us to seamlessly mount directories into containers, generate files in those containers all under the same user as is used under the host OS

case "$(uname -s)" in
	CYGWIN*|MINGW32*|MSYS*|MINGW*)
		# if windows, permissions don't matter, just set to root
		export BUILDER_UID=0;
		export BUILDER_GIDS=0;
		;;

	*)
		# the user id of the person building the image
		export BUILDER_UID=$(id -u);

		# the ids of all the user groups that the person building the image belongs to
		export BUILDER_GIDS=$(groups | xargs -n 1 | xargs -I % getent group % | cut -d: -f3);
		;;
esac
