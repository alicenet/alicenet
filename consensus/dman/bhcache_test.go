package dman

import (
	"reflect"
	"testing"

	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/crypto"
)

func goodBH() *objs.BlockHeader {
	bh := &objs.BlockHeader{
		BClaims: &objs.BClaims{
			ChainID:    1,
			Height:     1,
			TxCount:    0,
			PrevBlock:  make([]byte, 32),
			TxRoot:     crypto.Hasher([]byte{}),
			StateRoot:  make([]byte, 32),
			HeaderRoot: make([]byte, 32),
		},
		TxHshLst: [][]byte{},
		SigGroup: make([]byte, 192),
	}
	return bh
}

func badBH() *objs.BlockHeader {
	bh := &objs.BlockHeader{
		BClaims: &objs.BClaims{
			ChainID:    1,
			Height:     1,
			TxCount:    1,
			PrevBlock:  make([]byte, 32),
			TxRoot:     make([]byte, 32),
			StateRoot:  make([]byte, 32),
			HeaderRoot: make([]byte, 32),
		},
		TxHshLst: [][]byte{},
		SigGroup: nil,
	}
	return bh
}

func emptyCache() *bHCache {
	bhc := &bHCache{}
	bhc.Init()
	return bhc
}

func Test_bHCache_Init(t *testing.T) {
	tests := []struct {
		name    string
		bhc     *bHCache
		wantErr bool
	}{
		{"bh cache init", &bHCache{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.bhc.Init(); (err != nil) != tt.wantErr {
				t.Errorf("bHCache.Init() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_bHCache_Add(t *testing.T) {
	type args struct {
		bh *objs.BlockHeader
	}
	tests := []struct {
		name    string
		bhc     *bHCache
		args    args
		wantErr bool
	}{
		{"bh cache add good block", emptyCache(), args{bh: goodBH()}, false},
		{"bh cache add nil block", emptyCache(), args{bh: nil}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.bhc.Add(tt.args.bh); (err != nil) != tt.wantErr {
				t.Errorf("bHCache.Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_bHCache_Contains(t *testing.T) {
	type args struct {
		height uint32
	}

	newCache := emptyCache()
	newCache.Add(goodBH())

	tests := []struct {
		name string
		bhc  *bHCache
		args args
		want bool
	}{
		{"Does contain", newCache, args{height: 1}, true},
		{"Does not contain", newCache, args{height: 2}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.bhc.Contains(tt.args.height); got != tt.want {
				t.Errorf("bHCache.Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_bHCache_Get(t *testing.T) {
	type args struct {
		height uint32
	}

	newCache := emptyCache()
	newCache.Add(goodBH())

	badCache := emptyCache()
	badCache.Add(badBH())

	tests := []struct {
		name  string
		bhc   *bHCache
		args  args
		want  *objs.BlockHeader
		want1 bool
	}{
		{"Get existing", newCache, args{height: 1}, goodBH(), true},
		{"Get not existing", newCache, args{height: 2}, nil, false},
		{"Get bad block", badCache, args{height: 1}, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := tt.bhc.Get(tt.args.height)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("bHCache.Get() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("bHCache.Get() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_bHCache_Del(t *testing.T) {
	newCache := emptyCache()
	newCache.Add(goodBH())
	newCache.Del(1)
}

func Test_bHCache_DropBeforeHeight(t *testing.T) {
	newCache := emptyCache()
	newCache.Add(goodBH())
	newCache.DropBeforeHeight(1)
	newCache.DropBeforeHeight(257)
}
