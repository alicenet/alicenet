package accusation

import (
	"fmt"
	"time"

	"github.com/MadBase/MadNet/consensus/objs"
)

// a function that returns an Accusation interface object, and a bool indicating if an accusation has been found (true) or not (false)
type detectorF = func(rs *objs.RoundState) (*Accusation, bool)

// Detector detects malicious behaviours in round state objects
type Detector struct {
	detectors   []detectorF
	rsToProcess chan objs.RoundState
	queue       []objs.RoundState
}

func NewDetector(detectors []detectorF) *Detector {
	d := &Detector{
		detectors:   detectors,
		rsToProcess: make(chan objs.RoundState),
	}

	go d.run()
	return d
}

func (d *Detector) HandleRS(rs objs.RoundState) {
	d.queue = append(d.queue, rs)
}

func (d *Detector) run() {
	for {
		time.Sleep(1 * time.Second)

		// todo: lock here

		if len(d.queue) <= 0 {
			// todo: unlock here
			// should skip if there's not items in the queue
			continue
		}

		// pop first round state from queue
		rs, new_queue := d.queue[0], d.queue[1:]
		d.queue = new_queue

		// todo: unlock here

		fmt.Printf("Detector: processing round state %#v\n", rs)
	}
}
