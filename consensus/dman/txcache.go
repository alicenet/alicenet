package dman

import (
	"sync"

	"github.com/alicenet/alicenet/interfaces"
	"github.com/alicenet/alicenet/utils"
)

type txCache struct {
	sync.Mutex
	app    txMarshaller
	cache  map[string][]byte
	rcache map[string]uint32
}

func (txc *txCache) Init(app txMarshaller) error {
	txc.app = app
	txc.cache = make(map[string][]byte)
	txc.rcache = make(map[string]uint32)
	return nil
}

func (txc *txCache) Add(height uint32, tx interfaces.Transaction) error {
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
	txc.rcache[string(txHash)] = height
	txc.cache[string(txHash)] = utils.CopySlice(txb)
	return nil
}

func (txc *txCache) Contains(txHsh []byte) bool {
	txc.Lock()
	defer txc.Unlock()
	if _, ok := txc.getInternal(txHsh); ok {
		return true
	}
	return false
}

func (txc *txCache) Get(txHsh []byte) (interfaces.Transaction, bool) {
	txc.Lock()
	defer txc.Unlock()
	return txc.getInternal(txHsh)
}

func (txc *txCache) getInternal(txHsh []byte) (interfaces.Transaction, bool) {
	txb, ok := txc.cache[string(txHsh)]
	if ok {
		tx, err := txc.app.UnmarshalTx(utils.CopySlice(txb))
		if err != nil {
			txc.delInternal(txHsh)
			return nil, false
		}
		return tx, ok
	}
	return nil, false
}

func (txc *txCache) GetHeight(height uint32) ([]interfaces.Transaction, [][]byte) {
	txc.Lock()
	defer txc.Unlock()
	out := []interfaces.Transaction{}
	outhsh := [][]byte{}
	for hash, rh := range txc.rcache {
		hash, rh := hash, rh
		if rh == height {
			if txi, ok := txc.getInternal([]byte(hash)); ok {
				out = append(out, txi)
				outhsh = append(outhsh, []byte(hash))
			}
		}
	}
	return out, outhsh
}

func (txc *txCache) Del(txHsh []byte) {
	txc.Lock()
	defer txc.Unlock()
	txc.delInternal(txHsh)
}

func (txc *txCache) delInternal(txHsh []byte) {
	delete(txc.cache, string(txHsh))
	delete(txc.rcache, string(txHsh))
}

func (txc *txCache) DropBeforeHeight(dropHeight uint32) []string {
	out := []string{}
	if dropHeight-256 > dropHeight {
		return out
	}
	txc.Lock()
	defer txc.Unlock()
	for hash, height := range txc.rcache {
		hash, height := hash, height
		if height <= uint32(dropHeight) {
			txc.delInternal([]byte(hash))
			out = append(out, hash)
		}
	}
	return out
}
