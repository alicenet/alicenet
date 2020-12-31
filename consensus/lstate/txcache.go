package lstate

import (
	"sync"

	"github.com/MadBase/MadNet/consensus/appmock"
	"github.com/MadBase/MadNet/interfaces"
	"github.com/MadBase/MadNet/utils"
	lru "github.com/hashicorp/golang-lru"
)

type txCache struct {
	sync.RWMutex
	cache *lru.Cache
	app   appmock.Application
}

func (txc *txCache) init() error {
	cache, err := lru.New(4096)
	if err != nil {
		return err
	}
	txc.cache = cache
	return nil
}

func (txc *txCache) add(tx interfaces.Transaction) error {
	txc.Lock()
	defer txc.Unlock()
	txHash, err := tx.TxHash()
	if err != nil {
		return err
	}
	txb, err := tx.MarshalBinary()
	if err != nil {
		return err
	}
	txc.cache.Add(string(txHash), string(txb))
	return nil
}

func (txc *txCache) containsTxHsh(txHsh []byte) bool {
	txc.RLock()
	defer txc.RUnlock()
	return txc.cache.Contains(string(txHsh))
}

func (txc *txCache) containsTx(tx interfaces.Transaction) (bool, error) {
	txc.RLock()
	defer txc.RUnlock()
	txHsh, err := tx.TxHash()
	if err != nil {
		return false, err
	}
	return txc.cache.Contains(string(txHsh)), nil
}

func (txc *txCache) get(txHsh []byte) (interfaces.Transaction, bool) {
	txc.RLock()
	defer txc.RUnlock()
	txIf, ok := txc.cache.Get(string(txHsh))
	if ok {
		txb, ok := txIf.(string)
		if !ok {
			return nil, false
		}
		tx, err := txc.app.UnmarshalTx(utils.CopySlice([]byte(txb)))
		if err != nil {
			return nil, false
		}
		return tx, ok
	}
	return nil, false
}

func (txc *txCache) getMany(txHashes [][]byte) ([]interfaces.Transaction, [][]byte) {
	txc.RLock()
	defer txc.RUnlock()
	result := []interfaces.Transaction{}
	missing := [][]byte{}
	for i := 0; i < len(txHashes); i++ {
		txIf, ok := txc.cache.Get(string(txHashes[i]))
		if !ok {
			continue
		}
		txb, ok := txIf.(string)
		if !ok {
			missing = append(missing, utils.CopySlice(txHashes[i]))
			continue
		}
		tx, err := txc.app.UnmarshalTx(utils.CopySlice([]byte(txb)))
		if err != nil {
			missing = append(missing, utils.CopySlice(txHashes[i]))
			continue
		}
		result = append(result, tx)
	}
	return result, missing
}

func (txc *txCache) removeTx(tx interfaces.Transaction) (bool, error) {
	txc.Lock()
	defer txc.Unlock()
	txHsh, err := tx.TxHash()
	if err != nil {
		return false, err
	}
	return txc.cache.Remove(string(txHsh)), nil
}

func (txc *txCache) del(txHsh []byte) bool {
	txc.Lock()
	defer txc.Unlock()
	return txc.cache.Remove(string(txHsh))
}

func (txc *txCache) purge() {
	txc.Lock()
	defer txc.Unlock()
	txc.cache.Purge()
}
