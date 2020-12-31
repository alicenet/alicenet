package lstate

import (
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/utils"
	lru "github.com/hashicorp/golang-lru"
)

type bHCache struct {
	cache *lru.Cache
}

func (bhc *bHCache) init() error {
	cache, err := lru.New(int(constants.EpochLength * 4))
	if err != nil {
		return err
	}
	bhc.cache = cache
	return nil
}

func (bhc *bHCache) add(bh *objs.BlockHeader) error {
	bHsh, err := bh.BClaims.BlockHash()
	if err != nil {
		return err
	}
	bhBytes, err := bh.MarshalBinary()
	if err != nil {
		return err
	}
	bhc.cache.Add(string(bHsh), string(bhBytes))
	return nil
}

func (bhc *bHCache) containsBlockHash(bHsh []byte) bool {
	return bhc.cache.Contains(string(bHsh))
}

func (bhc *bHCache) get(bHsh []byte) (*objs.BlockHeader, bool) {
	bhIf, ok := bhc.cache.Get(string(bHsh))
	if ok {
		bhString := bhIf.(string)
		bhBytes := []byte(bhString)
		bhCopy := utils.CopySlice(bhBytes)
		bh := &objs.BlockHeader{}
		err := bh.UnmarshalBinary(bhCopy)
		if err != nil {
			bhc.removeBlockHash(bHsh)
			return nil, false
		}
		return bh, true
	}
	return nil, false
}

func (bhc *bHCache) removeBlockHash(bHsh []byte) bool {
	return bhc.cache.Remove(string(bHsh))
}

func (bhc *bHCache) purge() {
	bhc.cache.Purge()
}
