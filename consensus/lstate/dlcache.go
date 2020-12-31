package lstate

import (
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"
)

type dLCache struct {
	sync.RWMutex
	defaultTO time.Duration
	cache     *lru.Cache
}

func (txc *dLCache) init(to time.Duration) error {
	cache, err := lru.New(1024)
	if err != nil {
		return err
	}
	txc.cache = cache
	txc.defaultTO = to
	return nil
}

func (txc *dLCache) add(txHsh []byte) error {
	txc.Lock()
	defer txc.Unlock()
	if txc.cache.Contains(string(txHsh)) {
		return nil
	}
	txc.cache.Add(string(txHsh), time.Now().Add(txc.defaultTO))
	return nil
}

func (txc *dLCache) expired(txHsh []byte) bool {
	deadline := txc.get(txHsh)
	now := time.Now()
	return now.After(deadline)
}

func (txc *dLCache) cancelOne(txHsh []byte) bool {
	txc.Lock()
	defer txc.Unlock()
	return txc.cache.Remove(string(txHsh))
}

func (txc *dLCache) cancelAll() {
	txc.Lock()
	defer txc.Unlock()
	txc.cache.Purge()
}

func (txc *dLCache) containsTxHsh(txHsh []byte) bool {
	txc.RLock()
	defer txc.RUnlock()
	return txc.cache.Contains(string(txHsh))
}

func (txc *dLCache) get(txHsh []byte) time.Time {
	txc.RLock()
	defer txc.RUnlock()
	deadlineIf, ok := txc.cache.Peek(string(txHsh))
	if ok {
		return deadlineIf.(time.Time)
	}
	return time.Now().Add(-1 * time.Second)
}
