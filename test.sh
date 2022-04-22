#!/bin/bash

set -ex


sed -n '/abiExporter:/,/}/p' ./hardhat.config.ts | \
sed -n '/only:/,/]/p' | \
sed 's/only://' | \
sed 's/[][,"]//g' | \
xargs -n 1 -I {} echo {}".sol" | tr '\n' ' ' > ./myfile.txt
IFS=$' '
rm -rf ./myhash.txt
for i in $(cat ./myfile.txt)
do
    find ./contracts -type f -iname "$i" -print -exec cat {} \; | sha256sum | cut -d' ' -f1 >> ./myhash.txt
done
cat ./myhash.txt | sha256sum | cut -d' ' -f1