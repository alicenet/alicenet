#!/bin/sh

cat <<EOF | while read addr; do ./madnet --config ./assets/config/owner.toml --ethereum.defaultAccount $addr utils unregister; done
0xde2328ce51aab2087bc60b46e00f4ce587c7a8a9
0x44a9ce0afd70ccae70b8ab5b6772e5ed8d695ea7
0x7fae97e8ef6abc96456b60bd6e97523e4c6fc9a2
0x23ea3bad9115d436190851cf4c49c1032fa7579a
0x33d0141b5647d554c5481204473fd81850f2970d
0xeba70dc631ea59a2201ee0b3213dca1549ffab48
EOF
