package accusation

import (
	"sync"
	"time"

	"github.com/MadBase/MadNet/consensus/objs"
)

// Manager polls validators' roundStates and forwards them to a Detector. Also handles detected accusations.
type Manager struct {
	sync.Mutex
	detector *Detector
}

func NewManager() *Manager {
	// todo: populate detector functions here
	detector := NewDetector(make([]func(rs *objs.RoundState) (*Accusation, bool), 0))
	m := &Manager{
		detector: detector,
	}

	go m.run()
	return m
}

func (m *Manager) run() {
	for {
		rss, err := m.pollRS()
		if err != nil {
			panic("AccusationManager could not poll roundStates")
		}

		for _, rs := range rss {
			// send round states to detector to be processed
			m.detector.HandleRS(rs)
		}

		time.Sleep(1 * time.Second)
	}
}

func (m *Manager) pollRS() ([]*objs.RoundState, error) {
	// todo: read from DB to get validators' round states
	return make([]*objs.RoundState, 0), nil
}
