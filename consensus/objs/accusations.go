package objs

import (
	"errors"

	"github.com/google/uuid"
)

type AccusationState uint8

const (
	// Created means the accusation has been identified but not yet persisted in any way.
	Created AccusationState = 0
	// Persisted means the accusation has been identified and persisted in the consensus DB before being processed
	Persisted AccusationState = 1
	// ScheduledForExecution means the accusation is persisted in consesus DB and scheduled for execution, or executing, on the task manager. An accusation should only be scheduled for execution when in the Persisted state
	ScheduledForExecution AccusationState = 2
	// Completed means the accusation has been executed and no action is needed. An accusation should only be completed when in the ScheduledForExecution state
	Completed AccusationState = 3
)

type Accusation interface {
	SubmitToSmartContracts() error
	GetUUID() uuid.UUID
	SetUUID(uuid uuid.UUID)
	GetPersistenceTimestamp() uint64
	SetPersistenceTimestamp(timestamp uint64)
	GetState() AccusationState
	SetState(state AccusationState)
}

var ErrNotImpl = errors.New("not implemented")

type BaseAccusation struct {
	UUID                 uuid.UUID
	PersistenceTimestamp uint64
	State                AccusationState
}

func (a *BaseAccusation) SubmitToSmartContracts() error {
	return ErrNotImpl
}

func (a *BaseAccusation) GetUUID() uuid.UUID {
	return a.UUID
}

func (a *BaseAccusation) SetUUID(uuid uuid.UUID) {
	a.UUID = uuid
}

func (a *BaseAccusation) GetPersistenceTimestamp() uint64 {
	return a.PersistenceTimestamp
}

func (a *BaseAccusation) SetPersistenceTimestamp(timestamp uint64) {
	a.PersistenceTimestamp = timestamp
}

func (a *BaseAccusation) GetState() AccusationState {
	return a.State
}

func (a *BaseAccusation) SetState(state AccusationState) {
	a.State = state
}

var _ Accusation = &BaseAccusation{}
