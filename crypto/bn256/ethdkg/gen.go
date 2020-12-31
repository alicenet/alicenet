package ethdkg

//go:generate solc -o contract/ --overwrite --optimize --abi --bin ./contract/ETHDKG_mod.sol
//go:generate abigen --abi=./contract/ETHDKG.abi --bin=./contract/ETHDKG.bin --pkg ethdkg --out=ETHDKG_mod.go
