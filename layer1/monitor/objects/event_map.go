package objects

import (
	"sync"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"

	"github.com/alicenet/alicenet/layer1"
)

type EventProcessor func(eth layer1.Client, contracts layer1.AllSmartContracts, logger *logrus.Entry, state *MonitorState, log types.Log) error

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

func (em *EventMap) registerLocked(ID, name string, fn EventProcessor) error {
	em.registry[ID] = &EventInformation{Processor: fn, Name: name}

	return nil
}

func (em *EventMap) Register(ID, name string, fn EventProcessor) error {
	em.Lock()
	defer em.Unlock()

	err := em.registerLocked(ID, name, fn)
	if err != nil {
		return err
	}

	return nil
}

func (em *EventMap) Lookup(ID string) (*EventInformation, bool) {
	em.RLock()
	defer em.RUnlock()

	info, present := em.registry[ID]

	return info, present
}
