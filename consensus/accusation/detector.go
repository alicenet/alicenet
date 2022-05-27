package accusation

import (
	"fmt"
	"sync"
	"time"

	"github.com/MadBase/MadNet/consensus/objs"
)

// a function that returns an Accusation interface object, and a bool indicating if an accusation has been found (true) or not (false)
type detectorLogic = func(rs *objs.RoundState) (*Accusation, bool)

// Detector detects malicious behaviours in round state objects
type Detector struct {
	sync.Mutex
	manager            *Manager
	processingPipeline []detectorLogic
	queueToProcess     []*objs.RoundState
}

func NewDetector(manager *Manager, detectors []detectorLogic) *Detector {
	d := &Detector{
		manager:            manager,
		processingPipeline: detectors,
	}

	return d
}

func (d *Detector) Start() {
	go d.run()
}

// HandleRoundState adds a round state to the queue to be processed
func (d *Detector) HandleRoundState(rs *objs.RoundState) {
	d.Lock()
	defer d.Unlock()
	d.queueToProcess = append(d.queueToProcess, rs)
}

func (d *Detector) run() {
	for {
		time.Sleep(1 * time.Second)
		d.Lock()

		// only process round states if there's a manager to report to
		if d.manager == nil {
			d.Unlock()
			continue
		}

		if len(d.queueToProcess) <= 0 {
			d.Unlock()
			// should skip if there are not items in the queue
			continue
		}

		// pop first round state from queue
		rs, new_queue := d.queueToProcess[0], d.queueToProcess[1:]
		d.queueToProcess = new_queue

		d.Unlock()

		fmt.Printf("Detector: processing round state %#v\n", rs)

		for _, detector := range d.processingPipeline {
			accusation, found := detector(rs)
			if found {
				fmt.Printf("Detector: found accusation: %#v\n", accusation)
				// todo: send Accusation to Manager
				err := d.manager.HandleAccusation(accusation)
				if err != nil {
					panic(fmt.Sprintf("Detector: could not handle accusation: %v", err))
				}

				// todo: block until we get a response from the accusation
			}
		}
	}
}
