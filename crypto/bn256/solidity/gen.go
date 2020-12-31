package solidity

//go:generate solc -o contract/ --overwrite --optimize --abi --bin ./contract/crypto.sol
//go:generate abigen --abi=./contract/crypto.abi --bin=./contract/crypto.bin --pkg solidity --out=crypto.go
