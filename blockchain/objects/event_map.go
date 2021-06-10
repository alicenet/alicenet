package objects

import (
	"sync"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

type EventProcessor func(logger *logrus.Logger, state *MonitorState, log types.Log) error

type EventInformation struct {
	Name      string
	Processor EventProcessor
}

type EventMap struct {
	sync.RWMutex
	registry map[string]*EventInformation
}

func NewEventMap() *EventMap {
	return &EventMap{registry: make(map[string]*EventInformation)}
}

func (em *EventMap) RegisterLocked(ID string, name string, fn EventProcessor) error {

	em.registry[ID] = &EventInformation{Processor: fn, Name: name}

	return nil
}

func (em *EventMap) Register(ID string, name string, fn EventProcessor) error {
	em.Lock()
	defer em.Unlock()

	em.RegisterLocked(ID, name, fn)

	return nil
}

func (em *EventMap) Lookup(ID string) (*EventInformation, bool) {
	em.RLock()
	defer em.RUnlock()

	info, present := em.registry[ID]

	return info, present
}
