package mocks

// Mocks created from interfaces:
// * Mocks to be used directly:
//go:generate go-mockgen -f -i Watcher -i ReceiptResponse -o interfaces_transaction.mockgen.go ../../layer1/transaction
//go:generate go-mockgen -f -i DepositHandler -i AdminHandler -i AdminClient -o interfaces_monitor.mockgen.go ../../layer1/monitor/interfaces
//go:generate go-mockgen -f -i Client -i AllSmartContracts -i EthereumContracts -o interfaces_ethereum.mockgen.go ../../layer1

// Mocks created from bindings:

//go:generate go-mockgen -f -i IETHDKG -i IGovernance -i IALCA -i IALCB -i IAliceNetFactory -i IPublicStaking -i ISnapshots -i IValidatorPool -i IValidatorStaking -i IDynamics -o bindings.mockgen.go ../../bridge/bindings
