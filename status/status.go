package status

import (
	"runtime"
	"sync"
	"time"

	"github.com/alicenet/alicenet/blockchain/monitor"
	"github.com/alicenet/alicenet/consensus/admin"
	"github.com/alicenet/alicenet/consensus/lstate"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/logging"
	"github.com/alicenet/alicenet/peering"
	"github.com/sirupsen/logrus"
)

// Logger is a status logging object. This object aggregates state summaries
// from services to print the status logs seen during normal operation.
type Logger struct {
	sync.Mutex
	closeChan chan struct{}
	closeOnce sync.Once
	wg        sync.WaitGroup
	log       *logrus.Logger
	ce        *lstate.Engine
	pm        *peering.PeerManager
	ad        *admin.Handlers
	mon       monitor.Monitor
}

// Init initalizes the object
func (sl *Logger) Init(ce *lstate.Engine, pm *peering.PeerManager, ad *admin.Handlers, mon monitor.Monitor) {
	sl.log = logging.GetLogger(constants.StatusLogger)
	sl.ce = ce
	sl.pm = pm
	sl.ad = ad
	sl.mon = mon
	sl.closeChan = make(chan struct{})
	sl.closeOnce = sync.Once{}
	sl.wg = sync.WaitGroup{}
}

// Close closes the object
func (sl *Logger) Close() {
	sl.closeOnce.Do(func() {
		sl.Lock()
		close(sl.closeChan)
		sl.Unlock()
		sl.wg.Wait()
	})
}

// Run starts a locking loop that will print status logs
func (sl *Logger) Run() {
	sl.Lock()
	select {
	case <-sl.closeChan:
		return
	default:
		sl.wg.Add(1)
		defer sl.wg.Done()
	}
	sl.Unlock()
	brOld := ""
	startTime := time.Now()
	oldMsg := ""
	oldMap := make(map[string]interface{})
	for {
		select {
		case <-sl.closeChan:
			return
		default:
		}
		_, err := sl.pm.Status(oldMap)
		if err != nil {
			continue
		}
		oldMap[constants.StatusGRCnt] = runtime.NumGoroutine()
		select {
		case <-sl.closeChan:
			return
		case msg := <-sl.mon.GetStatus():
			oldMsg = msg
			sl.log.WithFields(oldMap).Info(oldMsg)
		case <-time.After(time.Second):
			smap := make(map[string]interface{})
			if sl.ad.IsSynchronized() {
				_, err := sl.ce.Status(smap)
				if err != nil {
					continue
				}
				_, err = sl.pm.Status(smap)
				if err != nil {
					continue
				}
				br, ok := smap[constants.StatusBlkRnd].(string)
				if !ok {
					continue
				}
				if br != brOld {
					smap[constants.StatusBlkTime] = time.Since(startTime).Round(time.Millisecond * 10)
					smap[constants.StatusGRCnt] = runtime.NumGoroutine()
					startTime = time.Now()
					brOld = br
					oldMap = smap
					select {
					case msg := <-sl.mon.GetStatus():
						oldMsg = msg
						sl.log.WithFields(smap).Info(msg)
					default:
						sl.log.WithFields(smap).Info(oldMsg)
					}
				}
			}
		}
	}
}
