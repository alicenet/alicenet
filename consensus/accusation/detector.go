package accusation

import (
	"fmt"
	"sync"
	"time"

	"github.com/MadBase/MadNet/consensus/objs"
)

// a function that returns an Accusation interface object, and a bool indicating if an accusation has been found (true) or not (false)
type detectorF = func(rs *objs.RoundState) (*Accusation, bool)

// Detector detects malicious behaviours in round state objects
type Detector struct {
	sync.Mutex
	detectors      []detectorF
	queueToProcess []*objs.RoundState
}

func NewDetector(detectors []detectorF) *Detector {
	d := &Detector{
		detectors: detectors,
	}

	go d.run()
	return d
}

// HandleRS adds a round state to the queue to be processed
func (d *Detector) HandleRS(rs *objs.RoundState) {
	d.Lock()
	defer d.Unlock()
	d.queueToProcess = append(d.queueToProcess, rs)
}

func (d *Detector) run() {
	for {
		time.Sleep(1 * time.Second)
		d.Lock()

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

		for _, detector := range d.detectors {
			accusation, found := detector(rs)
			if found {
				fmt.Printf("Detector: found accusation: %#v\n", accusation)
				// todo: block until we get a response from the accusation
				// (*accusation).SubmitToSmartContracts()
			}
		}
	}
}
