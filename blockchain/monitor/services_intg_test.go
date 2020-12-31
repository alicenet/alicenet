package monitor_test

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/MadBase/MadNet/blockchain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ServicesSuite struct {
	suite.Suite
	commit func()
	eth    blockchain.Ethereum
}

func (s *ServicesSuite) SetupTest() {
	t := s.T()

	eth, commit, err := blockchain.NewEthereumSimulator(
		"../../assets/test/keys",
		"../../assets/test/passcodes.txt",
		3,
		5*time.Second,
		0,
		big.NewInt(9223372036854775807),
		"0x26D3D8Ab74D62C26f1ACc220dA1646411c9880Ac",
		"0x546F99F244b7B58B855330AE0E2BC1b30b41302F")

	assert.Nil(t, err, "Error creating Ethereum simulator")

	s.eth = eth
	s.commit = commit
}

func (s *ServicesSuite) TestRegistrationOpenEvent() {
	t := s.T()
	eth := s.eth
	c := eth.Contracts()
	assert.NotNil(t, c, "Need a *Contracts")

	height, err := s.eth.GetCurrentHeight(context.TODO())
	assert.Nil(t, err, "could not get height")
	assert.Equal(t, uint64(0), height, "Height should be 0")

	s.commit()

	height, err = s.eth.GetCurrentHeight(context.TODO())
	assert.Nil(t, err, "could not get height")
	assert.Equal(t, uint64(1), height, "Height should be 1")
}

func TestServicesSuite(t *testing.T) {
	suite.Run(t, new(ServicesSuite))
}
