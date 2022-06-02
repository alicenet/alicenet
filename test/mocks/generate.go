package mocks

// Mocks created from interfaces:
// * Mocks to be used directly:
//go:generate go-mockgen -f -i ITask -o interfaces_executor.mockgen.go --prefix=Base ../../blockchain/executor/interfaces
//go:generate go-mockgen -f -i ITransactionWatcher -o interfaces_txwatcher.mockgen.go ../../blockchain/txwatcher/interfaces
//go:generate go-mockgen -f -i IAdminHandler -o interfaces_monitor.mockgen.go ../../blockchain/monitor/interfaces
//go:generate go-mockgen -f -i IEthereum -i IEthereumClient -i IContracts -o interfaces_ethereum.mockgen.go ../../blockchain/ethereum/interfaces

// Mocks created from bindings:

//go:generate go-mockgen -f -i IETHDKG -i IGovernance -i IAToken -i IBToken -i IAliceNetFactory -i IPublicStaking -i ISnapshots -i IValidatorPool -i IValidatorStaking  -o bindings.mockgen.go ../../bridge/bindings
