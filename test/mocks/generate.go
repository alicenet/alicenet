package mocks

// Mocks created from interfaces:
// * Mocks to be used directly:
//go:generate go-mockgen -f -i Task -o interfaces_executor.mockgen.go ../../layer1/executor/tasks
//go:generate go-mockgen -f -i Watcher -i ReceiptResponse -o interfaces_transaction.mockgen.go ../../layer1/transaction
//go:generate go-mockgen -f -i AdminHandler -o interfaces_monitor.mockgen.go ../../layer1/monitor/interfaces
//go:generate go-mockgen -f -i Client -o interfaces_ethereum.mockgen.go ../../layer1

// Mocks created from bindings:

//go:generate go-mockgen -f -i IETHDKG -i IGovernance -i IAToken -i IBToken -i IAliceNetFactory -i IPublicStaking -i ISnapshots -i IValidatorPool -i IValidatorStaking  -o bindings.mockgen.go ../../bridge/bindings
