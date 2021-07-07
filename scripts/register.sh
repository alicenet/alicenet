#!/bin/sh

for addr in `cat <<EOF 
0x9AC1c9afBAec85278679fF75Ef109217f26b1417
0x615695C4a4D6a60830e5fca4901FbA099DF26271
0x63a6627b79813A7A43829490C4cE409254f64177
0x16564cF3e880d9F5d09909F51b922941EbBbC24d
0x546F99F244b7B58B855330AE0E2BC1b30b41302F
EOF`; do

./madnet --config ./assets/config/owner.toml --ethereum.defaultAccount $addr utils register &

done

wait
