package objs

// type AccusationState uint8

// const (
// 	// This means the accusation is not yet initalized and thus not valid.
// 	Unset AccusationState = iota
// 	// Created means the accusation has been identified but not yet persisted in any way.
// 	Created AccusationState = iota
// 	// Persisted means the accusation has been identified and persisted in the consensus DB before being processed
// 	Persisted AccusationState = iota
// 	// ScheduledForExecution means the accusation is persisted in consesus DB and scheduled for execution, or executing, on the task manager. An accusation should only be scheduled for execution when in the Persisted state
// 	ScheduledForExecution AccusationState = iota
// 	// Completed means the accusation has been executed and no action is needed. An accusation should only be completed when in the ScheduledForExecution state
// 	Completed AccusationState = iota
// )

// type Accusation interface {
// 	SubmitToSmartContracts() error
// 	GetID() [32]byte
// 	SetID(id [32]byte)
// 	GetPersistenceTimestamp() uint64
// 	SetPersistenceTimestamp(timestamp uint64)
// 	GetState() AccusationState
// 	SetState(state AccusationState)
// }

// var ErrNotImpl = errors.New("not implemented")

// type BaseAccusation struct {
// 	ID                   [32]byte
// 	PersistenceTimestamp uint64
// 	State                AccusationState
// }

// func (a *BaseAccusation) SubmitToSmartContracts() error {
// 	return ErrNotImpl
// }

// func (a *BaseAccusation) GetID() [32]byte {
// 	return a.ID
// }

// func (a *BaseAccusation) SetID(id [32]byte) {
// 	a.ID = id
// }

// func (a *BaseAccusation) GetPersistenceTimestamp() uint64 {
// 	return a.PersistenceTimestamp
// }

// func (a *BaseAccusation) SetPersistenceTimestamp(timestamp uint64) {
// 	a.PersistenceTimestamp = timestamp
// }

// func (a *BaseAccusation) GetState() AccusationState {
// 	return a.State
// }

// func (a *BaseAccusation) SetState(state AccusationState) {
// 	a.State = state
// }

// var _ Accusation = &BaseAccusation{}
