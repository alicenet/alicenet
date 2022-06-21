package peering

import (
	"errors"

	"github.com/alicenet/alicenet/config"
	"github.com/alicenet/alicenet/interfaces"
	"github.com/alicenet/alicenet/transport"
)

type bootNodeList struct{}

func (bnl *bootNodeList) randomBootNode() (interfaces.NodeAddr, error) {
	lst := []interfaces.NodeAddr{}
	for _, bn := range config.Configuration.Transport.BootNodes() {
		addr, err := transport.NewNodeAddr(bn)
		if err != nil {
			return nil, err
		}
		lst = append(lst, addr)
	}
	if len(lst) == 0 {
		return nil, errors.New("no boot nodes exist")
	}
	idx, err := randomElement(len(lst))
	if err != nil {
		return nil, err
	}
	return lst[idx], nil
}
